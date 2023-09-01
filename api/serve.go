package api

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/gofiber/template/html/v2"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/seekr-osint/seekr/api/apierror"
	"github.com/seekr-osint/seekr/api/config"
	"github.com/seekr-osint/seekr/api/seekrauth"
	"github.com/swaggo/fiber-swagger"
)

//	@title			Seekr
//	@version		1.0
//	@description	Seekr api

//	@contact.name	seekr github
//	@contact.url	http://github.com/seekr-osint/seekr
//	@contact.email	seekr-osint@proton.me

//	@license.name	GPL v3
//	@license.url	https://github.com/seekr-osint/seekr/blob/main/LICENSE

//	@host		/api/v1
//	@BasePath	/v1

func Serve(config config.Config, fs embed.FS, db *gorm.DB, users seekrauth.Users) error {
	engine := html.NewFileSystem(http.FS(fs), ".html")

	app := fiber.New(fiber.Config{
		// ServerHeader: "Seekr",
		AppName: "Seekr",
		Views:   engine,
		// Global custom error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(apierror.GlobalErrorHandlerResp{
				Success: false,
				Message: err.Error(),
			})
		},
	})
	fav, err := fs.ReadFile("web/images/favicon.ico")
	if err != nil {
		return err
	}
	app.Use(favicon.New(favicon.Config{
		Data: fav,
	}))

	app.Use(basicauth.New(basicauth.Config{
		Users: users.ToMap(),
	}))
	// Logging remote IP and Port
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Use(helmet.New(helmet.Config{
		XSSProtection:             "0",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		ReferrerPolicy:            "no-referrer",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
		OriginAgentCluster:        "?1",
		XDNSPrefetchControl:       "off",
		XDownloadOptions:          "noopen",
		XPermittedCrossDomain:     "none",
	}))

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("web/html/index", fiber.Map{
			"Title": "Hello, World!",
		}, "web/html/layouts/layout")
	})

	api := app.Group("/api")
	v1 := api.Group("/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})

	v1.Get("/people/:id", FiberHandler(GetPerson, db))
	v1.Get("/people", FiberHandler(GetPeople, db))
	v1.Delete("/people/:id", FiberHandler(DeletePerson, db))
	v1.Patch("/people/:id", FiberHandler(PatchPerson, db))
	v1.Post("/people", FiberHandler(PostPerson, db))

	v1.Get("/restart", FiberHandler(Restart, db))
	v1.Post("/detect-language", FiberHandler(DetectLanguage, db))
	v1.Get("/scanAccounts/:username", FiberHandler(Restart, db))

	app.Get("/metrics", monitor.New(monitor.Config{Title: "Seekr Metrics Page"}))
	for _, route := range app.GetRoutes(true) {
		fmt.Printf("%s\t-> %s\n", route.Method, route.Path)
	}

	return app.Listen(config.Address())
}

// @Param		request	body		main.MyHandler.request	true	"query params"
// @Success	200		{object}	main.MyHandler.response
// @Router		/test [post]
func FiberHandler(fn func(*fiber.Ctx, *gorm.DB) error, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		localdb := UserDB(c.Locals("username").(string), db)
		return fn(c, localdb)
	}
}

func UserDB(username string, db *gorm.DB) *gorm.DB {
	return db.Where("owner = ?", username)
}