package main

import (
	"log"
	"net/http"
	"rest-go-demo/controllers"
	"rest-go-demo/database"
	"rest-go-demo/entity"

	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Required for MySQL dialect
)

func main() {
	initDB()
	log.Println("Starting the HTTP server on port 8000")

	router := mux.NewRouter().StrictSlash(true)
	initaliseHandlers(router)
	log.Fatal(http.ListenAndServe(":8000", router))
}

func initaliseHandlers(router *mux.Router) {
	router.HandleFunc("/create", controllers.CreateUser).Methods("POST")
	router.HandleFunc("/get", controllers.GetAllUser).Methods("GET")
	router.HandleFunc("/get/{id}", controllers.GetUserByID).Methods("GET")
	router.HandleFunc("/update/{id}", controllers.UpdateUserByID).Methods("PUT")
	router.HandleFunc("/delete/{id}", controllers.DeleteUserByID).Methods("DELETE")
	router.HandleFunc("/delete", controllers.DeleteUserByID).Methods("DELETE")
	router.HandleFunc("/withdraw/{id}", controllers.UserWithdraw).Methods("PUT")
	router.HandleFunc("/deposit/{id}", controllers.UserDeposit).Methods("PUT")
	router.HandleFunc("/transfer", controllers.UserTransfer).Methods("PUT")
}

func initDB() {
	config :=
		database.Config{
			ServerName: "127.0.0.1:3306",
			User:       "root",
			Password:   "",
			DB:         "data1",
		}

	connectionString := database.GetConnectionString(config)
	err := database.Connect(connectionString)
	if err != nil {
		panic(err.Error())
	}
	database.Migrate(&entity.User{})
}
