package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/noirbizarre/gonja"
	"plexcorp.tech/scriptable/console"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

var domainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
var gitUrlPattern = regexp.MustCompile(`^[\w.@:/~-]+$`)
var gitBranchPattern = regexp.MustCompile(`^[\w./-]+$`)
var webrootPattern = regexp.MustCompile(`^[\w./-]+$`)
var phpVersionPattern = regexp.MustCompile(`^\d\.\d$`)
var environmentPattern = regexp.MustCompile(`^[\w.-]+$`)

func findMissingSitePrerequisites() []string {
	missing := []string{}

	if !models.IsPackageInstalled("nginx") {
		missing = append(missing, "Nginx")
	}

	if !models.IsPackageInstalled("mysql-server") && !models.IsPackageInstalled("mariadb-server") {
		missing = append(missing, "MySQL or MariaDB")
	}

	return missing
}

func (c *Controller) CreateSite(gctx *gin.Context) {
	if missing := findMissingSitePrerequisites(); len(missing) > 0 {
		c.Render("general/warning", gonja.Context{
			"title":      "Missing applications",
			"highlight":  "sites",
			"warningMsg": "Please install " + strings.Join(missing, " and ") + " from the <a href=\"/applications\">applications</a> page before setting up a site.",
		}, gctx)

		return
	}

	password := utils.GenPassword()
	c.Render("sites/create", gonja.Context{
		"title":                   "Setup website",
		"domain":                  "",
		"webroot":                 "public",
		"php_version":             "",
		"letsencrypt_certificate": 0,
		"git_url":                 "",
		"scriptables":             "laravel",
		"mysql_password":          password,
		"mysql_password_confirm":  password,
		"environment":             "prod",
		"branch":                  "master",
		"highlight":               "sites",
	}, gctx)
}

func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	for _, unwanted := range []string{"https://", "http://", "://", "/"} {
		domain = strings.ReplaceAll(domain, unwanted, "")
	}

	return domain
}

