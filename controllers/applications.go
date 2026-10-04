package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexcorp.tech/scriptable/models"
)

func hasInstallInProgress(applications []models.ApplicationOverview) bool {
	for _, application := range applications {
		if application.Status == models.STATUS_QUEUED || application.Status == models.STATUS_RUNNING {
			return true
		}
	}

	return false
}

func (c *Controller) Applications(gctx *gin.Context) {
	applications := models.GetApplicationsOverview(c.GetDB(gctx))

	c.Render("applications/list", gonja.Context{
		"title":             "Applications",
		"highlight":         "applications",
		"applications":      applications,
		"installInProgress": hasInstallInProgress(applications),
		"STATUS_AVAILABLE":  models.STATUS_AVAILABLE,
	}, gctx)
}

func (c *Controller) InstallApplication(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)

	catalogApplication, found := models.FindCatalogApplication(gctx.PostForm("slug"))
	if !found {
		c.FlashError(gctx, "Sorry, that application is not available to install.")
		gctx.Redirect(http.StatusFound, "/applications")
		return
	}

	application := models.QueueApplicationInstall(c.GetDB(gctx), catalogApplication.Slug, sessUser.TeamId)

	c.FlashSuccess(gctx, "Successfully queued "+catalogApplication.Name+" for installation.")
	gctx.Redirect(http.StatusFound, fmt.Sprintf("/logs/application/%d", application.ID))
}
