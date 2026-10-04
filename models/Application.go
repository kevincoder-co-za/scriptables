package models

import (
	"os/exec"
	"strings"
	"time"

	"gorm.io/gorm"
)

const STATUS_AVAILABLE = "available"

type Application struct {
	gorm.Model
	ID        int64     `gorm:"column:id"`
	Slug      string    `gorm:"column:slug;type:varchar(100);uniqueIndex"`
	Status    string    `gorm:"column:status;type:varchar(100)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	TeamId    int64     `gorm:"column:team_id"`
}

type CatalogApplication struct {
	Slug        string
	Name        string
	Description string
	Icon        string
	Package     string
	Scriptable  string
	Variables   map[string]string
}

type ApplicationOverview struct {
	CatalogApplication
	ID     int64
	Status string
}

func phpCatalogApplication(version string) CatalogApplication {
	return CatalogApplication{
		Slug:        "php" + version,
		Name:        "PHP " + version,
		Description: "PHP " + version + " with FPM, Composer and the extensions most frameworks need.",
		Icon:        "brand-php",
		Package:     "php" + version + "-fpm",
		Scriptable:  "php",
		Variables: map[string]string{
			"#PHP_VERSION#": version,
			"#FPM_PORT#":    PhpFpmPort(version),
		},
	}
}

func GetApplicationCatalog() []CatalogApplication {
	return []CatalogApplication{
		{Slug: "nginx", Name: "Nginx", Description: "High performance web server and reverse proxy.", Icon: "server-2", Package: "nginx", Scriptable: "nginx"},
		{Slug: "apache", Name: "Apache", Description: "The Apache HTTP server. Listens on port 8080 when Nginx already owns port 80.", Icon: "feather", Package: "apache2", Scriptable: "apache"},
		{Slug: "mysql", Name: "MySQL", Description: "MySQL database server, reachable from this machine only.", Icon: "brand-mysql", Package: "mysql-server", Scriptable: "mysql"},
		{Slug: "mariadb", Name: "MariaDB", Description: "MariaDB database server, a drop-in replacement for MySQL.", Icon: "database", Package: "mariadb-server", Scriptable: "mariadb"},
		{Slug: "postgresql", Name: "PostgreSQL", Description: "PostgreSQL relational database server.", Icon: "database", Package: "postgresql", Scriptable: "postgresql"},
		{Slug: "redis", Name: "Redis", Description: "In-memory store for caching, sessions and queues.", Icon: "bolt", Package: "redis-server", Scriptable: "redis"},
		{Slug: "memcached", Name: "Memcached", Description: "Distributed memory object caching system.", Icon: "stack-2", Package: "memcached", Scriptable: "memcached"},
		phpCatalogApplication("8.5"),
		phpCatalogApplication("8.4"),
		phpCatalogApplication("8.3"),
		{Slug: "nodejs", Name: "Node.js", Description: "The current Node.js LTS release with npm.", Icon: "brand-nodejs", Package: "nodejs", Scriptable: "nodejs"},
		{Slug: "docker", Name: "Docker", Description: "Docker engine for running containers.", Icon: "brand-docker", Package: "docker-ce", Scriptable: "docker"},
		{Slug: "certbot", Name: "Certbot", Description: "Free Let's Encrypt SSL certificates with automatic renewal.", Icon: "certificate", Package: "certbot", Scriptable: "certbot"},
		{Slug: "supervisor", Name: "Supervisor", Description: "Process manager for queue workers and long running jobs.", Icon: "binary-tree", Package: "supervisor", Scriptable: "supervisor"},
		{Slug: "fail2ban", Name: "Fail2ban", Description: "Bans IP addresses that repeatedly fail to log in over SSH.", Icon: "shield-lock", Package: "fail2ban", Scriptable: "fail2ban"},
		{Slug: "nvidia_drivers", Name: "NVIDIA drivers", Description: "GPU drivers and the CUDA toolkit for Ubuntu 22.04.", Icon: "cpu", Package: "cuda-toolkit-12-3", Scriptable: "nvidia_drivers"},
		{Slug: "nvidia_container_toolkit", Name: "NVIDIA container toolkit", Description: "Run GPU accelerated Docker containers.", Icon: "cpu", Package: "nvidia-container-toolkit", Scriptable: "nvidia_container_toolkit"},
	}
}

func FindCatalogApplication(slug string) (CatalogApplication, bool) {
	for _, application := range GetApplicationCatalog() {
		if application.Slug == slug {
			return application, true
		}
	}

	return CatalogApplication{}, false
}

func (application CatalogApplication) ReplaceScriptableVariables(script string) string {
	for placeholder, value := range application.Variables {
		script = strings.ReplaceAll(script, placeholder, value)
	}

	return script
}

func findInstalledPackages(packages []string) map[string]bool {
	args := append([]string{"-W", "-f=${Package} ${db:Status-Status}\n"}, packages...)
	output, _ := exec.Command("dpkg-query", args...).Output()

	installed := map[string]bool{}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == "installed" {
			installed[fields[0]] = true
		}
	}

	return installed
}

func IsPackageInstalled(name string) bool {
	return findInstalledPackages([]string{name})[name]
}

func isInstallInProgressOrFailed(status string) bool {
	return status == STATUS_QUEUED || status == STATUS_RUNNING || status == STATUS_FAILED
}

func GetApplicationsOverview(db *gorm.DB) []ApplicationOverview {
	catalog := GetApplicationCatalog()

	packages := []string{}
	for _, application := range catalog {
		packages = append(packages, application.Package)
	}
	installed := findInstalledPackages(packages)

	var records []Application
	db.Find(&records)
	recordsBySlug := map[string]Application{}
	for _, record := range records {
		recordsBySlug[record.Slug] = record
	}

	overview := []ApplicationOverview{}
	for _, application := range catalog {
		record := recordsBySlug[application.Slug]

		status := STATUS_AVAILABLE
		if installed[application.Package] {
			status = STATUS_COMPLETE
		}
		if isInstallInProgressOrFailed(record.Status) {
			status = record.Status
		}

		overview = append(overview, ApplicationOverview{CatalogApplication: application, ID: record.ID, Status: status})
	}

	return overview
}

func QueueApplicationInstall(db *gorm.DB, slug string, teamId int64) Application {
	var application Application
	db.Where("slug = ?", slug).Limit(1).Find(&application)

	application.Slug = slug
	application.Status = STATUS_QUEUED
	application.TeamId = teamId
	application.UpdatedAt = time.Now()
	if application.ID == 0 {
		application.CreatedAt = time.Now()
	}

	db.Save(&application)
	return application
}

func GetQueuedApplications(db *gorm.DB) []Application {
	var applications []Application
	db.Where("status = ?", STATUS_QUEUED).Order("updated_at asc").Find(&applications)
	return applications
}

func SetApplicationStatus(db *gorm.DB, id int64, status string) {
	db.Model(&Application{}).Where("id = ?", id).Update("status", status)
}
