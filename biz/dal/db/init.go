package db

import (
	"fmt"
	"log"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	sqlite "github.com/glebarez/sqlite"
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

	// Check if tables exist to skip initialization if requested
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
		migrator.HasTable(&po.BackupRecord{}) {
		log.Println("Database tables exist, skipping schema migration.")
		return
	}

	// Migrate the schema
	err = DB.AutoMigrate(&po.Repo{}, &po.SyncTask{}, &po.SyncRun{}, &po.AuditLog{}, &po.SystemConfig{}, &po.CommitStat{}, &po.NotificationChannel{}, &po.NotificationEventTemplate{}, &po.SSHKey{}, &po.BackupRecord{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
}
