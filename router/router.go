package router

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func InitRouter(router *fiber.App) {
	log.Println("init StabilityAiRouter ...")
	StabilityAiRouter(router)
}
