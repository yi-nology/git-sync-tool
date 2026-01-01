package dal

import (
	"fmt"
	"log"

	"github.com/yi-nology/git-manage-service/biz/config"
	"github.com/yi-nology/git-manage-service/biz/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	var dialector gorm.Dialector

	dbConfig := config.GlobalConfig.Database

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

	// Migrate the schema
	err = DB.AutoMigrate(&model.Repo{}, &model.SyncTask{}, &model.SyncRun{}, &model.AuditLog{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
}
