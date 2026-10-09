package models

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/utils"
)

type SiteQueue struct {
	ID int64 `gorm:"column:id"`

	SiteID    int64
	Status    string
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Site struct {
	gorm.Model
	ID                     int64     `gorm:"column:id"`
	Domain                 string    `gorm:"column:domain;type:varchar(255)"`
	ScriptableName         string    `gorm:"column:scriptable_name;type:varchar(100)"`
	DeployScriptables      string    `gorm:"column:deploy_scriptables;type:varchar(100)"`
	SiteName               string    `gorm:"column:site_name"`
	Webroot                string    `gorm:"column:webroot;type:varchar(100)"`
	PhpVersion             string    `gorm:"column:php_version;type:varchar(50)"`
	LetsEncryptCertificate int       `gorm:"column:lets_encrypt_certificate;type:tinyint(3)"`
	MysqlPassword          string    `gorm:"column:mysql_password;type:varchar(155)"`
	GitURL                 string    `gorm:"column:git_url;type:varchar(255)"`
	Branch                 string    `gorm:"column:branch;type:varchar(100)"`
	Environment            string    `gorm:"column:environment;type:varchar(50)"`
	Status                 string    `gorm:"column:status;type:varchar(50)"`
	DeployToken            string    `gorm:"column:deploy_token;type:varchar(255)"`
	CreatedAt              time.Time `gorm:"column:created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at"`
	TeamId                 int64     `gorm:"column:team_id"`
}

func GetSitesList(db *gorm.DB, page, perPage int, search, status string, teamId int64) []Site {
	offset := (page - 1) * perPage
	var sites []Site

	query := db.Where("team_id = ?", teamId)

	if search != "" {
		query = query.Where("domain LIKE ?", search+"%")
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	query.Limit(perPage).Offset(offset).Find(&sites)

	return sites
}

func GetSiteByTokenAndId(db *gorm.DB, token string, id int64) *Site {
	var site *Site
	db.Where("deploy_token=? and id=?", token, id).First(&site)
	return site
}

func GetSiteByIdNoTeam(db *gorm.DB, id int64) *Site {
	var site *Site
	db.Where("id=?", id).First(&site)
	return site
}

func GetQueuedSites(db *gorm.DB) []Site {
	var sites []Site
	db.Where("status = ?", STATUS_QUEUED).Find(&sites)
	return sites
}

func GetSiteIdsQueuedForDeploy(db *gorm.DB) []int64 {
	var siteIds []int64
	db.Table("site_queues").Distinct("site_id").Where("status = ?", STATUS_QUEUED).Scan(&siteIds)
	return siteIds
}

func MarkQueuedDeploysComplete(db *gorm.DB, siteId int64) {
	db.Table("site_queues").Where("site_id = ? AND status = ?", siteId, STATUS_QUEUED).
		Updates(map[string]interface{}{"status": STATUS_COMPLETE, "updated_at": time.Now()})
}

func SetSiteStatus(db *gorm.DB, id int64, status string) {
	db.Model(&Site{}).Where("id = ?", id).Update("status", status)
}

func PhpFpmPort(phpVersion string) string {
	return "90" + strings.ReplaceAll(phpVersion, ".", "")
}

func (site *Site) ReplaceScriptableVariables(db *gorm.DB, script string) string {
	username := utils.Slugify(site.SiteName)
	script = strings.ReplaceAll(script, "#SITE_NAME#", site.SiteName)
	script = strings.ReplaceAll(script, "#SITE_SLUG#", username)
	script = strings.ReplaceAll(script, "#MYSQL_PASSWORD#", utils.Decrypt(site.MysqlPassword))
	script = strings.ReplaceAll(script, "#PHP_VERSION#", site.PhpVersion)
	script = strings.ReplaceAll(script, "#FPM_PORT#", PhpFpmPort(site.PhpVersion))
	script = strings.ReplaceAll(script, "#BRANCH#", site.Branch)
	script = strings.ReplaceAll(script, "#GIT_URL#", site.GitURL)
	script = strings.ReplaceAll(script, "#ENVIRONMENT#", strings.ReplaceAll(site.Environment, ".env", ""))
	script = strings.ReplaceAll(script, "#WEBROOT#", site.Webroot)
	script = strings.ReplaceAll(script, "#DOMAIN#", site.Domain)
	script = strings.ReplaceAll(script, "#KEY_PATH#", GetSiteDeployKeyPath(site.ID, site.SiteName, username))
	script = strings.ReplaceAll(script, "#USER_DIRECTORY#", "/home/"+username)

	var user User
	db.Table("users").Where("team_id = ?", site.TeamId).Order("id asc").Limit(1).Scan(&user)
	script = strings.ReplaceAll(script, "#NOTIFY_EMAIL#", user.Email)

	return script
}

func GetSiteDeployKeyPath(id int64, siteName string, username string) string {
	return "/home/" + username + "/.ssh/" + strconv.Itoa(int(id)) + siteName
}
