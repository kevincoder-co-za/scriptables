package models

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"gorm.io/gorm"
	"plexscriptables.com/scriptables/utils"
)

const DOCKER_LOG_ENTITY = "docker"

type DockerContainer struct {
	ID       string `json:"ID"`
	Name     string `json:"Names"`
	Image    string `json:"Image"`
	State    string `json:"State"`
	Status   string `json:"Status"`
	Ports    string `json:"Ports"`
	Networks string `json:"Networks"`
	Created  string `json:"RunningFor"`
}

type DockerImage struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	Size       string `json:"Size"`
	Created    string `json:"CreatedSince"`
	Containers string `json:"Containers"`
}

type DockerNetwork struct {
	ID     string `json:"ID"`
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
	Scope  string `json:"Scope"`
}

type DockerDiskUsage struct {
	Type        string `json:"Type"`
	Total       string `json:"TotalCount"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

type DockerPruneOptions struct {
	AllImages bool
	Volumes   bool
}

type DockerRunSpec struct {
	Name          string
	Image         string
	RestartPolicy string
	Network       string
	Ports         []string
	Volumes       []string
	Environment   []string
	Command       string
}

var builtInDockerNetworks = map[string]bool{"bridge": true, "host": true, "none": true}

func (container DockerContainer) IsRunning() bool {
	return container.State == "running"
}

func (network DockerNetwork) IsBuiltIn() bool {
	return builtInDockerNetworks[network.Name]
}

func (image DockerImage) Reference() string {
	if image.Repository == "<none>" {
		return image.ID
	}

	return image.Repository + ":" + image.Tag
}

func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func runDocker(args ...string) (string, error) {
	output, err := utils.RunCommandAsRoot("docker", args...)
	if err != nil {
		return output, errors.New(strings.TrimSpace(output))
	}

	return output, nil
}

func decodeDockerLines[T any](output string) []T {
	items := []T{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item T
		if json.Unmarshal([]byte(line), &item) == nil {
			items = append(items, item)
		}
	}

	return items
}

func listDocker[T any](args ...string) ([]T, error) {
	output, err := runDocker(append(args, "--format", "{{json .}}")...)
	if err != nil {
		return []T{}, err
	}

	return decodeDockerLines[T](output), nil
}

func ListDockerContainers() ([]DockerContainer, error) {
	return listDocker[DockerContainer]("ps", "-a")
}

func ListDockerImages() ([]DockerImage, error) {
	return listDocker[DockerImage]("images")
}

func ListDockerNetworks() ([]DockerNetwork, error) {
	return listDocker[DockerNetwork]("network", "ls")
}

func GetDockerDiskUsage() ([]DockerDiskUsage, error) {
	return listDocker[DockerDiskUsage]("system", "df")
}

func DockerNetworkExists(name string) bool {
	networks, err := ListDockerNetworks()
	if err != nil {
		return false
	}

	for _, network := range networks {
		if network.Name == name {
			return true
		}
	}

	return false
}

func loggedDockerCommand(db *gorm.DB, teamId int64, summary string, args ...string) (string, error) {
	output, err := runDocker(args...)
	command := "docker " + strings.Join(args, " ")

	if err != nil {
		LogError(db, 0, DOCKER_LOG_ENTITY, command+"\n"+output, "Failed: "+summary, teamId)
		return output, err
	}

	LogInfo(db, 0, DOCKER_LOG_ENTITY, command+"\n"+output, summary, teamId)
	return output, nil
}

func StartDockerContainer(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Started container "+id, "start", id)
	return err
}

func StopDockerContainer(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Stopped container "+id, "stop", id)
	return err
}

func RestartDockerContainer(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Restarted container "+id, "restart", id)
	return err
}

func RemoveDockerContainer(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Removed container "+id, "rm", "-f", id)
	return err
}

func PullDockerImage(db *gorm.DB, image string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Pulled image "+image, "pull", image)
	return err
}

func RemoveDockerImage(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Removed image "+id, "rmi", id)
	return err
}

func CreateDockerNetwork(db *gorm.DB, name string, driver string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Created network "+name, "network", "create", "--driver", driver, name)
	return err
}

func RemoveDockerNetwork(db *gorm.DB, id string, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Removed network "+id, "network", "rm", id)
	return err
}

func PruneDockerSystem(db *gorm.DB, options DockerPruneOptions, teamId int64) (string, error) {
	args := []string{"system", "prune", "--force"}
	if options.AllImages {
		args = append(args, "--all")
	}
	if options.Volumes {
		args = append(args, "--volumes")
	}

	output, err := loggedDockerCommand(db, teamId, "Pruned the Docker system", args...)
	if err != nil || !options.Volumes {
		return output, err
	}

	volumeOutput, err := loggedDockerCommand(db, teamId, "Pruned unused named volumes", "volume", "prune", "--all", "--force")
	return output + "\n" + volumeOutput, err
}

func (spec DockerRunSpec) Arguments() []string {
	args := []string{"run", "--detach", "--name", spec.Name, "--restart", spec.RestartPolicy}

	if spec.Network != "" {
		args = append(args, "--network", spec.Network)
	}

	for _, port := range spec.Ports {
		args = append(args, "--publish", port)
	}

	for _, volume := range spec.Volumes {
		args = append(args, "--volume", volume)
	}

	for _, variable := range spec.Environment {
		args = append(args, "--env", variable)
	}

	args = append(args, spec.Image)

	if spec.Command != "" {
		args = append(args, strings.Fields(spec.Command)...)
	}

	return args
}

func RunDockerContainer(db *gorm.DB, spec DockerRunSpec, teamId int64) error {
	_, err := loggedDockerCommand(db, teamId, "Created container "+spec.Name+" from "+spec.Image, spec.Arguments()...)
	return err
}
