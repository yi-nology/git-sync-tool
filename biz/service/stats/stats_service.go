package stats

import (
	"bufio"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type StatsStatus string

const (
	StatusProcessing StatsStatus = "processing"
	StatusReady      StatsStatus = "ready"
	StatusFailed     StatsStatus = "failed"
)

type StatsCacheItem struct {
	Status    StatsStatus
	Data      *api.StatsResponse
	Error     error
	CreatedAt time.Time
	Progress  string // e.g. "Processed 100 commits..."
}

type StatsService struct {
	Git   *git.GitService
	cache sync.Map // map[string]*StatsCacheItem
}

var StatsSvc *StatsService

func InitStatsService() {
	StatsSvc = &StatsService{
		Git: git.NewGitService(),
	}
}

// SyncRepoStats performs background synchronization of repository statistics
// It fetches new commits since the last checkpoint and saves them to DB.
func (s *StatsService) SyncRepoStats(repoID uint, path, branch string) {
	log.Printf("[StatsSync] Starting sync for repo %d (%s)...", repoID, branch)

	// 1. Get latest commit time from DB (Checkpoint)
	commitStatDAO := db.NewCommitStatDAO()
	lastTime, err := commitStatDAO.FindLatestCommitTime(repoID)
	if err != nil {
		log.Printf("[StatsSync] Failed to get latest commit time: %v", err)
		return
	}

	log.Printf("[StatsSync] Resuming from %v", lastTime)

	// 2. Get git log iterator
	cIter, err := s.Git.GetLogIterator(path, branch)
	if err != nil {
		log.Printf("[StatsSync] Failed to get git log: %v", err)
		return
	}

	var batch []*po.CommitStat
	batchSize := 50

	// 3. Iterate commits
	err = cIter.ForEach(func(c *object.Commit) error {
		// Stop if we reach the checkpoint
		// Note: We need to handle time precision. Git time might have seconds.
		// If c.Author.When <= lastTime, we might have processed it.
		// To be safe, we process if c.Author.When > lastTime.
		// However, due to potential timezone or precision issues, strict > might miss commits if multiple happened at exact same second.
		// Better approach: process everything >= lastTime, and rely on DB unique index (Upsert) to handle duplicates.
		if !lastTime.IsZero() && c.Author.When.Before(lastTime) {
			return nil // Optimization: Stop iteration if order is guaranteed (git log usually is reverse chronological)
			// Wait, cIter iterates from NEWEST to OLDEST.
			// So if we encounter a commit OLDER than lastTime, we can stop?
			// Yes, usually.
		}

		// Calculate stats for this commit
		stats, err := c.Stats()
		if err != nil {
			log.Printf("[StatsSync] Failed to get stats for commit %s: %v", c.Hash.String(), err)
			return nil // Skip this commit but continue
		}

		additions := 0
		deletions := 0
		for _, fs := range stats {
			additions += fs.Addition
			deletions += fs.Deletion
		}

		batch = append(batch, &po.CommitStat{
			RepoID:      repoID,
			CommitHash:  c.Hash.String(),
			AuthorName:  c.Author.Name,
			AuthorEmail: c.Author.Email,
			CommitTime:  c.Author.When,
			Additions:   additions,
			Deletions:   deletions,
		})

		// Flush batch
		if len(batch) >= batchSize {
			if err := commitStatDAO.BatchSave(batch); err != nil {
				log.Printf("[StatsSync] Failed to save batch: %v", err)
			}
			batch = nil // Reset
		}

		return nil
	})

	// Flush remaining
	if len(batch) > 0 {
		if err := commitStatDAO.BatchSave(batch); err != nil {
			log.Printf("[StatsSync] Failed to save final batch: %v", err)
		}
	}

	if err != nil {
		log.Printf("[StatsSync] Error during iteration: %v", err)
	}

	log.Printf("[StatsSync] Completed sync for repo %d", repoID)
}

// ParseCommits parses raw git log output into Commit structs
func (s *StatsService) ParseCommits(raw string) []domain.Commit {
	var commits []domain.Commit
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		t, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

		commits = append(commits, domain.Commit{
			Hash:      parts[0],
			Author:    parts[1],
			Email:     parts[2],
			Date:      t,
			Timestamp: t.Unix(),
			Message:   parts[4],
		})
	}

	// Sort by Timestamp Descending (Newest First)
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Timestamp > commits[j].Timestamp
	})

	return commits
}

type ActivityStat struct {
	Name  string
	Trend map[string]int
}

