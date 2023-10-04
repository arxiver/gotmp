package Routes

import (
	"server/Auth"
	"server/Controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRoute(route fiber.Router) {
	route.Post("/new", Controllers.UserNew)
	route.Post("/get_all", Controllers.UserGetAll)
	route.Post("/modify/:id", Controllers.UserModifyById)
	route.Get("/:id", Controllers.UserGetById)
	route.Post("/login", Auth.Login)
	route.Post("/logout", Auth.Logout)
}
