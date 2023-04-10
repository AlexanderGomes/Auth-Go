package routes

import (
	"auth-go/backend/controllers"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/info", controllers.GetUser)
	app.Post("api/logout", controllers.Logout)
}
