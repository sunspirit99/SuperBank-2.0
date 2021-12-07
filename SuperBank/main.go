package main

import (
	db "SuperBank/database"

	r "SuperBank/Router"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	log.Println("Starting the HTTP server on port 8000")

	db.Init()
	router := mux.NewRouter().StrictSlash(true)
	r.InitaliseHandlers(router)

	// controller.InitWorker()

	log.Fatal(http.ListenAndServe(":8000", router))

}
