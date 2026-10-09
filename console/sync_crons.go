package console

import (
	"os"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/models"
	"plexscriptables.com/scriptables/utils"
)

const cronFilePath = "/etc/cron.d/scriptables"
const cronLogEntity = "cron"

func writeCronFile(cronFile string) (string, error) {
	if cronFile == "" {
		return utils.RunCommandAsRoot("rm", "-f", cronFilePath)
	}

	stagedFile, err := os.CreateTemp("", "scriptables-crons-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(stagedFile.Name())

	if _, err := stagedFile.WriteString(cronFile); err != nil {
		return "", err
	}
	stagedFile.Close()

	return utils.RunCommandAsRoot("install", "-o", "root", "-g", "root", "-m", "644", stagedFile.Name(), cronFilePath)
}

func SyncQueuedCrons(db *gorm.DB) {
	queuedCrons := models.GetQueuedCronsIncludingDisabled(db)
	if len(queuedCrons) == 0 {
		return
	}

	cronFile := models.BuildCronFile(db)
	output, err := writeCronFile(cronFile)

	for _, cron := range queuedCrons {
		if err != nil {
			models.LogError(db, cron.ID, cronLogEntity, err.Error()+". Command output: "+output,
				"Failed to update "+cronFilePath, cron.TeamID)
			models.SetCronStatus(db, cron.ID, models.STATUS_FAILED)
			continue
		}

		models.LogInfo(db, cron.ID, cronLogEntity, output+"\n --- cron list ---\n"+cronFile,
			"Cronfile update log", cron.TeamID)
		models.SetCronStatus(db, cron.ID, models.STATUS_COMPLETE)
	}
}
