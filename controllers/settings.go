package controllers

import (
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexscriptables.com/scriptables/models"
	"plexscriptables.com/scriptables/utils"
)

var domainNamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`)
var hostNamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`)

func looksLikeEmail(email string) bool {
	return len(email) >= 5 && strings.Contains(email, "@") && !strings.ContainsAny(email, " \r\n")
}

type smtpForm struct {
	Host      string
	Port      string
	Username  string
	Password  string
	FromEmail string
}

func readSmtpForm(gctx *gin.Context) smtpForm {
	return smtpForm{
		Host:      strings.TrimSpace(gctx.PostForm("smtp_host")),
		Port:      strings.TrimSpace(gctx.PostForm("smtp_port")),
		Username:  strings.TrimSpace(gctx.PostForm("smtp_username")),
		Password:  gctx.PostForm("smtp_password"),
		FromEmail: strings.TrimSpace(gctx.PostForm("smtp_from_email")),
	}
}

func smtpFormFromSetting(setting models.Setting) smtpForm {
	config := setting.MailConfig()
	return smtpForm{Host: config.Host, Port: config.Port, Username: config.Username, FromEmail: config.FromEmail}
}

func (form smtpForm) validate() []string {
	errors := []string{}

	if !hostNamePattern.MatchString(form.Host) {
		errors = append(errors, "Please enter the SMTP server host name, e.g. smtp.mailgun.org.")
	}

	if port, err := strconv.Atoi(form.Port); err != nil || port < 1 || port > 65535 {
		errors = append(errors, "Please enter the SMTP port, usually 587 or 465.")
	}

	if form.Username == "" {
		errors = append(errors, "Please enter the SMTP username.")
	}

	if _, err := mail.ParseAddress(form.FromEmail); err != nil {
		errors = append(errors, "Please enter the address mail should be sent from, e.g. Scriptables <noreply@example.com>.")
	}

	return errors
}

func (form smtpForm) mailConfig(setting models.Setting) utils.MailConfig {
	password := form.Password
	if password == "" {
		password = setting.DecryptedSmtpPassword()
	}

	return utils.MailConfig{
		Host:      form.Host,
		Port:      form.Port,
		Username:  form.Username,
		Password:  password,
		FromEmail: form.FromEmail,
		BaseUrl:   setting.PublicUrl(),
	}
}

func (c *Controller) renderSettingsPage(gctx *gin.Context, domain string, certificateEmail string, smtp smtpForm, errors []string) {
	setting := models.GetSetting(c.GetDB(gctx))

	vars := gonja.Context{
		"title":             "Settings",
		"highlight":         "settings",
		"setting":           setting,
		"domainProvisioned": models.IsDomainProvisioned(),
		"domainInProgress":  isInProgress(setting.DomainStatus),
		"domain":            domain,
		"certificate_email": certificateEmail,
		"smtp":              smtp,
		"publicUrl":         setting.PublicUrl(),
		"panelPort":         utils.PanelPort(),
	}

	if len(errors) > 0 {
		vars["errors"] = errors
	}

	c.Render("settings/index", vars, gctx)
}

func (c *Controller) Settings(gctx *gin.Context) {
	setting := models.GetSetting(c.GetDB(gctx))

	domain := setting.Domain
	if models.IsDomainProvisioned() {
		domain = models.ProvisionedDomain()
	}

	c.renderSettingsPage(gctx, domain, setting.CertificateEmail, smtpFormFromSetting(setting), nil)
}

func (c *Controller) ApplyDomainSetting(gctx *gin.Context) {
	db := c.GetDB(gctx)
	sessUser := c.GetSessionUser(gctx)
	setting := models.GetSetting(db)

	domain := strings.ToLower(strings.TrimSpace(gctx.PostForm("domain")))
	certificateEmail := strings.TrimSpace(gctx.PostForm("certificate_email"))
	smtp := smtpFormFromSetting(setting)

	if models.IsDomainProvisioned() {
		c.FlashError(gctx, "The domain was set by the installer and cannot be changed here.")
		gctx.Redirect(http.StatusFound, "/settings")
		return
	}

	if isInProgress(setting.DomainStatus) {
		c.FlashError(gctx, "The domain is already being set up. Please wait for that to finish.")
		gctx.Redirect(http.StatusFound, "/settings")
		return
	}

	errors := []string{}
	if !domainNamePattern.MatchString(domain) || !strings.Contains(domain, ".") {
		errors = append(errors, "Please enter a valid domain name such as panel.example.com.")
	}

	if certificateEmail != "" && !looksLikeEmail(certificateEmail) {
		errors = append(errors, "Please enter a valid email address for certificate expiry notices, or leave it empty.")
	}

	if len(errors) > 0 {
		c.renderSettingsPage(gctx, domain, certificateEmail, smtp, errors)
		return
	}

	setting.Domain = domain
	setting.CertificateEmail = certificateEmail
	setting.DomainStatus = models.STATUS_QUEUED
	setting.TeamId = sessUser.TeamId
	models.SaveSetting(db, &setting)

	c.FlashSuccess(gctx, "Domain setup queued. Make sure "+domain+" already points at this server, or the certificate request will fail.")
	gctx.Redirect(http.StatusFound, "/logs/settings/"+strconv.FormatInt(setting.ID, 10))
}

func (c *Controller) openOutgoingSmtpPort(gctx *gin.Context, port string, teamId int64) {
	if _, err := models.GetFirewallState(); err != nil {
		return
	}

	models.AddFirewallRule(c.GetDB(gctx), "allow out "+port+"/tcp", teamId)
}

func (c *Controller) SaveSmtpSettings(gctx *gin.Context) {
	db := c.GetDB(gctx)
	sessUser := c.GetSessionUser(gctx)
	setting := models.GetSetting(db)
	form := readSmtpForm(gctx)

	domain := setting.Domain
	if models.IsDomainProvisioned() {
		domain = models.ProvisionedDomain()
	}

	if errors := form.validate(); len(errors) > 0 {
		c.renderSettingsPage(gctx, domain, setting.CertificateEmail, form, errors)
		return
	}

	if form.Password == "" && setting.DecryptedSmtpPassword() == "" {
		c.renderSettingsPage(gctx, domain, setting.CertificateEmail, form, []string{"Please enter the SMTP password."})
		return
	}

	if gctx.PostForm("action") == "test" {
		vars := gonja.Context{"subject": "Scriptables test email", "name": sessUser.Name}
		err := utils.SendEmail(form.mailConfig(setting), "Scriptables test email", "", []string{sessUser.Email}, vars, "test")
		if err != nil {
			c.renderSettingsPage(gctx, domain, setting.CertificateEmail, form, []string{"Test email failed: " + err.Error()})
			return
		}

		c.FlashSuccess(gctx, "Test email sent to "+sessUser.Email+". These settings are not saved yet.")
		c.renderSettingsPage(gctx, domain, setting.CertificateEmail, form, nil)
		return
	}

	setting.SmtpHost = form.Host
	setting.SmtpPort = form.Port
	setting.SmtpUsername = form.Username
	setting.SmtpFromEmail = form.FromEmail
	if form.Password != "" {
		setting.SmtpPassword = utils.Encrypt(form.Password)
	}
	setting.TeamId = sessUser.TeamId
	models.SaveSetting(db, &setting)

	c.openOutgoingSmtpPort(gctx, form.Port, sessUser.TeamId)

	c.FlashSuccess(gctx, "SMTP settings saved.")
	gctx.Redirect(http.StatusFound, "/settings")
}
