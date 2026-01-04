package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm/clause"
)

type CommitStatDAO struct{}

func NewCommitStatDAO() *CommitStatDAO {
	return &CommitStatDAO{}
}

// FindLatestCommitTime returns the latest commit time for a repo
func (dao *CommitStatDAO) FindLatestCommitTime(repoID uint) (time.Time, error) {
	var stat po.CommitStat
	err := DB.Where("repo_id = ?", repoID).Order("commit_time desc").First(&stat).Error
	if err != nil {
		return time.Time{}, nil // Return zero time if not found (start from beginning)
	}
	return stat.CommitTime, nil
}

// BatchSave inserts or updates commit stats
func (dao *CommitStatDAO) BatchSave(stats []*po.CommitStat) error {
	if len(stats) == 0 {
		return nil
	}
	// Upsert on conflict
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "repo_id"}, {Name: "commit_hash"}},
		DoUpdates: clause.AssignmentColumns([]string{"additions", "deletions", "author_name", "author_email", "commit_time"}),
	}).CreateInBatches(stats, 100).Error
}

// GetByRepoAndHashes finds existing stats for a repo and list of hashes
func (dao *CommitStatDAO) GetByRepoAndHashes(repoID uint, hashes []string) (map[string]*po.CommitStat, error) {
	if len(hashes) == 0 {
		return nil, nil
	}

	result := make(map[string]*po.CommitStat)
	chunkSize := 500

	for i := 0; i < len(hashes); i += chunkSize {
		end := i + chunkSize
		if end > len(hashes) {
			end = len(hashes)
		}
		
		var chunk []po.CommitStat
		err := DB.Where("repo_id = ? AND commit_hash IN ?", repoID, hashes[i:end]).Find(&chunk).Error
		if err != nil {
			return nil, err
		}
		
		for j := range chunk {
			result[chunk[j].CommitHash] = &chunk[j]
		}
	}

	return result, nil
}
