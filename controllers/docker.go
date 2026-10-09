package controllers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noirbizarre/gonja"
	"plexscriptables.com/scriptables/models"
)

var dockerNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
var dockerReferencePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.:/@-]*$`)
var dockerPortPattern = regexp.MustCompile(`^(\d{1,3}(\.\d{1,3}){3}:)?\d{1,5}:\d{1,5}(/(tcp|udp))?$`)
var dockerVolumePattern = regexp.MustCompile(`^(/[^:\s]*|[a-zA-Z0-9][a-zA-Z0-9_.-]*):/[^:\s]*(:(ro|rw))?$`)
var dockerEnvPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.*$`)

var dockerRestartPolicies = []string{"no", "always", "unless-stopped", "on-failure"}
var dockerNetworkDrivers = []string{"bridge", "overlay", "macvlan", "ipvlan"}

const dockerTabContainers = "containers"
const dockerTabImages = "images"
const dockerTabNetworks = "networks"

func isOneOf(value string, allowed []string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}

	return false
}

func splitLines(text string) []string {
	lines := []string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func (c *Controller) Docker(gctx *gin.Context) {
	if !models.IsDockerInstalled() {
		c.Render("docker/not_installed", gonja.Context{
			"title":     "Docker",
			"highlight": "docker",
		}, gctx)
		return
	}

	tab := gctx.Query("tab")
	if !isOneOf(tab, []string{dockerTabContainers, dockerTabImages, dockerTabNetworks}) {
		tab = dockerTabContainers
	}

	diskUsage, _ := models.GetDockerDiskUsage()

	c.Render("docker/index", gonja.Context{
		"title":     "Docker",
		"highlight": "docker",
		"tab":       tab,
		"diskUsage": diskUsage,
	}, gctx)
}

func (c *Controller) renderDockerTab(gctx *gin.Context, tab string, successMsg string, actionErr error) {
	vars := gonja.Context{"_csrf_token": c.SetAndGetCSRFToken(gctx)}
	if successMsg != "" {
		vars["successMsg"] = successMsg
	}

	var fetchErr error
	switch tab {
	case dockerTabImages:
		vars["images"], fetchErr = models.ListDockerImages()
	case dockerTabNetworks:
		vars["networks"], fetchErr = models.ListDockerNetworks()
		vars["drivers"] = dockerNetworkDrivers
	default:
		vars["containers"], fetchErr = models.ListDockerContainers()
	}

	if actionErr != nil {
		vars["errorMsg"] = actionErr.Error()
	} else if fetchErr != nil {
		vars["errorMsg"] = fetchErr.Error()
	}

	c.RenderWithoutLayout("docker/_"+tab, vars, gctx)
}

func (c *Controller) DockerContainers(gctx *gin.Context) {
	c.renderDockerTab(gctx, dockerTabContainers, "", nil)
}

func (c *Controller) DockerImages(gctx *gin.Context) {
	c.renderDockerTab(gctx, dockerTabImages, "", nil)
}

func (c *Controller) DockerNetworks(gctx *gin.Context) {
	c.renderDockerTab(gctx, dockerTabNetworks, "", nil)
}

func (c *Controller) DockerContainerAction(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	db := c.GetDB(gctx)
	id := strings.TrimSpace(gctx.PostForm("id"))
	action := gctx.PostForm("action")

	if !dockerNamePattern.MatchString(id) {
		c.renderDockerTab(gctx, dockerTabContainers, "", errors.New("Bad container ID."))
		return
	}

	var err error
	switch action {
	case "start":
		err = models.StartDockerContainer(db, id, sessUser.TeamId)
	case "stop":
		err = models.StopDockerContainer(db, id, sessUser.TeamId)
	case "restart":
		err = models.RestartDockerContainer(db, id, sessUser.TeamId)
	case "delete":
		err = models.RemoveDockerContainer(db, id, sessUser.TeamId)
	default:
		err = errors.New("Unknown container action.")
	}

	if err != nil {
		c.renderDockerTab(gctx, dockerTabContainers, "", err)
		return
	}

	c.renderDockerTab(gctx, dockerTabContainers, "Container "+action+" succeeded.", nil)
}

func (c *Controller) DockerPullImage(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	image := strings.TrimSpace(gctx.PostForm("image"))

	if !dockerReferencePattern.MatchString(image) {
		c.renderDockerTab(gctx, dockerTabImages, "", errors.New("Please enter an image name such as nginx:latest or ghcr.io/org/app:1.0."))
		return
	}

	if err := models.PullDockerImage(c.GetDB(gctx), image, sessUser.TeamId); err != nil {
		c.renderDockerTab(gctx, dockerTabImages, "", err)
		return
	}

	c.renderDockerTab(gctx, dockerTabImages, "Pulled "+image+".", nil)
}

func (c *Controller) DockerDeleteImage(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	id := strings.TrimSpace(gctx.PostForm("id"))

	if !dockerReferencePattern.MatchString(id) {
		c.renderDockerTab(gctx, dockerTabImages, "", errors.New("Bad image ID."))
		return
	}

	if err := models.RemoveDockerImage(c.GetDB(gctx), id, sessUser.TeamId); err != nil {
		c.renderDockerTab(gctx, dockerTabImages, "", err)
		return
	}

	c.renderDockerTab(gctx, dockerTabImages, "Removed the image.", nil)
}

func (c *Controller) DockerCreateNetwork(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	name := strings.TrimSpace(gctx.PostForm("name"))
	driver := gctx.PostForm("driver")

	if !dockerNamePattern.MatchString(name) {
		c.renderDockerTab(gctx, dockerTabNetworks, "", errors.New("Network names may only contain letters, numbers, dots, dashes and underscores."))
		return
	}

	if !isOneOf(driver, dockerNetworkDrivers) {
		c.renderDockerTab(gctx, dockerTabNetworks, "", errors.New("Please choose a network driver."))
		return
	}

	if err := models.CreateDockerNetwork(c.GetDB(gctx), name, driver, sessUser.TeamId); err != nil {
		c.renderDockerTab(gctx, dockerTabNetworks, "", err)
		return
	}

	c.renderDockerTab(gctx, dockerTabNetworks, "Created network "+name+".", nil)
}

func (c *Controller) DockerDeleteNetwork(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	id := strings.TrimSpace(gctx.PostForm("id"))

	if !dockerNamePattern.MatchString(id) {
		c.renderDockerTab(gctx, dockerTabNetworks, "", errors.New("Bad network ID."))
		return
	}

	if err := models.RemoveDockerNetwork(c.GetDB(gctx), id, sessUser.TeamId); err != nil {
		c.renderDockerTab(gctx, dockerTabNetworks, "", err)
		return
	}

	c.renderDockerTab(gctx, dockerTabNetworks, "Removed the network.", nil)
}

func (c *Controller) DockerPrune(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)

	options := models.DockerPruneOptions{
		AllImages: gctx.PostForm("all_images") == "1",
		Volumes:   gctx.PostForm("volumes") == "1",
	}

	output, err := models.PruneDockerSystem(c.GetDB(gctx), options, sessUser.TeamId)
	if err != nil {
		c.FlashError(gctx, "Prune failed: "+err.Error())
	} else {
		c.FlashSuccess(gctx, "Prune finished. "+lastLine(output))
	}

	gctx.Redirect(http.StatusFound, "/docker")
}

