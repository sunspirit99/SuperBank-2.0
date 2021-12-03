package database

import (
	cfg "SuperBank/Config"
	model "SuperBank/Model"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connector variable
var Connector *gorm.DB

// init function will be called when the package is imported
func init() {

	c := cfg.GetConfig()
	servername := c.GetString("ServerName")
	user := c.GetString("User")
	password := c.GetString("Password")
	db := c.GetString("DB")

	config :=
		model.DBConfig{
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
	Migrate(&model.Account{}, &model.Transaction{})

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

// Migrate create/updates database table
func Migrate(accountTable *model.Account, transactionTable *model.Transaction) {
	Connector.AutoMigrate(&accountTable, &transactionTable)
	log.Println("Tables migrated")
}

var GetConnectionString = func(config model.DBConfig) string {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true", config.User, config.Password, config.ServerName, config.DB)
	return connectionString
}
