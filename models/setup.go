package models

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetAppDB() (*gorm.DB, error) {
	var err error
	var db *gorm.DB

	for i := 0; i <= 3; i++ {
		db, err := gorm.Open(mysql.Open(os.Getenv("MYSQL_DSN")), &gorm.Config{})
		if err == nil {
			return db, err
		}

		time.Sleep(5 * time.Second)
	}

	return db, err
}

func SetDBConnection(c *gin.Context) {
	var err error = nil
	var sqlDB *sql.DB
	var gdb *gorm.DB

	DB, exists := c.Get("db")

	if !exists {
		err = errors.New("connection to mysql failed")
	} else {

		gdb = DB.(*gorm.DB)
		sqlDB, err = gdb.DB()
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(100)
				sqlDB.SetConnMaxLifetime(time.Minute * 30)
			}
		}
	}

	if err != nil {
		gdb, err = GetAppDB()
		if err == nil {
			c.Set("db", gdb)
			return
		}
	}
}
