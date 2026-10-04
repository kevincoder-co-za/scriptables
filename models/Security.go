package models

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"plexcorp.tech/scriptable/utils"
)

const SECURITY_LOG_ENTITY = "security"
const DEFAULT_SSH_PORT = 22
const SUGGESTED_SSH_PORT = 2022

type SecuritySetting struct {
	gorm.Model
	ID           int64     `gorm:"column:id"`
	SshPort      int       `gorm:"column:ssh_port"`
	SudoUsername string    `gorm:"column:sudo_username;type:varchar(100)"`
	PublicKey    string    `gorm:"column:public_key"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255)"`
	Status       string    `gorm:"column:status;type:varchar(100)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	TeamId       int64     `gorm:"column:team_id"`
}

type SecurityCheck struct {
	Title  string
	Detail string
	Passed bool
}

func GetSecuritySetting(db *gorm.DB) SecuritySetting {
	var setting SecuritySetting
	db.Order("id asc").Limit(1).Find(&setting)
	return setting
}

func GetQueuedSecuritySettings(db *gorm.DB) []SecuritySetting {
	var settings []SecuritySetting
	db.Where("status = ?", STATUS_QUEUED).Find(&settings)
	return settings
}

func SetSecuritySettingStatus(db *gorm.DB, id int64, status string) {
	db.Model(&SecuritySetting{}).Where("id = ?", id).Update("status", status)
}

func ForgetSecurityPasswordHash(db *gorm.DB, id int64) {
	db.Model(&SecuritySetting{}).Where("id = ?", id).Update("password_hash", "")
}

func (setting SecuritySetting) ReplaceScriptableVariables(script string, panelPort string) string {
	script = strings.ReplaceAll(script, "#SSH_PORT#", strconv.Itoa(setting.SshPort))
	script = strings.ReplaceAll(script, "#SUDO_USERNAME#", setting.SudoUsername)
	script = strings.ReplaceAll(script, "#PUBLIC_KEY#", setting.PublicKey)
	script = strings.ReplaceAll(script, "#PASSWORD_HASH#", setting.PasswordHash)
	script = strings.ReplaceAll(script, "#PANEL_PORT#", panelPort)
	return script
}

func readEffectiveSshdConfig() map[string]string {
	config := map[string]string{}

	output, err := utils.RunCommandAsRoot("sshd", "-T")
	if err != nil {
		return config
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if _, alreadySet := config[fields[0]]; !alreadySet {
			config[fields[0]] = fields[1]
		}
	}

	return config
}

func GetCurrentSshPort() int {
	port, err := strconv.Atoi(readEffectiveSshdConfig()["port"])
	if err != nil {
		return DEFAULT_SSH_PORT
	}

	return port
}

const sshdUnreadableDetail = "Could not read the SSH server configuration. Is openssh-server installed?"

func checkSshPortChanged(sshd map[string]string) SecurityCheck {
	check := SecurityCheck{Title: "SSH port changed from the default"}

	port, found := sshd["port"]
	if !found {
		check.Detail = sshdUnreadableDetail
		return check
	}

	check.Passed = port != strconv.Itoa(DEFAULT_SSH_PORT)
	check.Detail = "SSH is listening on port " + port + "."
	return check
}

func checkRootLoginDisabled(sshd map[string]string) SecurityCheck {
	check := SecurityCheck{Title: "Root login over SSH disabled"}

	permitRootLogin, found := sshd["permitrootlogin"]
	if !found {
		check.Detail = sshdUnreadableDetail
		return check
	}

	check.Passed = permitRootLogin == "no"
	check.Detail = "PermitRootLogin is set to " + permitRootLogin + "."
	return check
}

func checkPasswordLoginDisabled(sshd map[string]string) SecurityCheck {
	check := SecurityCheck{Title: "SSH password login disabled"}

	passwordAuthentication, found := sshd["passwordauthentication"]
	if !found {
		check.Detail = sshdUnreadableDetail
		return check
	}

	check.Passed = passwordAuthentication == "no"
	if check.Passed {
		check.Detail = "Only SSH keys are accepted."
	} else {
		check.Detail = "Passwords are still accepted over SSH."
	}
	return check
}

func SystemUserExists(username string) bool {
	return exec.Command("id", username).Run() == nil
}

func SystemUserHasPassword(username string) bool {
	output, err := utils.RunCommandAsRoot("passwd", "-S", username)
	fields := strings.Fields(output)
	return err == nil && len(fields) > 1 && fields[1] == "P"
}

func isSystemUserInSudoGroup(username string) bool {
	output, _ := exec.Command("id", "-nG", username).Output()
	for _, group := range strings.Fields(string(output)) {
		if group == "sudo" {
			return true
		}
	}

	return false
}

func canSystemUserSudoWithoutPassword(username string) bool {
	output, _ := utils.RunCommandAsRoot("sudo", "-l", "-U", username)
	return strings.Contains(output, "NOPASSWD")
}

func checkSudoUser(username string) SecurityCheck {
	check := SecurityCheck{Title: "Dedicated password protected sudo user"}

	switch {
	case username == "":
		check.Detail = "No sudo user has been set up yet."
	case !SystemUserExists(username):
		check.Detail = "The user " + username + " does not exist."
	case !isSystemUserInSudoGroup(username):
		check.Detail = username + " is not allowed to use sudo."
	case !SystemUserHasPassword(username):
		check.Detail = username + " has no password set."
	case canSystemUserSudoWithoutPassword(username):
		check.Detail = username + " can use sudo without a password."
	default:
		check.Passed = true
		check.Detail = username + " logs in with an SSH key and needs a password for sudo."
	}

	return check
}

func checkFail2ban() SecurityCheck {
	check := SecurityCheck{Title: "Fail2ban protecting SSH"}

	if exec.Command("systemctl", "is-active", "--quiet", "fail2ban").Run() != nil {
		check.Detail = "Fail2ban is not running."
		return check
	}

	if _, err := utils.RunCommandAsRoot("fail2ban-client", "status", "sshd"); err != nil {
		check.Detail = "Fail2ban is running but the sshd jail is not enabled."
		return check
	}

	check.Passed = true
	check.Detail = "Repeated failed SSH logins are banned."
	return check
}

func checkFirewall() SecurityCheck {
	check := SecurityCheck{Title: "Firewall enabled"}

	state, err := GetFirewallState()
	switch {
	case err != nil:
		check.Detail = "Could not read the firewall status. Is ufw installed?"
	case !state.Active:
		check.Detail = "ufw is installed but inactive."
	default:
		check.Passed = true
		check.Detail = "ufw is active."
	}

	return check
}

func GetSecurityChecks(sudoUsername string) []SecurityCheck {
	sshd := readEffectiveSshdConfig()

	return []SecurityCheck{
		checkSshPortChanged(sshd),
		checkRootLoginDisabled(sshd),
		checkPasswordLoginDisabled(sshd),
		checkSudoUser(sudoUsername),
		checkFail2ban(),
		checkFirewall(),
	}
}
