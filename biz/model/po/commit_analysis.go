package po

import (
	"time"
)

type CommitAnalysis struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	RepoKey         string    `json:"repoKey" gorm:"index;not null"`
	CommitHash      string    `json:"commitHash" gorm:"index;not null"`
	Author          string    `json:"author"`
	AuthorEmail     string    `json:"authorEmail"`
	Committer       string    `json:"committer"`
	CommitterEmail  string    `json:"committerEmail"`
	Message         string    `json:"message" gorm:"type:text"`
	CommitTime      time.Time `json:"commitTime" gorm:"index;not null"`
	AnalysisTime    time.Time `json:"analysisTime" gorm:"index;not null"`
	AddedLines      int       `json:"addedLines"`
	DeletedLines    int       `json:"deletedLines"`
	TotalChanges    int       `json:"totalChanges"`
	FileChanges     string    `json:"fileChanges" gorm:"type:text"`     // JSON 格式存储
	LanguageChanges string    `json:"languageChanges" gorm:"type:text"` // JSON 格式存储
	CommitType      string    `json:"commitType" gorm:"index"`          // feat, fix, docs, style, refactor, test, chore
	IsMerge         bool      `json:"isMerge"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CommitPattern struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	RepoKey      string    `json:"repoKey" gorm:"index;not null"`
	PatternType  string    `json:"patternType" gorm:"index;not null"`  // daily, weekly, monthly, language
	PatternValue string    `json:"patternValue" gorm:"index;not null"` // 具体的值，如 "Monday", "Go"
	CommitCount  int       `json:"commitCount"`
	ChangeCount  int       `json:"changeCount"`
	StartDate    time.Time `json:"startDate" gorm:"index;not null"`
	EndDate      time.Time `json:"endDate" gorm:"index;not null"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type SyncRecommendation struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	RepoKey        string    `json:"repoKey" gorm:"index;not null"`
	TaskKey        string    `json:"taskKey" gorm:"index"`
	Recommendation string    `json:"recommendation" gorm:"type:text"`
	SyncFrequency  string    `json:"syncFrequency"` // 建议的同步频率
	Confidence     float64   `json:"confidence"`    // 推荐的置信度
	LastAnalysis   time.Time `json:"lastAnalysis" gorm:"index;not null"`
	IsApplied      bool      `json:"isApplied" gorm:"default:false"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
