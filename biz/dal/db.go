package dal

import (
	"github.com/yi-nology/git-sync-tool/biz/model"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("git_sync.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Migrate the schema
	err = DB.AutoMigrate(&model.Repo{}, &model.SyncTask{}, &model.SyncRun{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
}
