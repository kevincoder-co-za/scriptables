package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"gorm.io/gorm"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

type loggedEntity struct {
	Title      string
	Section    string
	Highlight  string
	InProgress bool
}

func isInProgress(status string) bool {
	return status == models.STATUS_QUEUED || status == models.STATUS_RUNNING
}

func describeLoggedEntity(db *gorm.DB, entity string, id int64, teamId int64) (loggedEntity, bool) {
	switch entity {
	case "application":
		var application models.Application
		db.Where("id = ? AND team_id = ?", id, teamId).Limit(1).Find(&application)
		catalogApplication, found := models.FindCatalogApplication(application.Slug)
		return loggedEntity{
			Title:      "Install logs for " + catalogApplication.Name,
			Section:    "Applications",
			Highlight:  "applications",
			InProgress: isInProgress(application.Status),
		}, found

	case "site":
		var site models.Site
		db.Where("id = ? AND team_id = ?", id, teamId).Limit(1).Find(&site)
		return loggedEntity{
			Title:      "Site logs for " + site.SiteName,
			Section:    "Sites",
			Highlight:  "sites",
			InProgress: isInProgress(site.Status),
		}, site.ID != 0

	case "cron":
		var cron models.Cron
		db.Unscoped().Where("id = ? AND team_id = ?", id, teamId).Limit(1).Find(&cron)
		return loggedEntity{
			Title:      "Cron logs for " + cron.CronName,
			Section:    "Crons",
			Highlight:  "crons",
			InProgress: isInProgress(cron.Status),
		}, cron.ID != 0

	case models.SECURITY_LOG_ENTITY:
		setting := models.GetSecuritySetting(db)
		return loggedEntity{
			Title:      "Security logs",
			Section:    "Security",
			Highlight:  "security",
			InProgress: isInProgress(setting.Status),
		}, setting.ID == id && setting.TeamId == teamId

	case models.FIREWALL_LOG_ENTITY:
		return loggedEntity{Title: "Firewall logs", Section: "Firewall", Highlight: "firewall"}, true
	}

	return loggedEntity{}, false
}

func (c *Controller) EntityLogs(gctx *gin.Context) {
	entity := gctx.Param("entity")
	entityId, _ := strconv.ParseInt(gctx.Param("id"), 10, 64)
	sessUser := c.GetSessionUser(gctx)
	db := c.GetDB(gctx)

	described, found := describeLoggedEntity(db, entity, entityId, sessUser.TeamId)
	if !found {
		c.FlashError(gctx, "Sorry, those logs could not be found.")
		gctx.Redirect(http.StatusFound, "/")
		return
	}

	page, err := strconv.Atoi(gctx.Query("page"))
	if err != nil {
		page = 1
	}

	perPage, err := strconv.Atoi(gctx.Query("perPage"))
	if err != nil {
		perPage = 20
	}

	logLevel := gctx.Query("log_level")
	logLevelQuery := ""
	if logLevel != "" {
		logLevelQuery = "&log_level=" + logLevel
	}

	c.Render("logs/list", gonja.Context{
		"title":         described.Title,
		"section":       described.Section,
		"highlight":     described.Highlight,
		"inProgress":    described.InProgress,
		"entity":        entity,
		"entityId":      entityId,
		"logs":          models.GetOperationLogs(db, page, perPage, entity, entityId, logLevel, sessUser.TeamId),
		"log_level":     logLevel,
		"logLevelQuery": logLevelQuery,
		"page":          page,
		"nextPage":      page + 1,
		"prevPage":      page - 1,
	}, gctx)
}

func (c *Controller) FullLog(gctx *gin.Context) {
	logId, _ := strconv.ParseInt(gctx.Param("id"), 10, 64)
	sessUser := c.GetSessionUser(gctx)

	var log models.OperationLog
	c.GetDB(gctx).Where("id = ? AND entity = ? AND team_id = ?", logId, gctx.Param("entity"), sessUser.TeamId).
		Limit(1).Find(&log)

	if log.ID == 0 {
		c.FlashError(gctx, "Sorry, log not found.")
		gctx.Redirect(http.StatusFound, "/")
		return
	}

	c.RenderWithoutLayout("logs/view_log", gonja.Context{
		"log": utils.Decrypt(log.Log),
	}, gctx)
}
