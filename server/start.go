package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/switchboard-org/switchboard/parsecfg"
	"io"
)

func StartServer(parser parsecfg.Parser) error {
	app := fiber.New()
	adminGroup := app.Group("/admin")
	adminGroup.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": "password",
		},
	}))
	apiGroup := app.Group("/api")

	adminGroup.Get("/")
	apiGroup.Put("/upload_blob", func(c *fiber.Ctx) error {
		parser.Parse()
		file, err := c.FormFile("config")
		if err != nil {
			return err
		}
		openFile, err := file.Open()
		if err != nil {
			return err
		}
		io.ReadAll(openFile)
		return nil
	})

	return nil
	//TODO: Register admin endpoints (deploy, log stream, trigger list, workflow list, deployed sha)

	//TODO: Init config if present
	//TODO: parse config if present

	//TODO: register all plugins and add to shared map

	//TODO: register all workflow processors

	//TODO: register all triggers (webhooks) - pass list of workflows that rely on them
}
