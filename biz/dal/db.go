package dal

import (
	"log"
	"os"

	"github.com/yi-nology/git-manage-service/biz/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "git_sync.db"
	}
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Migrate the schema
	err = DB.AutoMigrate(&model.Repo{}, &model.SyncTask{}, &model.SyncRun{}, &model.AuditLog{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
}
