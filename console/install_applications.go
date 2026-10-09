package console

import (
	"gorm.io/gorm"
	"plexscriptables.com/scriptables/models"
	"plexscriptables.com/scriptables/utils"
)

const applicationLogEntity = "application"

func installApplication(db *gorm.DB, application models.Application) {
	models.SetApplicationStatus(db, application.ID, models.STATUS_RUNNING)

	run := scriptableRun{
		db:       db,
		entity:   applicationLogEntity,
		entityID: application.ID,
		teamID:   application.TeamId,
	}

	catalogApplication, found := models.FindCatalogApplication(application.Slug)
	if !found {
		run.logError("Unknown application: "+application.Slug, "No scripts to run.")
		models.SetApplicationStatus(db, application.ID, models.STATUS_FAILED)
		return
	}

	scripts := utils.FindScriptables(catalogApplication.Scriptable)
	if len(scripts) == 0 {
		run.logError("No scriptables found for application: "+application.Slug, "No scripts to run.")
		models.SetApplicationStatus(db, application.ID, models.STATUS_FAILED)
		return
	}

	rootPassword := application.RootPassword()
	run.replaceVariables = func(script string) string {
		return catalogApplication.ReplaceScriptableVariables(script, rootPassword)
	}

	succeeded := run.runAllScripts(scripts)
	models.ForgetApplicationSecret(db, application.ID)
	models.SetApplicationStatus(db, application.ID, statusForOutcome(succeeded))
}

func InstallQueuedApplications(db *gorm.DB) {
	for _, application := range models.GetQueuedApplications(db) {
		installApplication(db, application)
	}
}
