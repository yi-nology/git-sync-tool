package service

import (
	"bufio"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model"
)

type StatsService struct {
	Git *GitService
}

var StatsSvc *StatsService

func InitStatsService() {
	StatsSvc = &StatsService{
		Git: NewGitService(),
	}
}

// ParseCommits parses raw git log output into Commit structs
func (s *StatsService) ParseCommits(raw string) []model.Commit {
	var commits []model.Commit
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

		commits = append(commits, model.Commit{
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

// CalculateStats computes effective line counts per author
func (s *StatsService) CalculateStats(path, branch, since, until string) (*model.StatsResponse, error) {
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

	authorStats := make(map[string]*model.AuthorStat)
	var mu sync.Mutex

	// Worker pool to process files
	// Limiting concurrency to avoid overwhelming system
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, file := range files {
		if strings.TrimSpace(file) == "" {
			continue
		}

		// Skip binary files or unlikely source code files based on extension?
		// For now, let's process everything but skip obvious binaries if blame fails or takes too long.
		// Actually git blame works on text.

		wg.Add(1)
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			rawBlame, err := s.Git.BlameFile(path, branch, f)
			if err != nil {
				// Log error or ignore
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
					authorStats[line.Email] = &model.AuthorStat{
						Name:      line.Author,
						Email:     line.Email,
						FileTypes: make(map[string]int),
						TimeTrend: make(map[string]int),
					}
				}

				stat := authorStats[line.Email]
				stat.TotalLines++
				stat.FileTypes[line.Extension]++

				dateKey := line.Date.Format("2006-01-02")
				stat.TimeTrend[dateKey]++
			}
		}(file)
	}

	wg.Wait()

	// Convert map to slice
	resp := &model.StatsResponse{
		Authors: make([]*model.AuthorStat, 0, len(authorStats)),
	}
	for _, stat := range authorStats {
		resp.Authors = append(resp.Authors, stat)
		resp.TotalLines += stat.TotalLines
	}

	return resp, nil
}

func (s *StatsService) parseBlame(raw, filename string) []model.LineStat {
	var stats []model.LineStat
	scanner := bufio.NewScanner(strings.NewReader(raw))

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 0 {
		ext = ext[1:] // remove dot
	} else {
		ext = "unknown"
	}

	// Blame --line-porcelain format:
	// <hash> <orig_line> <final_line> <num_lines>
	// author <name>
	// author-mail <email>
	// author-time <timestamp>
	// ...
	// filename <filename>
	// \t<content>

	var currentAuthor, currentEmail string
	var currentDate time.Time

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "author ") {
			currentAuthor = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			currentEmail = strings.Trim(strings.TrimPrefix(line, "author-mail "), "<>")
		} else if strings.HasPrefix(line, "author-time ") {
			ts, _ := strconv.ParseInt(strings.TrimPrefix(line, "author-time "), 10, 64)
			currentDate = time.Unix(ts, 0)
		} else if strings.HasPrefix(line, "\t") {
			// This is the content line
			content := strings.TrimPrefix(line, "\t")
			if s.isEffectiveLine(content, ext) {
				stats = append(stats, model.LineStat{
					Author:    currentAuthor,
					Email:     currentEmail,
					Date:      currentDate,
					Extension: ext,
				})
			}
		}
	}
	return stats
}

func (s *StatsService) isEffectiveLine(content, ext string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	// Basic comment filtering
	// This is not perfect but covers common cases
	if strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "--") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") {
		return false
	}

	return true
}
