package console

import (
	"gorm.io/gorm"
	"plexscriptables.com/scriptables/models"
	"plexscriptables.com/scriptables/utils"
)

func applyDomainSetting(db *gorm.DB, setting models.Setting) {
	models.SetDomainStatus(db, setting.ID, models.STATUS_RUNNING)

	run := scriptableRun{
		db:               db,
		entity:           models.SETTINGS_LOG_ENTITY,
		entityID:         setting.ID,
		teamID:           setting.TeamId,
		replaceVariables: setting.ReplaceScriptableVariables,
	}

	scripts := utils.FindScriptables("panel_domain")
	if len(scripts) == 0 {
		run.logError("No scriptables found in scriptables/panel_domain.", "No scripts to run.")
		models.SetDomainStatus(db, setting.ID, models.STATUS_FAILED)
		return
	}

	models.SetDomainStatus(db, setting.ID, statusForOutcome(run.runAllScripts(scripts)))
}

func ApplyQueuedDomainSettings(db *gorm.DB) {
	for _, setting := range models.GetQueuedDomainSettings(db) {
		applyDomainSetting(db, setting)
	}
}
