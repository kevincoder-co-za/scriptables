package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexscriptables.com/scriptables/models"
)

var linuxUsernamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

type cronForm struct {
	User           string
	Task           string
	CronExpression string
	CronName       string
}

func readCronForm(gctx *gin.Context) cronForm {
	form := cronForm{
		User:           strings.TrimSpace(gctx.PostForm("user")),
		Task:           strings.TrimSpace(gctx.PostForm("task")),
		CronExpression: strings.TrimSpace(gctx.PostForm("cron_expression")),
		CronName:       gctx.PostForm("cron_name"),
	}

	if form.User == "" {
		form.User = "root"
	}

	return form
}

func (form cronForm) validate() []string {
	errors := []string{}

	if !models.IsValidCronExpression(form.CronExpression) {
		errors = append(errors, "Sorry, the cron expression entered is invalid.")
	}

	if form.Task == "" || strings.ContainsAny(form.Task, "\r\n") {
		errors = append(errors, "Command to execute must be a single line and cannot be empty.")
	}

	if !linuxUsernamePattern.MatchString(form.User) {
		errors = append(errors, "Cron user must be a valid Linux username.")
	}

	return errors
}

func (form cronForm) templateContext(action string, status string) gonja.Context {
	return gonja.Context{
		"title":           "Setup cron",
		"user":            form.User,
		"task":            form.Task,
		"cron_expression": form.CronExpression,
		"cron_name":       form.CronName,
		"status":          status,
		"action":          action,
		"highlight":       "crons",
	}
}

func (c *Controller) findSessionTeamCron(gctx *gin.Context, id string) models.Cron {
	cronId, _ := strconv.ParseInt(id, 10, 64)
	sessUser := c.GetSessionUser(gctx)

	var cron models.Cron
	if cronId != 0 {
		c.GetDB(gctx).Where("id = ? AND team_id = ?", cronId, sessUser.TeamId).Limit(1).Find(&cron)
	}

	return cron
}

func (c *Controller) CreateCron(gctx *gin.Context) {
	form := cronForm{User: "root", CronExpression: "* * * * *"}
	c.Render("crons/form", form.templateContext("/cron/save", "pending"), gctx)
}

func (c *Controller) SaveCron(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	form := readCronForm(gctx)

	if errors := form.validate(); len(errors) > 0 {
		ctx := form.templateContext("/cron/save", models.STATUS_QUEUED)
		ctx["errors"] = errors
		c.Render("crons/form", ctx, gctx)
		return
	}

	cron := models.Cron{
		User:           form.User,
		Task:           form.Task,
		Status:         models.STATUS_QUEUED,
		CronExpression: form.CronExpression,
		CronName:       form.CronName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TeamID:         sessUser.TeamId,
	}
	c.GetDB(gctx).Create(&cron)

	c.FlashSuccess(gctx, "Successfully queued cron for deployment. Please check the logs for progress.")
	gctx.Redirect(http.StatusFound, "/crons")
}

func (c *Controller) Crons(gctx *gin.Context) {
	page, err := strconv.Atoi(gctx.Query("page"))
	sessUser := c.GetSessionUser(gctx)

	if err != nil {
		page = 1
	}

	perPage, err := strconv.Atoi(gctx.Query("perPage"))
	if err != nil {
		perPage = 20
	}

	search := gctx.Query("search")
	crons := models.GetCrons(c.GetDB(gctx), page, perPage, search, sessUser.TeamId)

	c.Render("crons/list", gonja.Context{
		"title":     "Cron Jobs",
		"crons":     crons,
		"nextPage":  page + 1,
		"prevPage":  page - 1,
		"search":    search,
		"numCrons":  len(crons),
		"highlight": "crons",
	}, gctx)
}

func (c *Controller) EditCron(gctx *gin.Context) {
	cron := c.findSessionTeamCron(gctx, gctx.Param("id"))

	if cron.ID == 0 {
		c.FlashError(gctx, "Cron with ID "+gctx.Param("id")+" does not exist.")
		gctx.Redirect(http.StatusFound, "/crons")
		return
	}

	form := cronForm{User: cron.User, Task: cron.Task, CronExpression: cron.CronExpression, CronName: cron.CronName}
	c.Render("crons/form", form.templateContext(fmt.Sprintf("/cron/update/%d", cron.ID), cron.Status), gctx)
}

func (c *Controller) UpdateCron(gctx *gin.Context) {
	cron := c.findSessionTeamCron(gctx, gctx.Param("id"))

	if cron.ID == 0 {
		c.FlashError(gctx, "Cron with ID "+gctx.Param("id")+" does not exist.")
		gctx.Redirect(http.StatusFound, "/crons")
		return
	}

	form := readCronForm(gctx)

	if errors := form.validate(); len(errors) > 0 {
		ctx := form.templateContext(fmt.Sprintf("/cron/update/%d", cron.ID), models.STATUS_QUEUED)
		ctx["errors"] = errors
		c.Render("crons/form", ctx, gctx)
		return
	}

	cron.User = form.User
	cron.Task = form.Task
	cron.Status = models.STATUS_QUEUED
	cron.UpdatedAt = time.Now()
	cron.CronName = form.CronName
	cron.CronExpression = form.CronExpression
	c.GetDB(gctx).Save(&cron)

	c.FlashSuccess(gctx, "Successfully updated cron.")
	gctx.Redirect(http.StatusFound, "/crons")
}

func (c *Controller) DisableCron(gctx *gin.Context) {
	cronId, _ := strconv.ParseInt(gctx.PostForm("id"), 10, 64)
	sessUser := c.GetSessionUser(gctx)

	if cronId == 0 {
		c.FlashError(gctx, "Invalid Cron ID - please try again.")
		gctx.Redirect(http.StatusFound, "/crons")
		return
	}

	c.GetDB(gctx).Exec("UPDATE crons SET deleted_at = ?, status = ? WHERE id = ? and team_id = ?",
		time.Now(), models.STATUS_QUEUED, cronId, sessUser.TeamId)
	c.FlashSuccess(gctx, "Successfully queued cron for deletion.")
	gctx.Redirect(http.StatusFound, "/crons")
}

func (c *Controller) RetryCronBuild(gctx *gin.Context) {
	cronId, _ := strconv.ParseInt(gctx.PostForm("retryBuildId"), 10, 64)
	sessUser := c.GetSessionUser(gctx)

	if cronId == 0 {
		c.FlashError(gctx, "Sorry, failed to queue cron deploy. Please try again.")
		gctx.Redirect(http.StatusFound, "/crons")
		return
	}

	c.GetDB(gctx).Exec("UPDATE crons set status = ? where id = ? and team_id = ?",
		models.STATUS_QUEUED, cronId, sessUser.TeamId)
	c.FlashSuccess(gctx, "Successfully queued cron for deployment.")
	gctx.Redirect(http.StatusFound, "/crons")
}
