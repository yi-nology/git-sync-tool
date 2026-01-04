package stats

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"

	gogit "github.com/go-git/go-git/v5"
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
	return commits
}

type ActivityStat struct {
	Name  string
	Trend map[string]int
}

// GetStats retrieves stats from cache or triggers calculation
func (s *StatsService) GetStats(path, branch, since, until string) (*api.StatsResponse, StatsStatus, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", path, branch, since, until)

	// 1. Check cache
	if val, ok := s.cache.Load(key); ok {
		item := val.(*StatsCacheItem)
		// Simple TTL: 1 hour
		if time.Since(item.CreatedAt) < time.Hour {
			return item.Data, item.Status, item.Error
		}
	}

	// 2. Initialize cache item (Processing)
	newItem := &StatsCacheItem{
		Status:    StatusProcessing,
		CreatedAt: time.Now(),
	}
	// Use LoadOrStore to prevent duplicate concurrent calculations
	actual, loaded := s.cache.LoadOrStore(key, newItem)

	if loaded {
		item := actual.(*StatsCacheItem)
		if time.Since(item.CreatedAt) < time.Hour {
			return item.Data, item.Status, item.Error
		}
		return item.Data, item.Status, item.Error
	}

	// 3. Start async calculation
	go func() {
		data, err := s.calculateStatsInternal(path, branch, since, until)
		if err != nil {
			s.cache.Store(key, &StatsCacheItem{
				Status:    StatusFailed,
				Error:     err,
				CreatedAt: time.Now(),
			})
		} else {
			s.cache.Store(key, &StatsCacheItem{
				Status:    StatusReady,
				Data:      data,
				CreatedAt: time.Now(),
			})
		}
	}()

	return nil, StatusProcessing, nil
}

// calculateStatsInternal computes effective line counts per author
func (s *StatsService) calculateStatsInternal(path, branch, since, until string) (*api.StatsResponse, error) {
	files, err := s.Git.GetRepoFiles(path, branch)
	if err != nil {
		return nil, err
	}

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

	authorStats := make(map[string]*api.AuthorStat)
	var mu sync.Mutex

	// Worker pool to process files
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, file := range files {
		if strings.TrimSpace(file) == "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			rawBlame, err := s.Git.BlameFile(path, branch, f)
			if err != nil {
				return
			}

			lines := s.parseBlame(rawBlame, f)

			mu.Lock()
			defer mu.Unlock()

			for _, line := range lines {
				// Date Filter
				if !sinceTime.IsZero() && line.Date.Before(sinceTime) {
					continue
				}
				if !untilTime.IsZero() && line.Date.After(untilTime) {
					continue
				}

				if _, exists := authorStats[line.Email]; !exists {
					authorStats[line.Email] = &api.AuthorStat{
						Name:      line.Author,
						Email:     line.Email,
						FileTypes: make(map[string]int),
						TimeTrend: make(map[string]int),
					}
				}

				stat := authorStats[line.Email]
				stat.TotalLines++
				stat.FileTypes[line.Extension]++
			}
		}(file)
	}

	wg.Wait()

	// Calculate Activity Trend using git log with DB caching
	activityTrends, err := s.getContributionStats(path, branch, since, until)
	if err == nil {
		for email, data := range activityTrends {
			if _, exists := authorStats[email]; !exists {
				authorStats[email] = &api.AuthorStat{
					Name:       data.Name,
					Email:      email,
					FileTypes:  make(map[string]int),
					TimeTrend:  data.Trend,
					TotalLines: 0,
				}
			} else {
				authorStats[email].TimeTrend = data.Trend
			}
		}
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

func (s *StatsService) parseBlame(result *gogit.BlameResult, filename string) []domain.LineStat {
	var stats []domain.LineStat

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 0 {
		ext = ext[1:] // remove dot
	} else {
		ext = "unknown"
	}

	for _, line := range result.Lines {
		if s.isEffectiveLine(line.Text, ext) {
			stats = append(stats, domain.LineStat{
				Author:    line.Author,
				Email:     line.Author, // go-git Line.Author is typically the email
				Date:      line.Date,
				Extension: ext,
			})
		}
	}

	return stats
}

func (s *StatsService) isEffectiveLine(content, ext string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "--") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") {
		return false
	}
	return true
}

// getContributionStats uses DB cache to speed up stats calculation
func (s *StatsService) getContributionStats(path, branch, since, until string) (map[string]*ActivityStat, error) {
	// 1. Get Repo ID
	repo, err := db.NewRepoDAO().FindByPath(path)
	if err != nil {
		return nil, err
	}

	// 2. Parse dates
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}
	if until != "" {
		untilTime, _ = time.Parse("2006-01-02", until)
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}

	// 3. Trigger Async Sync to ensure we are up to date (Best Effort)
	// We don't wait for it to finish, but it helps populate DB for next time if missing
	go s.SyncRepoStats(repo.ID, path, branch)

	// 4. Get all commits from git log
	cIter, err := s.Git.GetLogIterator(path, branch)
	if err != nil {
		return nil, err
	}

	var allHashes []string
	commitMap := make(map[string]*object.Commit)

	err = cIter.ForEach(func(c *object.Commit) error {
		if !untilTime.IsZero() && c.Author.When.After(untilTime) {
			return nil
		}
		if !sinceTime.IsZero() && c.Author.When.Before(sinceTime) {
			return nil
		}
		// Skip merges
		if len(c.ParentHashes) > 1 {
			return nil
		}

		allHashes = append(allHashes, c.Hash.String())
		commitMap[c.Hash.String()] = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 5. Batch get from DB
	commitStatDAO := db.NewCommitStatDAO()
	cachedStats, err := commitStatDAO.GetByRepoAndHashes(repo.ID, allHashes)
	if err != nil {
		cachedStats = make(map[string]*po.CommitStat)
	}

	// 6. Identify missing commits & Calculate
	var missingStats []*po.CommitStat
	results := make(map[string]*ActivityStat)

	for _, hash := range allHashes {
		var additions int
		var authorName, authorEmail string
		var commitDate time.Time

		if stat, ok := cachedStats[hash]; ok {
			// Hit cache
			additions = stat.Additions
			authorName = stat.AuthorName
			authorEmail = stat.AuthorEmail
			commitDate = stat.CommitTime
		} else {
			// Miss cache, calculate
			c := commitMap[hash]
			fileStats, err := c.Stats()
			if err != nil {
				continue
			}

			additions = 0
			deletions := 0
			for _, fs := range fileStats {
				additions += fs.Addition
				deletions += fs.Deletion
			}

			// Add to missing list for DB insertion
			missingStats = append(missingStats, &po.CommitStat{
				RepoID:      repo.ID,
				CommitHash:  hash,
				AuthorName:  c.Author.Name,
				AuthorEmail: c.Author.Email,
				CommitTime:  c.Author.When,
				Additions:   additions,
				Deletions:   deletions,
			})

			authorName = c.Author.Name
			authorEmail = c.Author.Email
			commitDate = c.Author.When
		}

		// Aggregate for result
		dateStr := commitDate.Format("2006-01-02")
		if _, ok := results[authorEmail]; !ok {
			results[authorEmail] = &ActivityStat{
				Name:  authorName,
				Trend: make(map[string]int),
			}
		}
		results[authorEmail].Trend[dateStr] += additions
	}

	// 7. Async save missing stats to DB
	if len(missingStats) > 0 {
		go func() {
			_ = commitStatDAO.BatchSave(missingStats)
		}()
	}

	return results, nil
}