func lastLine(output string) string {
	lines := splitLines(output)
	if len(lines) == 0 {
		return ""
	}

	return lines[len(lines)-1]
}

type dockerRunForm struct {
	Name          string
	Image         string
	RestartPolicy string
	Network       string
	NewNetwork    string
	Ports         string
	Volumes       string
	Environment   string
	Command       string
}

func readDockerRunForm(gctx *gin.Context) dockerRunForm {
	return dockerRunForm{
		Name:          strings.TrimSpace(gctx.PostForm("name")),
		Image:         strings.TrimSpace(gctx.PostForm("image")),
		RestartPolicy: gctx.PostForm("restart_policy"),
		Network:       strings.TrimSpace(gctx.PostForm("network")),
		NewNetwork:    strings.TrimSpace(gctx.PostForm("new_network")),
		Ports:         gctx.PostForm("ports"),
		Volumes:       gctx.PostForm("volumes"),
		Environment:   gctx.PostForm("environment"),
		Command:       strings.TrimSpace(gctx.PostForm("command")),
	}
}

func validateEachLine(text string, pattern *regexp.Regexp, message string) ([]string, []string) {
	lines := splitLines(text)
	errors := []string{}

	for _, line := range lines {
		if !pattern.MatchString(line) {
			errors = append(errors, message+": "+line)
		}
	}

	return lines, errors
}

