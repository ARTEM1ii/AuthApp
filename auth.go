package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
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

func Login(c echo.Context) error {
	loginRequest := new(LoginRequest)
	if err := c.Bind(loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	
	var user User
	if err := DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	token, err := GenereteJWT(user.ID)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return c.String(http.StatusInternalServerError, "Error generating token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login succesful",
		"token": token,
	})
}

func Register(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {			//Вписываем в поля User имя с http 
		fmt.Println("Invalid input")
		return c.String(http.StatusBadRequest, "Invalid input")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error hashing password")
	}
	user.Password = string(hashedPassword)

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

func News(c echo.Context) error {
	// Привязываем данные title и description
	task := new(Task)
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid input data",
		})
	}

	// Получение файла из запроса
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "File is required",
		})
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	// Создаем папку для загрузки, если она не существует
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, 0755); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create upload directory",
			})
		}
	}

	// Сохраняем файл
	filePath := uploadDir + "/" + file.Filename
	dst, err := os.Create(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save file",
		})
	}
	defer dst.Close()

	// Копируем содержимое файла
	if _, err := dst.ReadFrom(src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to write file",
		})
	}

	// Добавляем путь к картинке в задачу
	task.Picture = "/uploads/" + file.Filename

	// Сохраняем данные в базу
	if err := DB.Create(&task).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save task to the database",
		})
	}

	// Возвращаем успешный ответ
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "News created successfully",
		"title":       task.Title,
		"description": task.Description,
		"picture_url": task.Picture,
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