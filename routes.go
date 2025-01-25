package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	e.POST("/register", Register)
	e.GET("/protected", ProtectedRoute, VerifyJWT)
	e.POST("/login", Login)
	e.POST("/news", News)
}


func ProtectedRoute(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "You have accessed a protected route!",
	})
}