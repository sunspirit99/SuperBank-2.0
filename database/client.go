package database

import (
	"log"
	"rest-go-demo/entity"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Required for MySQL dialect
)

//Connector variable used for CRUD operation's
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

//Connect creates MySQL connection
func Connect(connectionString string) error {
	var err error
	Connector, err = gorm.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	log.Println("Connection was successful!!")
	return nil
}

//Migrate create/updates database table
func Migrate(table *entity.User) {
	Connector.AutoMigrate(&table)
	log.Println("Table migrated")
}
