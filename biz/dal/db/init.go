package db

import (
	"fmt"
	"log"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	var dialector gorm.Dialector

	dbConfig := configs.GlobalConfig.Database

	switch dbConfig.Type {
	case "mysql":
		dsn := dbConfig.DSN
		if dsn == "" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)
		}
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := dbConfig.DSN
		if dsn == "" {
			dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
				dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port)
		}
		dialector = postgres.Open(dsn)
	case "sqlite":
		fallthrough
	default:
		dbPath := dbConfig.Path
		if dbPath == "" {
			dbPath = "git_sync.db"
		}
		dialector = sqlite.Open(dbPath)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	migrator := DB.Migrator()
	if migrator.HasTable(&po.Repo{}) &&
		migrator.HasTable(&po.SyncTask{}) &&
		migrator.HasTable(&po.SyncRun{}) &&
		migrator.HasTable(&po.AuditLog{}) &&
		migrator.HasTable(&po.SystemConfig{}) &&
		migrator.HasTable(&po.CommitStat{}) &&
		migrator.HasTable(&po.NotificationChannel{}) &&
		migrator.HasTable(&po.NotificationEventTemplate{}) &&
		migrator.HasTable(&po.SSHKey{}) &&
		migrator.HasTable(&po.BackupRecord{}) &&
		migrator.HasTable(&po.Credential{}) &&
		migrator.HasTable(&po.LintRule{}) {
		log.Println("Database tables exist, skipping schema migration.")
		return
	}

	err = DB.AutoMigrate(&po.Repo{}, &po.SyncTask{}, &po.SyncRun{}, &po.AuditLog{}, &po.SystemConfig{}, &po.CommitStat{}, &po.NotificationChannel{}, &po.NotificationEventTemplate{}, &po.SSHKey{}, &po.BackupRecord{}, &po.Credential{}, &po.LintRule{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
}

func InitLintRules() {
	dao := NewLintRuleDAO()

	count, err := dao.Count()
	if err != nil {
		log.Printf("Warning: failed to count lint rules: %v", err)
		return
	}

	if count > 0 {
		return
	}

	rules := getDefaultLintRules()
	now := time.Now()
	for i := range rules {
		rules[i].CreatedAt = now
		rules[i].UpdatedAt = now
	}

	if err := dao.BatchCreate(rules); err != nil {
		log.Printf("Warning: failed to insert default rules: %v", err)
	}
}

func getDefaultLintRules() []po.LintRule {
	return []po.LintRule{
		{
			ID:          "spec-header-required",
			Name:        "Required Header Fields",
			Description: "Spec file must contain Name, Version, and Release fields",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_name",
			Enabled:     true,
			Priority:    1,
		},
		{
			ID:          "spec-version-required",
			Name:        "Required Version Field",
			Description: "Spec file must contain Version field",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_version",
			Enabled:     true,
			Priority:    2,
		},
		{
			ID:          "spec-release-required",
			Name:        "Required Release Field",
			Description: "Spec file must contain Release field",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_release",
			Enabled:     true,
			Priority:    3,
		},
		{
			ID:          "spec-summary-required",
			Name:        "Required Summary Field",
			Description: "Spec file must contain Summary field",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_summary",
			Enabled:     true,
			Priority:    4,
		},
		{
			ID:          "spec-license-required",
			Name:        "Required License Field",
			Description: "Spec file must contain License field",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_license",
			Enabled:     true,
			Priority:    5,
		},
		{
			ID:          "spec-url-recommended",
			Name:        "Recommended URL Field",
			Description: "Spec file should contain URL field for project homepage",
			Category:    "required",
			Severity:    "warning",
			Pattern:     "required_url",
			Enabled:     true,
			Priority:    6,
		},
		{
			ID:          "spec-description-required",
			Name:        "Required Description Section",
			Description: "Spec file must contain %description section",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_description",
			Enabled:     true,
			Priority:    7,
		},
		{
			ID:          "spec-prep-required",
			Name:        "Required Prep Section",
			Description: "Spec file should contain %prep section for unpacking sources",
			Category:    "required",
			Severity:    "warning",
			Pattern:     "required_prep",
			Enabled:     true,
			Priority:    8,
		},
		{
			ID:          "spec-build-required",
			Name:        "Required Build Section",
			Description: "Spec file should contain %build section for compilation",
			Category:    "required",
			Severity:    "warning",
			Pattern:     "required_build",
			Enabled:     true,
			Priority:    9,
		},
		{
			ID:          "spec-install-required",
			Name:        "Required Install Section",
			Description: "Spec file should contain %install section for installation",
			Category:    "required",
			Severity:    "warning",
			Pattern:     "required_install",
			Enabled:     true,
			Priority:    10,
		},
		{
			ID:          "spec-files-required",
			Name:        "Required Files Section",
			Description: "Spec file must contain %files section listing packaged files",
			Category:    "required",
			Severity:    "error",
			Pattern:     "required_files",
			Enabled:     true,
			Priority:    11,
		},
		{
			ID:          "spec-empty-sections",
			Name:        "Empty Sections",
			Description: "Spec sections should not be empty",
			Category:    "style",
			Severity:    "warning",
			Pattern:     "empty_sections",
			Enabled:     true,
			Priority:    22,
		},
		{
			ID:          "spec-buildroot-usage",
			Name:        "BuildRoot Usage",
			Description: "BuildRoot should use %{_tmppath} macro",
			Category:    "best_practice",
			Severity:    "warning",
			Pattern:     "buildroot_usage",
			Enabled:     true,
			Priority:    10,
		},
		{
			ID:          "spec-macro-consistency",
			Name:        "Macro Consistency",
			Description: "Use consistent macro style throughout the spec file",
			Category:    "best_practice",
			Severity:    "info",
			Pattern:     "macro_consistency",
			Enabled:     true,
			Priority:    11,
		},
		{
			ID:          "spec-changelog-format",
			Name:        "Changelog Format",
			Description: "Changelog entries should follow RPM format (start with *)",
			Category:    "style",
			Severity:    "warning",
			Pattern:     "changelog_format",
			Enabled:     true,
			Priority:    20,
		},
		{
			ID:          "spec-no-tabs",
			Name:        "No Tabs",
			Description: "Use spaces instead of tabs for consistency",
			Category:    "style",
			Severity:    "info",
			Pattern:     "no_tabs",
			Enabled:     true,
			Priority:    21,
		},
	}
}
