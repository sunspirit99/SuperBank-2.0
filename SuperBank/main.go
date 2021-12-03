package main

import (
	controller "SuperBank/Controller"

	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	log.Println("Starting the HTTP server on port 8000")

	router := mux.NewRouter().StrictSlash(true)
	initaliseHandlers(router)

	controller.InitWorker()

	log.Fatal(http.ListenAndServe(":8000", router))

}
func initaliseHandlers(router *mux.Router) {

	router.HandleFunc("/create", controller.CreateAccount).Methods("POST")
	router.HandleFunc("/transfer", controller.AccountTransfer).Methods("PUT")
	router.HandleFunc("/get", controller.GetAllAccount).Methods("GET")
	router.HandleFunc("/get/{id}", controller.GetAccountByID).Methods("GET")
	router.HandleFunc("/update", controller.UpdateAccountByID).Methods("PUT")
	router.HandleFunc("/delete/{id}", controller.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/delete", controller.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/withdraw", controller.AccountWithdraw).Methods("PUT")
	router.HandleFunc("/deposit", controller.AccountDeposit).Methods("PUT")
	router.HandleFunc("/transfer", controller.AccountTransfer).Methods("PUT")
	router.HandleFunc("/transfers_CC", controller.AccountTransfer_CC).Methods("PUT")

}
