package console

import (
	"fmt"
	"os"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/models"
	"plexscriptables.com/scriptables/utils"
)

type scriptableRun struct {
	db               *gorm.DB
	entity           string
	entityID         int64
	teamID           int64
	replaceVariables func(script string) string
}

func (run scriptableRun) logError(log string, summary string) {
	models.LogError(run.db, run.entityID, run.entity, log, summary, run.teamID)
}

func (run scriptableRun) prepareScript(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	script, err := utils.ExpandScriptableImports(string(contents))
	return run.replaceVariables(script), err
}

func (run scriptableRun) runAllScripts(scripts []string) bool {
	succeeded := true

	for _, path := range scripts {
		if utils.LogVerbose() {
			fmt.Println("Running scriptable:", path, "for", run.entity, run.entityID)
		}

		script, err := run.prepareScript(path)
		if err != nil {
			run.logError(err.Error(), "Failed to run: "+path)
			return false
		}

		output, err := utils.RunScript(script)
		if err != nil {
			succeeded = false
			run.logError(err.Error()+". Command output: "+output, "Failed to run: "+path)

			if utils.MustStopOnFailure(script) {
				break
			}

			continue
		}

		models.LogInfo(run.db, run.entityID, run.entity, output, "Successfully ran: "+path, run.teamID)
	}

	return succeeded
}

func statusForOutcome(succeeded bool) string {
	if succeeded {
		return models.STATUS_COMPLETE
	}

	return models.STATUS_FAILED
}
