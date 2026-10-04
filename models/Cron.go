package models

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
	"gorm.io/gorm"
)

type Cron struct {
	gorm.Model
	ID             int64     `gorm:"column:id"`
	User           string    `gorm:"column:user"`
	Task           string    `gorm:"column:task;type:varchar(255)"`
	CronExpression string    `gorm:"column:cron_expression;type:varchar(100)"`
	CronName       string    `gorm:"column:cron_name"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	TeamID         int64     `gorm:"column:team_id"`
}

func IsValidCronExpression(input string) bool {
	_, err := cron.ParseStandard(input)
	return err == nil
}

func GetCrons(db *gorm.DB, page, perPage int, search string, teamId int64) []Cron {
	offset := (page - 1) * perPage
	var crons []Cron

	query := db.Where("team_id = ?", teamId)

	if search != "" {
		query = query.Where("cron_name LIKE ?", search+"%")
	}

	query.Limit(perPage).Offset(offset).Find(&crons)
	return crons
}

func GetQueuedCronsIncludingDisabled(db *gorm.DB) []Cron {
	var crons []Cron
	db.Unscoped().Where("status = ?", STATUS_QUEUED).Find(&crons)
	return crons
}

func BuildCronFile(db *gorm.DB) string {
	var crons []Cron
	db.Find(&crons)

	cronFile := ""
	for _, cron := range crons {
		cronFile += fmt.Sprintf("%s %s %s\n", cron.CronExpression, cron.User, cron.Task)
	}

	return cronFile
}

func SetCronStatus(db *gorm.DB, id int64, status string) {
	db.Unscoped().Model(&Cron{}).Where("id = ?", id).Update("status", status)
}
