package db

import (
	"fmt"
	"os"
	"time"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/config"
	"github.com/antonioalfa22/go-rest-template/internal/pkg/models/tokens"
	"github.com/antonioalfa22/go-rest-template/internal/pkg/models/users"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	DB  *gorm.DB
	err error
)

type Database struct {
	*gorm.DB
}

// SetupDB opens a database and saves the reference to `Database` struct.
func SetupDB() {
	var db = DB

	configuration := config.GetConfig()

	database := configuration.Database.Dbname
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	db, err = gorm.Open("postgres", "host="+host+" port="+port+" user="+username+" dbname="+database+"  sslmode=disable password="+password)
	if err != nil {
		fmt.Println("db err: ", err)
	}

	// Change this to true if you want to see SQL queries
	db.LogMode(true)
	db.DB().SetMaxIdleConns(configuration.Database.MaxIdleConns)
	db.DB().SetMaxOpenConns(configuration.Database.MaxOpenConns)
	db.DB().SetConnMaxLifetime(time.Duration(configuration.Database.MaxLifetime) * time.Second)
	DB = db
	migration()
}

// Auto migrate project models
func migration() {
	if (!DB.HasTable(&users.User{})) {
		DB.CreateTable(&users.User{})
	}
	if (!DB.HasTable(&tokens.Token{})) {
		DB.CreateTable(&tokens.Token{})
	}
	DB.Model(&users.User{}).DropColumn("hash")

	DB.Model(&users.User{}).Related(&tokens.Token{})
	DB.AutoMigrate(&tokens.Token{})
	DB.AutoMigrate(&users.User{})
}

func GetDB() *gorm.DB {
	return DB
}
