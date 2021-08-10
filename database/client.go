package database

import (
	"log"
	"rest-go-demo/entity"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// "github.com/jinzhu/gorm"
	// "github.com/jinzhu/gorm/dialects/mysql" //Required for MySQL dialect
)

// Connector variable
var Connector *gorm.DB

// init function will be called when the package is imported
func init() {
	config :=
		Config{
			ServerName: "127.0.0.1:3306",
			User:       "root",
			Password:   "Sunspirit9.9",
			DB:         "Bank",
		}

	connectionString := GetConnectionString(config)
	err := Connect(connectionString)
	if err != nil {
		panic(err.Error())
	}
	Migrate(&entity.User{})
}

// Connect creates MySQL connection
func Connect(connectionString string) error {
	var err error
	dsn := "root:Sunspirit9.9@tcp(127.0.0.1:3306)/Bank?charset=utf8mb4&parseTime=True&loc=Local"
	Connector, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	log.Println("Connection was successful!!")
	return nil
}

// Migrate create/updates database table
func Migrate(table *entity.User) {
	Connector.AutoMigrate(&table)
	log.Println("Table migrated")
}
