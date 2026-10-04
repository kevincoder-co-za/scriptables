package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"plexcorp.tech/scriptable/console"
	"plexcorp.tech/scriptable/controllers"
	"plexcorp.tech/scriptable/middleware"
	"plexcorp.tech/scriptable/models"
	"plexcorp.tech/scriptable/utils"
)

const queuePollInterval = 5 * time.Second

func runSafely(job func(db *gorm.DB), db *gorm.DB) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Caught and recovered from job runner crash:", r)
		}
	}()

	job(db)
}

func runForever(db *gorm.DB, jobs ...func(db *gorm.DB)) {
	for {
		for _, job := range jobs {
			runSafely(job, db)
		}

		time.Sleep(queuePollInterval)
	}
}

func startQueueWorkers(db *gorm.DB) {
	go runForever(db, console.InstallQueuedApplications, console.ApplyQueuedSecuritySettings, console.BuildQueuedSites, console.DeployQueuedSites)
	go runForever(db, console.SyncQueuedCrons)
}

func loadEnv() {
	path := os.Getenv("SCRIPTABLES_ENV_FILE")
	if path == "" {
		path = ".env"
	}

	if err := godotenv.Load(path); err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Could not read", path, "-", err)
			return
		}

		fmt.Println("No", path, "file found, relying on the environment.")
	}
}

func listenAddress() string {
	return os.Getenv("SCRIPTABLES_SERVER_DSN_HOST") + ":" + utils.PanelPort()
}

func trustedProxies() []string {
	var proxies []string
	for _, ip := range strings.Split(os.Getenv("ALLOWED_IPS"), ",") {
		if ip = strings.TrimSpace(ip); ip != "" {
			proxies = append(proxies, ip)
		}
	}

	return proxies
}

func registerRoutes(router *gin.Engine) {
	controller := controllers.Controller{}

	router.GET("/trial-expired", controller.TrialExpired)
	router.GET("/users/logout", controller.Logout)
	router.GET("/users/login", controller.LoginView)
	router.POST("/users/authenticate", controller.CheckLogin)
	router.POST("/users/2fa/authenticate", controller.TwoFactorAuthenticate)
	router.Any("/users/password/reset/:token", controller.ChangePassword)
	router.Any("/users/password/forgot", controller.ForgotPassword)
	router.POST("/users/register/complete", controller.RegistrationComplete)
	router.GET("/users/register", controller.RegisterForm)

	router.GET("/user/list", controller.ListUsers)
	router.POST("/user/actions", controller.HandleUserActionsFormPost)
	router.POST("/user/profile/update", controller.UpdateProfile)
	router.GET("/user/profile", controller.MyProfile)
	router.GET("/user/create", controller.NewUser)
	router.GET("/user/2factor/qrcode", controller.ShowQrCodePng)

	router.GET("/denied", controller.AccessDenied)

	router.GET("/", controller.Applications)
	router.GET("/applications", controller.Applications)
	router.POST("/application/install", controller.InstallApplication)

	router.GET("/logs/:entity/:id", controller.EntityLogs)
	router.GET("/log/full/:entity/:id", controller.FullLog)

	router.GET("/firewall", controller.Firewall)
	router.GET("/firewall/rules", controller.FirewallRules)
	router.POST("/firewall/rule/add", controller.AddFirewallRule)
	router.POST("/firewall/rule/delete", controller.DeleteFirewallRule)

	router.GET("/security", controller.Security)
	router.POST("/security/apply", controller.ApplySecuritySettings)

	router.GET("/site/deployKey/:id", controller.CreateSiteDeployKey)
	router.POST("/site/generateDeployKey", controller.GenerateDeployKey)
	router.POST("/site/deploy/", controller.DeployBranch)
	router.POST("/site/retrybuild", controller.RetrySiteBuild)
	router.POST("/site/confirm-deploy", controller.ConfirmSiteDeploy)
	router.GET("/sites", controller.Sites)
	router.GET("/site/create", controller.CreateSite)
	router.POST("/site/save", controller.SaveSite)

	router.GET("/crons", controller.Crons)
	router.GET("/cron/create", controller.CreateCron)
	router.POST("/cron/save", controller.SaveCron)
	router.GET("/cron/edit/:id", controller.EditCron)
	router.POST("/cron/update/:id", controller.UpdateCron)
	router.POST("/cron/disable/", controller.DisableCron)
	router.POST("/cron/retrybuild", controller.RetryCronBuild)

	router.GET("/webhooks/deploy/:sid/:token", controller.DeployWebhookSite)
}

func main() {
	loadEnv()

	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	location, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		fmt.Println("Timezone entered is invalid:", err)
		return
	}

	time.Local = location

	if os.Getenv("SESSION_SECRET") == "" {
		fmt.Println("SESSION_SECRET is not set. Add a long random value to your .env file.")
		return
	}

	db, err := models.OpenDatabase()
	if err != nil {
		fmt.Println("Could not open the sqlite database:", err)
		return
	}

	models.FailInterruptedJobs(db)
	startQueueWorkers(db)

	router := gin.Default()
	router.StaticFS("/static", http.Dir("./static"))

	if err := router.SetTrustedProxies(trustedProxies()); err != nil {
		fmt.Println("Invalid ALLOWED_IPS:", err)
		return
	}

	router.Use(middleware.DBMiddleware(db))
	router.Use(middleware.SetupSession())
	router.Use(middleware.AuthMiddleware())

	registerRoutes(router)

	router.Run(listenAddress())
}
