package main

import (
	"github.com/gofiber/fiber/v2"
	"golang-demo/router"
	"log"
)

func main() {
	app := fiber.New()
	router.InitRouter(app)
	log.Fatal(app.Listen(":8080"))
}