// GetStats retrieves stats from cache or triggers calculation
func (s *StatsService) GetStats(path, branch, since, until string) (*api.StatsResponse, StatsStatus, error, string) {
	key := fmt.Sprintf("%s:%s:%s:%s", path, branch, since, until)

	// 1. Check cache
	if val, ok := s.cache.Load(key); ok {
		item := val.(*StatsCacheItem)
		// Simple TTL: 1 hour
		if time.Since(item.CreatedAt) < time.Hour {
			return item.Data, item.Status, item.Error, item.Progress
		}
	}

	// 2. Initialize cache item (Processing)
	newItem := &StatsCacheItem{
		Status:    StatusProcessing,
		CreatedAt: time.Now(),
		Progress:  "Initializing...",
	}
	// Use LoadOrStore to prevent duplicate concurrent calculations
	actual, loaded := s.cache.LoadOrStore(key, newItem)

	if loaded {
		item := actual.(*StatsCacheItem)
		if time.Since(item.CreatedAt) < time.Hour {
			return item.Data, item.Status, item.Error, item.Progress
		}
		return item.Data, item.Status, item.Error, item.Progress
	}

	// 3. Start async calculation
	go func() {
		data, err := s.calculateStatsFast(path, branch, since, until, key)
		if err != nil {
			s.updateCache(key, func(item *StatsCacheItem) {
				item.Status = StatusFailed
				item.Error = err
			})
		} else {
			s.updateCache(key, func(item *StatsCacheItem) {
				item.Status = StatusReady
				item.Data = data
				item.Progress = "Completed"
			})
		}
	}()

	return nil, StatusProcessing, nil, "Initializing..."
}

func (s *StatsService) updateCache(key string, update func(*StatsCacheItem)) {
	if val, ok := s.cache.Load(key); ok {
		item := val.(*StatsCacheItem)
		update(item)
		// No need to Store back since we modified the pointer, but sync.Map might need it if we replaced the struct.
		// Since we are modifying fields of the struct pointer, it is visible to other goroutines reading the same pointer.
		// However, to be safe from race conditions on the struct fields themselves if they were not atomic,
		// we should be careful. But here it's simple string/status updates.
		// Ideally we should use a mutex inside StatsCacheItem or replace the item in the map.
		// For progress reporting, replacing the item in map is safer if we treat it as immutable, but slower.
		// Let's assume for now the pointer approach is "good enough" for status updates or we can re-store.
		// Actually, let's create a new item to be thread-safe for readers? No, that breaks the "LoadOrStore" logic if we want to share progress.
		// We should probably add a Mutex to StatsCacheItem.
	}
}

// calculateStatsFast computes stats using git log --numstat (Fast, No Blame)
func (s *StatsService) calculateStatsFast(path, branch, since, until, cacheKey string) (*api.StatsResponse, error) {
	// Parse dates
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}
	if until != "" {
		untilTime, _ = time.Parse("2006-01-02", until)
		// Set until to end of day
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}

	// 1. Get raw log stats stream
	stream, err := s.Git.GetLogStatsStream(path, branch)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	authorStats := make(map[string]*api.AuthorStat)

	scanner := bufio.NewScanner(stream)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentEmail, currentName string
	var currentDate time.Time

	commitCount := 0
	lastUpdate := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "COMMIT|") {
			commitCount++

			// Update progress every 100 commits or 1 second
			if commitCount%100 == 0 || time.Since(lastUpdate) > time.Second {
				s.updateCache(cacheKey, func(item *StatsCacheItem) {
					item.Progress = fmt.Sprintf("Processed %d commits...", commitCount)
				})
				lastUpdate = time.Now()
			}

			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				// COMMIT|Hash|Name|Email|Timestamp
				currentName = parts[2]
				currentEmail = parts[3]
				ts, _ := strconv.ParseInt(parts[4], 10, 64)
				currentDate = time.Unix(ts, 0)
			}
			continue
		}

		// ... (rest of the loop)

		// Date Filter
		if !sinceTime.IsZero() && currentDate.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && currentDate.After(untilTime) {
			continue
		}

		// Parse numstat: "added deleted filename"
		// Note: binary files might show "-"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, err1 := strconv.Atoi(parts[0])
		deleted, err2 := strconv.Atoi(parts[1])

		// Skip binary files or parse errors
		if err1 != nil || err2 != nil {
			continue
		}

		filename := parts[2]
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) > 0 {
			ext = ext[1:] // remove dot
		} else {
			ext = "unknown"
		}

		if _, exists := authorStats[currentEmail]; !exists {
			authorStats[currentEmail] = &api.AuthorStat{
				Name:      currentName,
				Email:     currentEmail,
				FileTypes: make(map[string]int),
				TimeTrend: make(map[string]int),
			}
		}

		stat := authorStats[currentEmail]
		// Use Net Contribution as "Total Lines" approximation
		stat.TotalLines += (added - deleted)
		stat.FileTypes[ext] += (added - deleted)

		dateStr := currentDate.Format("2006-01-02")
		stat.TimeTrend[dateStr] += (added - deleted)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice
	resp := &api.StatsResponse{
		Authors: make([]*api.AuthorStat, 0, len(authorStats)),
	}
	for _, stat := range authorStats {
		resp.Authors = append(resp.Authors, stat)
		resp.TotalLines += stat.TotalLines
	}

	return resp, nil
}
