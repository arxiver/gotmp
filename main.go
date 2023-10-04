package main

import (
	"fmt"
	"log"
	"os"
	"server/Auth"
	"server/DB"
	"server/Middleware"
	"server/Routes"
	"server/Scheduler"
	"server/Utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	if !DB.InitCollections() {
		fmt.Println("Failed to initialize collections in the database")
		return
	}
	app := fiber.New()
	app.Use(cors.New())
	if _, err := os.Stat(Utils.BaseStaticPath); os.IsNotExist(err) { // make sure base path exists
		os.MkdirAll(Utils.BaseStaticPath, 0755)
	}
	Scheduler.Init()
	app.Static("/", Utils.BaseStaticPath)
	app.Use(Middleware.AppGaurd)
	SetupRoutes(app)
	app.Use(recover.New())
	app.Use(logger.New())
	log.Fatal(app.Listen(":3333"))
}

func SetupRoutes(app *fiber.App) {
	Auth.SeedAdmin()
	Routes.UserRoute(app.Group("/user"))
}
