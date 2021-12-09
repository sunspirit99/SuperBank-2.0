package Model

import (
	cfg "SuperBank/Config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connector variable
var Connector *gorm.DB

// init function will be called when the package is imported
func Init() {

	c := cfg.GetConfig()
	servername := c.GetString("ServerName")
	user := c.GetString("User")
	password := c.GetString("Password")
	db := c.GetString("DB")

	config :=
		DBConfig{
			ServerName: servername,
			User:       user,
			Password:   password,
			DB:         db,
		}

	connectionString := GetConnectionString(config)
	err := Connect(connectionString)
	if err != nil {
		panic(err.Error())
	}
	Connector.AutoMigrate(
		&Account{},
		&Transaction{},
	)
	log.Println("Tables migrated")

}

// Connect creates MySQL connection
func Connect(connectionString string) error {
	var err error
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

var GetConnectionString = func(config DBConfig) string {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true", config.User, config.Password, config.ServerName, config.DB)
	return connectionString
}
