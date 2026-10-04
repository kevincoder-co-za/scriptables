package console

import (
	"gorm.io/gorm"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

const siteLogEntity = "site"

func RunSiteScriptables(db *gorm.DB, site *models.Site, scripts []string) {
	models.SetSiteStatus(db, site.ID, models.STATUS_RUNNING)

	run := scriptableRun{
		db:       db,
		entity:   siteLogEntity,
		entityID: site.ID,
		teamID:   site.TeamId,
		replaceVariables: func(script string) string {
			return site.ReplaceScriptableVariables(db, script)
		},
	}

	if len(scripts) == 0 {
		run.logError("No scriptables found for site: "+site.SiteName, "No scripts to run.")
		models.SetSiteStatus(db, site.ID, models.STATUS_FAILED)
		return
	}

	models.SetSiteStatus(db, site.ID, statusForOutcome(run.runAllScripts(scripts)))
}

func findSiteBuildScripts(site *models.Site) []string {
	scripts := utils.FindScriptables(site.ScriptableName)

	if len(scripts) > 0 && site.LetsEncryptCertificate == 1 {
		scripts = append(scripts, utils.FindScriptables("ssl_setup")...)
	}

	return scripts
}

func BuildQueuedSites(db *gorm.DB) {
	for _, site := range models.GetQueuedSites(db) {
		RunSiteScriptables(db, &site, findSiteBuildScripts(&site))
	}
}

func DeployQueuedSites(db *gorm.DB) {
	for _, siteId := range models.GetSiteIdsQueuedForDeploy(db) {
		models.MarkQueuedDeploysComplete(db, siteId)

		site := models.GetSiteByIdNoTeam(db, siteId)
		if site.ID == 0 {
			continue
		}

		RunSiteScriptables(db, site, utils.FindScriptables(site.DeployScriptables))
	}
}
