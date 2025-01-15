package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

func main() {
	InitDB()

	e := echo.New()

	SetupRoutes(e)

	fmt.Println("Server is running on http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}