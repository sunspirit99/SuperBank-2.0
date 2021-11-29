package database

import (
	"SuperBank/entity"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connector variable
var Connector *gorm.DB

// init function will be called when the package is imported
func init() {
	config :=
		Config{
			ServerName: "localhost:3306",
			User:       "root",
			Password:   "Password@99",
			DB:         "SuperBank",
		}

	connectionString := GetConnectionString(config)
	err := Connect(connectionString)
	if err != nil {
		panic(err.Error())
	}
	Migrate(&entity.Account{})
}

// Connect creates MySQL connection
func Connect(connectionString string) error {
	var err error
	// dsn := "root:Sunspirit9.9@tcp(127.0.0.1:3306)/Bank?charset=utf8mb4&parseTime=True&loc=Local"
	Connector, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{
		// SkipDefaultTransaction: true,
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}

	log.Println("Connection was successful!!")
	return nil
}

// Migrate create/updates database table
func Migrate(table *entity.Account) {
	Connector.AutoMigrate(&table)
	log.Println("Table migrated")
}