func (form dockerRunForm) networkName() string {
	if form.Network == "new" {
		return form.NewNetwork
	}

	return form.Network
}

func (form dockerRunForm) validate() (models.DockerRunSpec, []string) {
	errors := []string{}

	if !dockerNamePattern.MatchString(form.Name) {
		errors = append(errors, "Container names may only contain letters, numbers, dots, dashes and underscores.")
	}

	if !dockerReferencePattern.MatchString(form.Image) {
		errors = append(errors, "Please enter an image name such as nginx:latest or ghcr.io/org/app:1.0.")
	}

	if !isOneOf(form.RestartPolicy, dockerRestartPolicies) {
		errors = append(errors, "Please choose a restart policy.")
	}

	network := form.networkName()
	if network != "" && !dockerNamePattern.MatchString(network) {
		errors = append(errors, "Network names may only contain letters, numbers, dots, dashes and underscores.")
	}

	ports, portErrors := validateEachLine(form.Ports, dockerPortPattern, "Ports must look like 8080:80 or 127.0.0.1:8080:80/tcp")
	volumes, volumeErrors := validateEachLine(form.Volumes, dockerVolumePattern, "Volumes must look like /host/path:/container/path or data:/var/lib/data")
	environment, envErrors := validateEachLine(form.Environment, dockerEnvPattern, "Environment variables must look like KEY=value")
	errors = append(errors, portErrors...)
	errors = append(errors, volumeErrors...)
	errors = append(errors, envErrors...)

	if strings.HasPrefix(form.Command, "-") {
		errors = append(errors, "The command cannot start with a dash.")
	}

	return models.DockerRunSpec{
		Name:          form.Name,
		Image:         form.Image,
		RestartPolicy: form.RestartPolicy,
		Network:       network,
		Ports:         ports,
		Volumes:       volumes,
		Environment:   environment,
		Command:       form.Command,
	}, errors
}

func (c *Controller) renderDockerRunForm(gctx *gin.Context, form dockerRunForm, errors []string) {
	networks, _ := models.ListDockerNetworks()

	vars := gonja.Context{
		"title":           "Create a container",
		"highlight":       "docker",
		"form":            form,
		"networks":        networks,
		"restartPolicies": dockerRestartPolicies,
	}

	if len(errors) > 0 {
		vars["errors"] = errors
	}

	c.Render("docker/run", vars, gctx)
}

func (c *Controller) DockerCreateContainer(gctx *gin.Context) {
	c.renderDockerRunForm(gctx, dockerRunForm{RestartPolicy: "unless-stopped", Network: "bridge"}, nil)
}

func (c *Controller) DockerRunContainer(gctx *gin.Context) {
	sessUser := c.GetSessionUser(gctx)
	db := c.GetDB(gctx)
	form := readDockerRunForm(gctx)

	spec, errors := form.validate()
	if len(errors) > 0 {
		c.renderDockerRunForm(gctx, form, errors)
		return
	}

	if form.Network == "new" && !models.DockerNetworkExists(spec.Network) {
		if err := models.CreateDockerNetwork(db, spec.Network, "bridge", sessUser.TeamId); err != nil {
			c.renderDockerRunForm(gctx, form, []string{"Could not create the network: " + err.Error()})
			return
		}
	}

	if err := models.RunDockerContainer(db, spec, sessUser.TeamId); err != nil {
		c.renderDockerRunForm(gctx, form, []string{"docker run failed: " + err.Error()})
		return
	}

	c.FlashSuccess(gctx, "Container "+spec.Name+" is up.")
	gctx.Redirect(http.StatusFound, "/docker")
}
