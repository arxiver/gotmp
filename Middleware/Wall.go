package Middleware

import (
	"fmt"
	"regexp"
	"server/Auth"
	"server/Models"

	"github.com/gofiber/fiber/v2"
)

var publicRoutes = map[string]bool{
	"/user/login": true,
	"/":           true,
}
var userRoutes = map[string]bool{
	"/user/states": true,
}
var userRoutesRE = regexp.MustCompile(`/log/wa/*|/log/ss/*|/user/st/*`)

func AppGaurd(c *fiber.Ctx) error {
	fmt.Println(c.OriginalURL())
	userAPIKey := string(c.Request().Header.Peek(Auth.TOKEN_FIELD))
	// Based on the user role, check if the route is protected
	if publicRoutes[c.OriginalURL()] {
		return c.Next()
	}
	ok, session := Auth.IsSessionExist(userAPIKey)
	if !ok {
		return c.Status(401).SendString("Unauthorized")
	}
	Auth.UpdateSessionLastReqTime(session)
	switch session.GetRole() {
	case Models.ROLE_ADMIN:
		break
	case Models.ROLE_USER:
		if !userRoutes[c.OriginalURL()] && !userRoutesRE.MatchString(c.OriginalURL()) {
			return c.Status(403).SendString("Forbidden")
		}
		break
	default:
		return c.Status(403).SendString("Forbidden")
	}
	return c.Next()
}
