package console

import (
	"os"

	"gorm.io/gorm"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

func isPanelListeningPublicly() bool {
	host := os.Getenv("SCRIPTABLES_SERVER_DSN_HOST")
	return host != "127.0.0.1" && host != "localhost" && host != "::1"
}

func publicPanelPort() string {
	if !isPanelListeningPublicly() {
		return ""
	}

	return utils.PanelPort()
}

func applySecuritySetting(db *gorm.DB, setting models.SecuritySetting) {
	models.SetSecuritySettingStatus(db, setting.ID, models.STATUS_RUNNING)

	run := scriptableRun{
		db:       db,
		entity:   models.SECURITY_LOG_ENTITY,
		entityID: setting.ID,
		teamID:   setting.TeamId,
		replaceVariables: func(script string) string {
			return setting.ReplaceScriptableVariables(script, publicPanelPort())
		},
	}

	scripts := utils.FindScriptables("security")
	if len(scripts) == 0 {
		run.logError("No scriptables found in scriptables/security.", "No scripts to run.")
		models.SetSecuritySettingStatus(db, setting.ID, models.STATUS_FAILED)
		return
	}

	succeeded := run.runAllScripts(scripts)
	if succeeded {
		models.ForgetSecurityPasswordHash(db, setting.ID)
	}

	models.SetSecuritySettingStatus(db, setting.ID, statusForOutcome(succeeded))
}

func ApplyQueuedSecuritySettings(db *gorm.DB) {
	for _, setting := range models.GetQueuedSecuritySettings(db) {
		applySecuritySetting(db, setting)
	}
}