func (c *Controller) SaveSite(gctx *gin.Context) {
	domain := normalizeDomain(gctx.PostForm("domain"))
	webroot := gctx.PostForm("webroot")
	giturl := strings.TrimSpace(gctx.PostForm("git_url"))
	phpVersion := gctx.PostForm("php_version")
	scriptables := gctx.PostForm("scriptables")
	mysqlPassword := gctx.PostForm("mysql_password")
	mysqlPasswordConfirm := gctx.PostForm("mysql_password_confirm")
	environment := gctx.PostForm("environment")
	branch := gctx.PostForm("branch")

	letsEncryptCertificate := 0
	if gctx.PostForm("letsencrypt_certificate") == "on" {
		letsEncryptCertificate = 1
	}

	siteName := strings.ReplaceAll(domain, ".", "")

	if scriptables == "" {
		scriptables = "laravel"
	}

	if environment == "" {
		environment = "prod"
	}

	if branch == "" {
		branch = "master"
	}

	ctx := gonja.Context{
		"title":                   "Setup a website",
		"domain":                  domain,
		"webroot":                 webroot,
		"php_version":             phpVersion,
		"letsencrypt_certificate": letsEncryptCertificate,
		"git_url":                 giturl,
		"site_name":               siteName,
		"mysql_password":          mysqlPassword,
		"mysql_password_confirm":  mysqlPasswordConfirm,
		"branch":                  branch,
		"scriptables":             scriptables,
		"environment":             environment,
		"highlight":               "sites",
	}

	errors := []string{}

	if !domainPattern.MatchString(domain) || utils.Slugify(siteName) == "" {
		errors = append(errors, "Domain seems invalid. Please check the domain uses this format: domain.ext or www.domain.ext or subdomain.domain.ext")
	}

	if mysqlPassword != mysqlPasswordConfirm {
		errors = append(errors, "Mysql password and confirm password not the same.")
	}

	if strings.ContainsAny(mysqlPassword, "'\\\r\n") {
		errors = append(errors, "Mysql password cannot contain quotes, backslashes or line breaks.")
	}

	if strings.Contains(giturl, "https://") {
		errors = append(errors, "Please use only the SSH GIT URL. e.g.: git@github.com:username/app.git")
	} else if !gitUrlPattern.MatchString(giturl) {
		errors = append(errors, "Please enter a valid GIT URL.")
	}

	if !gitBranchPattern.MatchString(branch) {
		errors = append(errors, "Please enter a valid GIT branch.")
	}

	if !webrootPattern.MatchString(webroot) || strings.Contains(webroot, "..") {
		errors = append(errors, "Please enter the path to your websites web root folder, relative to the project root.")
	}

	if !phpVersionPattern.MatchString(phpVersion) {
		errors = append(errors, "Please select a version of PHP to configure with this app.")
	}

	if !environmentPattern.MatchString(environment) {
		errors = append(errors, "Please enter a valid config file name e.g. prod.env")
	}

	if len(utils.FindScriptables(scriptables)) == 0 || strings.ContainsAny(scriptables, "./") {
		errors = append(errors, "Sorry, that site type is not supported.")
	}

	if letsEncryptCertificate == 1 && !models.IsPackageInstalled("certbot") {
		errors = append(errors, "Please install Certbot from the applications page before requesting a Let's Encrypt certificate.")
	}

	var found int64
	c.GetDB(gctx).Model(&models.Site{}).Where("domain = ? OR site_name = ?", domain, siteName).Count(&found)

	if found > 0 {
		errors = append(errors, "Sorry, domain already in use. You can have multiple subdomains but only one root domain.")
	}

	if len(errors) > 0 {
		ctx["errors"] = errors
		c.Render("sites/create", ctx, gctx)
		return
	}

	sessUser := c.GetSessionUser(gctx)
	site := models.Site{
		Domain:                 domain,
		SiteName:               siteName,
		PhpVersion:             phpVersion,
		Webroot:                webroot,
		LetsEncryptCertificate: letsEncryptCertificate,
		Status:                 models.STATUS_CONNECTING,
		ScriptableName:         scriptables,
		DeployScriptables:      scriptables + "_deploy",
		GitURL:                 giturl,
		MysqlPassword:          utils.Encrypt(mysqlPassword),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Environment:            environment,
		Branch:                 branch,
		DeployToken:            strings.ReplaceAll(uuid.New().String(), "-", ""),
		TeamId:                 sessUser.TeamId,
	}
	c.GetDB(gctx).Create(&site)

	gctx.Redirect(http.StatusFound, "/site/deployKey/"+strconv.Itoa(int(site.ID)))
}

func (c *Controller) Sites(gctx *gin.Context) {
	view := gctx.Query("view")
	status := gctx.Query("status")
	sessUser := c.GetSessionUser(gctx)
	if view == "" {
		view = "grid"
	}

	if status == "" {
		status = "all"
	}

	page, err := strconv.Atoi(gctx.Query("page"))
	if err != nil {
		page = 1
	}

	perPage, err := strconv.Atoi(gctx.Query("perPage"))
	if err != nil {
		perPage = 20
	}

	search := gctx.Query("search")
	sites := models.GetSitesList(c.GetDB(gctx), page, perPage, search, status, sessUser.TeamId)

	c.Render("sites/list", gonja.Context{
		"title":     "Sites",
		"sites":     sites,
		"nextPage":  page + 1,
		"prevPage":  page - 1,
		"search":    search,
		"view":      view,
		"status":    status,
		"numSites":  len(sites),
		"highlight": "sites",
	}, gctx)
}

func (c *Controller) findSessionTeamSite(gctx *gin.Context, id string) *models.Site {
	siteId, _ := strconv.ParseInt(id, 10, 64)
	sessUser := c.GetSessionUser(gctx)

	site := &models.Site{}
	if siteId != 0 {
		c.GetDB(gctx).Where("id = ? AND team_id = ?", siteId, sessUser.TeamId).Limit(1).Find(site)
	}

	return site
}

