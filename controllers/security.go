package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"golang.org/x/crypto/ssh"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

const minimumSudoPasswordLength = 8

var reservedSudoUsernames = []string{"root", "scriptables", "www-data", "git"}

type securityForm struct {
	SshPort         int
	SudoUsername    string
	PublicKey       string
	Password        string
	PasswordConfirm string
}

func readSecurityForm(gctx *gin.Context) securityForm {
	port, _ := strconv.Atoi(strings.TrimSpace(gctx.PostForm("ssh_port")))

	return securityForm{
		SshPort:         port,
		SudoUsername:    strings.TrimSpace(gctx.PostForm("sudo_username")),
		PublicKey:       strings.TrimSpace(gctx.PostForm("public_key")),
		Password:        gctx.PostForm("password"),
		PasswordConfirm: gctx.PostForm("password_confirm"),
	}
}

func isReservedSudoUsername(username string) bool {
	for _, reserved := range reservedSudoUsernames {
		if username == reserved {
			return true
		}
	}

	return false
}

func isPortUsedByWebTraffic(port int) bool {
	return port == 80 || port == 443 || strconv.Itoa(port) == utils.PanelPort()
}

func isValidPublicKey(publicKey string) bool {
	if strings.ContainsAny(publicKey, "\r\n'") {
		return false
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	return err == nil
}

func (form securityForm) validate() []string {
	errors := []string{}

	if form.SshPort < 1 || form.SshPort > 65535 {
		errors = append(errors, "Please enter an SSH port between 1 and 65535.")
	} else if isPortUsedByWebTraffic(form.SshPort) {
		errors = append(errors, "That port is already used for web traffic. Please pick another SSH port.")
	}

	if len(form.SudoUsername) < 3 || !linuxUsernamePattern.MatchString(form.SudoUsername) {
		errors = append(errors, "Username must be a valid Linux username of at least 3 characters.")
	} else if isReservedSudoUsername(form.SudoUsername) {
		errors = append(errors, "That username is reserved. Please pick another one.")
	}

	if !isValidPublicKey(form.PublicKey) {
		errors = append(errors, "Please paste a single valid SSH public key, e.g. the contents of ~/.ssh/id_ed25519.pub.")
	}

	if form.Password != form.PasswordConfirm {
		errors = append(errors, "Password and confirm password do not match.")
	}

	if form.Password == "" && len(errors) == 0 && !models.SystemUserHasPassword(form.SudoUsername) {
		errors = append(errors, "Please choose a sudo password for this user.")
	}

	if form.Password != "" && len(form.Password) < minimumSudoPasswordLength {
		errors = append(errors, "The sudo password must be at least 8 characters.")
	}

	return errors
}

func suggestSshPort(setting models.SecuritySetting) int {
	if setting.SshPort != 0 {
		return setting.SshPort
	}

	if currentPort := models.GetCurrentSshPort(); currentPort != models.DEFAULT_SSH_PORT {
		return currentPort
	}

	return models.SUGGESTED_SSH_PORT
}

func countPassedChecks(checks []models.SecurityCheck) int {
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		}
	}

	return passed
}

func (c *Controller) renderSecurityPage(gctx *gin.Context, form securityForm, errors []string) {
	setting := models.GetSecuritySetting(c.GetDB(gctx))
	checks := models.GetSecurityChecks(setting.SudoUsername)

	vars := gonja.Context{
		"title":         "Security",
		"highlight":     "security",
		"checks":        checks,
		"numChecks":     len(checks),
		"numPassed":     countPassedChecks(checks),
		"setting":       setting,
		"inProgress":    isInProgress(setting.Status),
		"ssh_port":      form.SshPort,
		"sudo_username": form.SudoUsername,
		"public_key":    form.PublicKey,
	}

	if len(errors) > 0 {
		vars["errors"] = errors
	}

	c.Render("security/index", vars, gctx)
}

func (c *Controller) Security(gctx *gin.Context) {
	setting := models.GetSecuritySetting(c.GetDB(gctx))

	c.renderSecurityPage(gctx, securityForm{
		SshPort:      suggestSshPort(setting),
		SudoUsername: setting.SudoUsername,
		PublicKey:    setting.PublicKey,
	}, nil)
}

func (c *Controller) ApplySecuritySettings(gctx *gin.Context) {
	db := c.GetDB(gctx)
	sessUser := c.GetSessionUser(gctx)
	form := readSecurityForm(gctx)

	setting := models.GetSecuritySetting(db)
	if isInProgress(setting.Status) {
		c.FlashError(gctx, "Security settings are already being applied. Please wait for that to finish.")
		gctx.Redirect(http.StatusFound, "/security")
		return
	}

	if errors := form.validate(); len(errors) > 0 {
		c.renderSecurityPage(gctx, form, errors)
		return
	}

	setting.PasswordHash = ""
	if form.Password != "" {
		passwordHash, err := utils.HashSystemPassword(form.Password)
		if err != nil {
			c.renderSecurityPage(gctx, form, []string{"Could not hash the sudo password: " + err.Error()})
			return
		}

		setting.PasswordHash = passwordHash
	}

	setting.SshPort = form.SshPort
	setting.SudoUsername = form.SudoUsername
	setting.PublicKey = form.PublicKey
	setting.Status = models.STATUS_QUEUED
	setting.TeamId = sessUser.TeamId
	setting.UpdatedAt = time.Now()
	if setting.ID == 0 {
		setting.CreatedAt = time.Now()
	}
	db.Save(&setting)

	c.FlashSuccess(gctx, "Security settings queued. Keep your current SSH session open until you have confirmed the new login works.")
	gctx.Redirect(http.StatusFound, "/logs/security/"+strconv.FormatInt(setting.ID, 10))
}
