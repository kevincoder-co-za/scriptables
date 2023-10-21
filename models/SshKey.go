package models

import (
	"time"

	"gorm.io/gorm"
)

type SshKey struct {
	gorm.Model
	Name       string    `gorm:"column:name;type:varchar(100)"`
	ID         int64     `gorm:"column:id"`
	PrivateKey string    `gorm:"column:private_key"`
	PublicKey  string    `gorm:"column:public_key"`
	Passphrase string    `gorm:"column:passphrase;type:varchar(255)"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	TeamId     int64     `gorm:"column:team_id"`
}

func GetSshKeysList(db *gorm.DB, page int, perPage int, search string, teamId int64) []SshKey {
	offset := (page - 1) * perPage
	var keys []SshKey

	if search != "" {
		db.Limit(perPage).Offset(offset).Where(
			"name LIKE ? and team_id=?",
			"%"+search+"%",
			teamId,
		).Find(&keys)

	} else {
		db.Limit(perPage).Offset(offset).Where("team_id=?", teamId).Find(&keys)
	}

	return keys
}
