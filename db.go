package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "postgresql://globaldb_4p6p_user:UKWQNnGrtmiOTCK7WJnEXTYN4h2uGMfP@dpg-cu3vuni3esus73c2lhlg-a.oregon-postgres.render.com/globaldb_4p6p"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	fmt.Println("Database connection successful!")


	DB.AutoMigrate(&User{}, &Task{})
	fmt.Println("User table created or updated!")
}

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

type Task struct {
	Title       string `json:"title" gorm:"primaryKey"`
	Description string `json:"description"`
	Picture     string `json:"picture"`
}

type LoginRequest struct{
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}