func (c *Controller) CreateSiteDeployKey(gctx *gin.Context) {
	site := c.findSessionTeamSite(gctx, gctx.Param("id"))

	if site.ID == 0 {
		c.FlashError(gctx, "Ooops, sorry seems like you do not have permission to access this site. Please try again.")
		gctx.Redirect(http.StatusFound, "/sites")
		return
	}

	c.Render("sites/deploykey", gonja.Context{
		"title":      "GIT setup for: " + site.SiteName,
		"siteId":     site.ID,
		"token":      site.DeployToken,
		"highlight":  "sites",
		"successMsg": "Successfully saved site: " + site.SiteName + ". Now generating deploy key..., once done please copy and add to your repos deploy keys.",
	}, gctx)
}

func (c *Controller) GenerateDeployKey(gctx *gin.Context) {
	site := c.findSessionTeamSite(gctx, gctx.PostForm("siteId"))
	db := c.GetDB(gctx)

	fail := func(reason string) {
		c.RenderWithoutLayout("sites/_deploykey", gonja.Context{
			"siteId":   site.ID,
			"errorMsg": reason,
		}, gctx)
	}

	if site.ID == 0 {
		fail("You do not have permission to access this site.")
		return
	}

	script, err := utils.ReadSharedScriptable("deploy_keysetup")
	if err != nil {
		fail("Could not find the deploy key setup script. Check that scriptables/__shared/deploy_keysetup.sh exists.")
		return
	}

	keyPath := models.GetSiteDeployKeyPath(site.ID, site.SiteName, utils.Slugify(site.SiteName))

	output, err := utils.RunScript(site.ReplaceScriptableVariables(db, script))
	if err != nil {
		models.LogError(db, site.ID, "site", err.Error()+". Command output: "+output,
			"Failed to create deploy key: "+site.SiteName, site.TeamId)
		fail("Failed to create the SSH key " + keyPath + ".")
		return
	}

	publicKey, err := utils.RunCommandAsRoot("cat", keyPath+".pub")
	publicKey = strings.TrimSpace(publicKey)

	if err != nil || publicKey == "" {
		fail("Created the key but could not read " + keyPath + ".pub.")
		return
	}

	c.RenderWithoutLayout("sites/_deploykey", gonja.Context{
		"siteId":      site.ID,
		"publicKey":   publicKey,
		"_csrf_token": c.SetAndGetCSRFToken(gctx),
	}, gctx)
}

func (c *Controller) DeployBranch(gctx *gin.Context) {
	site := c.findSessionTeamSite(gctx, gctx.PostForm("siteId"))

	if site.ID == 0 {
		c.FlashError(gctx, "Site ID is required")
		gctx.Redirect(http.StatusFound, "/sites")
		return
	}

	go console.RunSiteScriptables(c.GetDB(gctx), site, utils.FindScriptables(site.DeployScriptables))

	c.FlashSuccess(gctx, "Success! deploy will begin shortly...")
	gctx.Redirect(http.StatusFound, fmt.Sprintf("/logs/site/%d", site.ID))
}

func (c *Controller) queueSiteBuild(gctx *gin.Context, successMsg string) {
	site := c.findSessionTeamSite(gctx, gctx.PostForm("siteId"))

	if site.ID == 0 {
		c.FlashError(gctx, "Site ID is invalid or an unknown error has occurred. Please try again.")
		gctx.Redirect(http.StatusFound, "/sites")
		return
	}

	models.SetSiteStatus(c.GetDB(gctx), site.ID, models.STATUS_QUEUED)
	c.FlashSuccess(gctx, successMsg)
	gctx.Redirect(http.StatusFound, fmt.Sprintf("/logs/site/%d", site.ID))
}

func (c *Controller) ConfirmSiteDeploy(gctx *gin.Context) {
	c.queueSiteBuild(gctx, "Success! deploy will begin shortly...")
}

func (c *Controller) RetrySiteBuild(gctx *gin.Context) {
	c.queueSiteBuild(gctx, "Successfully queued site for re-deploy. Please check the logs for more information.")
}
