package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexscriptables.com/scriptables/models"
)

const minimumRootPasswordLength = 8
const rootPasswordForbiddenCharacters = "'\"`\\ \t\r\n"

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

func validateRootPassword(password string, confirm string) []string {
	errors := []string{}

	if len(password) < minimumRootPasswordLength {
		errors = append(errors, "The root password must be at least 8 characters.")
	}

	if strings.ContainsAny(password, rootPasswordForbiddenCharacters) {
		errors = append(errors, "The root password cannot contain spaces, quotes or backslashes.")
	}

	if password != confirm {
		errors = append(errors, "Password and confirm password do not match.")
	}

	return errors
}

func (c *Controller) renderInstallForm(gctx *gin.Context, application models.CatalogApplication, errors []string) {
	vars := gonja.Context{
		"title":       "Install " + application.Name,
		"highlight":   "applications",
		"application": application,
	}

	if len(errors) > 0 {
		vars["errors"] = errors
	}

	c.Render("applications/install", vars, gctx)
}

func (c *Controller) findInstallableApplication(gctx *gin.Context, slug string) (models.CatalogApplication, bool) {
	catalogApplication, found := models.FindCatalogApplication(slug)
	if !found {
		c.FlashError(gctx, "Sorry, that application is not available to install.")
		gctx.Redirect(http.StatusFound, "/applications")
	}

	return catalogApplication, found
}

func (c *Controller) InstallApplicationForm(gctx *gin.Context) {
	catalogApplication, found := c.findInstallableApplication(gctx, gctx.Param("slug"))
	if !found {
		return
	}

	if !catalogApplication.AsksForRootPassword() {
		gctx.Redirect(http.StatusFound, "/applications")
		return
	}

	c.renderInstallForm(gctx, catalogApplication, nil)
}

func (c *Controller) InstallApplication(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)

	catalogApplication, found := c.findInstallableApplication(gctx, gctx.PostForm("slug"))
	if !found {
		return
	}

	rootPassword := ""
	if catalogApplication.AsksForRootPassword() {
		rootPassword = gctx.PostForm("root_password")
		if errors := validateRootPassword(rootPassword, gctx.PostForm("root_password_confirm")); len(errors) > 0 {
			c.renderInstallForm(gctx, catalogApplication, errors)
			return
		}
	}

	application := models.QueueApplicationInstall(c.GetDB(gctx), catalogApplication.Slug, sessUser.TeamId, rootPassword)

	c.FlashSuccess(gctx, "Successfully queued "+catalogApplication.Name+" for installation.")
	gctx.Redirect(http.StatusFound, fmt.Sprintf("/logs/application/%d", application.ID))
}
