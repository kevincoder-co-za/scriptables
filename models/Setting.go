package models

import (
	"net/url"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/utils"
)

const SETTINGS_LOG_ENTITY = "settings"

type Setting struct {
	gorm.Model
	ID               int64     `gorm:"column:id"`
	Domain           string    `gorm:"column:domain;type:varchar(255)"`
	DomainStatus     string    `gorm:"column:domain_status;type:varchar(50)"`
	CertificateEmail string    `gorm:"column:certificate_email;type:varchar(255)"`
	SmtpHost         string    `gorm:"column:smtp_host;type:varchar(255)"`
	SmtpPort         string    `gorm:"column:smtp_port;type:varchar(10)"`
	SmtpUsername     string    `gorm:"column:smtp_username;type:varchar(255)"`
	SmtpPassword     string    `gorm:"column:smtp_password;type:varchar(255)"`
	SmtpFromEmail    string    `gorm:"column:smtp_from_email;type:varchar(255)"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
	TeamId           int64     `gorm:"column:team_id"`
}

func GetSetting(db *gorm.DB) Setting {
	var setting Setting
	db.Order("id asc").Limit(1).Find(&setting)
	return setting
}

func SaveSetting(db *gorm.DB, setting *Setting) {
	setting.UpdatedAt = time.Now()
	if setting.ID == 0 {
		setting.CreatedAt = time.Now()
	}

	db.Save(setting)
}

func (setting Setting) HasSmtpSettings() bool {
	return setting.SmtpHost != ""
}

func (setting Setting) DecryptedSmtpPassword() string {
	if setting.SmtpPassword == "" {
		return ""
	}

	return utils.Decrypt(setting.SmtpPassword)
}

func envMailConfig() utils.MailConfig {
	return utils.MailConfig{
		Host:      os.Getenv("SMTP_HOST"),
		Port:      os.Getenv("SMTP_PORT"),
		Username:  os.Getenv("SMTP_USERNAME"),
		Password:  os.Getenv("SMTP_PASSWORD"),
		FromEmail: os.Getenv("SMTP_FROM_EMAIL"),
	}
}

func (setting Setting) MailConfig() utils.MailConfig {
	if !setting.HasSmtpSettings() {
		return envMailConfig()
	}

	return utils.MailConfig{
		Host:      setting.SmtpHost,
		Port:      setting.SmtpPort,
		Username:  setting.SmtpUsername,
		Password:  setting.DecryptedSmtpPassword(),
		FromEmail: setting.SmtpFromEmail,
	}
}

func GetMailConfig(db *gorm.DB) utils.MailConfig {
	setting := GetSetting(db)
	config := setting.MailConfig()
	config.BaseUrl = setting.PublicUrl()
	return config
}

func ProvisionedDomain() string {
	parsed, err := url.Parse(os.Getenv("SCRIPTABLE_URL"))
	if err != nil || parsed.Scheme != "https" {
		return ""
	}

	return parsed.Hostname()
}

func IsDomainProvisioned() bool {
	return ProvisionedDomain() != ""
}

func (setting Setting) PublicUrl() string {
	if setting.Domain != "" && setting.DomainStatus == STATUS_COMPLETE && !IsDomainProvisioned() {
		return "https://" + setting.Domain
	}

	return strings.TrimRight(os.Getenv("SCRIPTABLE_URL"), "/")
}

func GetPublicUrl(db *gorm.DB) string {
	return GetSetting(db).PublicUrl()
}

func GetQueuedDomainSettings(db *gorm.DB) []Setting {
	var settings []Setting
	db.Where("domain_status = ?", STATUS_QUEUED).Find(&settings)
	return settings
}

func SetDomainStatus(db *gorm.DB, id int64, status string) {
	db.Model(&Setting{}).Where("id = ?", id).Update("domain_status", status)
}

func (setting Setting) ReplaceScriptableVariables(script string) string {
	script = strings.ReplaceAll(script, "#DOMAIN#", setting.Domain)
	script = strings.ReplaceAll(script, "#NOTIFY_EMAIL#", setting.CertificateEmail)
	script = strings.ReplaceAll(script, "#PANEL_PORT#", utils.PanelPort())
	return script
}
