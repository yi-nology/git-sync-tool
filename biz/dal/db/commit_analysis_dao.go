package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type CommitAnalysisDAO struct {
	db *gorm.DB
}

func NewCommitAnalysisDAO() *CommitAnalysisDAO {
	return &CommitAnalysisDAO{
		db: DB,
	}
}

func (dao *CommitAnalysisDAO) CreateCommitAnalysis(analysis *po.CommitAnalysis) error {
	return dao.db.Create(analysis).Error
}

func (dao *CommitAnalysisDAO) GetCommitAnalysisByHash(repoKey, commitHash string) (*po.CommitAnalysis, error) {
	var analysis po.CommitAnalysis
	err := dao.db.Where("repo_key = ? AND commit_hash = ?", repoKey, commitHash).First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (dao *CommitAnalysisDAO) GetCommitAnalysesByRepo(repoKey string, limit int) ([]po.CommitAnalysis, error) {
	var analyses []po.CommitAnalysis
	query := dao.db.Where("repo_key = ?", repoKey).
		Order("commit_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&analyses).Error
	return analyses, err
}

func (dao *CommitAnalysisDAO) GetCommitAnalysesByTimeRange(repoKey string, startTime, endTime time.Time) ([]po.CommitAnalysis, error) {
	var analyses []po.CommitAnalysis
	err := dao.db.Where("repo_key = ? AND commit_time BETWEEN ? AND ?", repoKey, startTime, endTime).
		Order("commit_time ASC").
		Find(&analyses).Error
	return analyses, err
}

func (dao *CommitAnalysisDAO) GetCommitCountByTimeRange(repoKey string, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := dao.db.Model(&po.CommitAnalysis{}).
		Where("repo_key = ? AND commit_time BETWEEN ? AND ?", repoKey, startTime, endTime).
		Count(&count).Error
	return count, err
}

func (dao *CommitAnalysisDAO) GetTotalChangesByTimeRange(repoKey string, startTime, endTime time.Time) (int, error) {
	type Result struct {
		Total int
	}
	var result Result
	err := dao.db.Model(&po.CommitAnalysis{}).
		Select("SUM(total_changes) as total").
		Where("repo_key = ? AND commit_time BETWEEN ? AND ?", repoKey, startTime, endTime).
		Scan(&result).Error
	return result.Total, err
}

func (dao *CommitAnalysisDAO) CreateCommitPattern(pattern *po.CommitPattern) error {
	return dao.db.Create(pattern).Error
}

func (dao *CommitAnalysisDAO) GetCommitPatternsByRepo(repoKey string, patternType string, limit int) ([]po.CommitPattern, error) {
	var patterns []po.CommitPattern
	query := dao.db.Where("repo_key = ?", repoKey)

	if patternType != "" {
		query = query.Where("pattern_type = ?", patternType)
	}

	query = query.Order("end_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&patterns).Error
	return patterns, err
}

func (dao *CommitAnalysisDAO) GetLatestCommitPattern(repoKey, patternType, patternValue string) (*po.CommitPattern, error) {
	var pattern po.CommitPattern
	err := dao.db.Where("repo_key = ? AND pattern_type = ? AND pattern_value = ?", repoKey, patternType, patternValue).
		Order("end_date DESC").
		First(&pattern).Error
	if err != nil {
		return nil, err
	}
	return &pattern, nil
}

func (dao *CommitAnalysisDAO) UpdateCommitPattern(pattern *po.CommitPattern) error {
	return dao.db.Save(pattern).Error
}

func (dao *CommitAnalysisDAO) CreateSyncRecommendation(recommendation *po.SyncRecommendation) error {
	return dao.db.Create(recommendation).Error
}

func (dao *CommitAnalysisDAO) GetSyncRecommendation(repoKey, taskKey string) (*po.SyncRecommendation, error) {
	var recommendation po.SyncRecommendation
	err := dao.db.Where("repo_key = ? AND task_key = ?", repoKey, taskKey).
		Order("last_analysis DESC").
		First(&recommendation).Error
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func (dao *CommitAnalysisDAO) GetSyncRecommendationsByRepo(repoKey string, limit int) ([]po.SyncRecommendation, error) {
	var recommendations []po.SyncRecommendation
	query := dao.db.Where("repo_key = ?", repoKey).
		Order("last_analysis DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&recommendations).Error
	return recommendations, err
}

func (dao *CommitAnalysisDAO) UpdateSyncRecommendation(recommendation *po.SyncRecommendation) error {
	return dao.db.Save(recommendation).Error
}

func (dao *CommitAnalysisDAO) DeleteOldCommitAnalyses(repoKey string, keepDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	return dao.db.Where("repo_key = ? AND commit_time < ?", repoKey, cutoffTime).Delete(&po.CommitAnalysis{}).Error
}

func (dao *CommitAnalysisDAO) DeleteOldCommitPatterns(repoKey string, keepDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	return dao.db.Where("repo_key = ? AND end_date < ?", repoKey, cutoffTime).Delete(&po.CommitPattern{}).Error
}
