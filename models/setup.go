package models

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func databasePath() string {
	if path := os.Getenv("SQLITE_PATH"); path != "" {
		return path
	}

	return "scriptables.db"
}

func OpenDatabase() (*gorm.DB, error) {
	dsn := databasePath() + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&Team{},
		&User{},
		&FailedLogins{},
		&Application{},
		&Site{},
		&SiteQueue{},
		&Cron{},
		&SecuritySetting{},
		&OperationLog{},
	)

	return db, err
}

func FailInterruptedJobs(db *gorm.DB) {
	for _, table := range []string{"applications", "sites", "crons", "security_settings"} {
		db.Table(table).Where("status = ?", STATUS_RUNNING).Update("status", STATUS_FAILED)
	}
}
