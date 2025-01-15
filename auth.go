package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your-secret-key")

func GenereteJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":	   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}


func Register(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {			//Вписываем в поля User имя с http 
		fmt.Println("Invalid input")
		return c.String(http.StatusBadRequest, "Invalid input")
	}
	
	var existingUser User
	res := DB.Where("email = ?", user.Email).First(&existingUser)
	if res.Error == nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "User with this email already exists",
		})
	}

	result := DB.Create(&user)						//Сохраняем в базу данных
	if result.Error != nil {
		fmt.Println("Error saving user:", result.Error)
		return c.String(http.StatusInternalServerError, "Error saving user")
	}

	token, err := GenereteJWT(user.ID)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return c.String(http.StatusInternalServerError, "Error generating token")
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "User registered successfully",
		"token":   token,
	})
}


func VerifyJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Missing or invalid token",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
		}
		return next(c)
	}
}