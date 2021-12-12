package middlewares

import (
	"strconv"

	"github.com/labstack/echo"
)

// AuthenticationExtractor is a middleware to extract value from X-User-Id header and set into the Echo context
func AuthenticationExtractor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := strconv.Atoi(c.Request().Header.Get("X-User-Id"))
		c.Set("user_id", userID)
		return next(c)
	}
}
