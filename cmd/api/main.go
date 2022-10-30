package main

import (
	"log"
	"os"

	_ "github.com/antonioalfa22/go-rest-template/docs"
	"github.com/antonioalfa22/go-rest-template/internal/api"
	"github.com/joho/godotenv"
)

// @Golang API REST
// @version 1.0
// @description API REST in Golang with Gin Framework

// @contact.name Antonio Paya Gonzalez
// @contact.url http://antoniopg.tk
// @contact.email antonioalfa22@gmail.com

// @license.name MIT
// @license.url https://github.com/antonioalfa22/go-rest-template/blob/master/LICENSE

// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {

	if os.Getenv("ENV") == "DEV" {
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatalf("Error loading .env file")
		}

	}

	api.Run("")
}
