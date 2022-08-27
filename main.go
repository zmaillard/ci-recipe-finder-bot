package main

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	keyauth "github.com/iwpnd/fiber-key-auth"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

func main() {
	config.Init()

	engine := html.New("./static", ".html")

	app := fiber.New(fiber.Config{Views: engine})
	api := app.Group("/api", keyauth.New())

	api.Post("/receivesms", handlers.ReceiveSMSHandler)
	api.Get("/help", handlers.HelpHandler)

	app.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(200)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		app.Shutdown()
	}()

	log.Fatal(app.Listen(":3000"))
}
