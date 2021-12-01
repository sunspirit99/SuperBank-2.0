package main

import (
	"SuperBank/controllers"
	"SuperBank/entity"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const maxWorker = 10
const maxTransaction = 1000
const maxAccount = 1000

var trans chan entity.Transaction
var accounts chan []entity.Account
var channelName string = "mychannel"
var contractName string = "basic"

func main() {

	log.Println("Starting the HTTP server on port 8000")

	router := mux.NewRouter().StrictSlash(true)
	initaliseHandlers(router)

	trans = make(chan entity.Transaction, maxTransaction)
	accounts = make(chan []entity.Account, maxAccount)

	for i := 0; i < maxWorker; i++ {
		fmt.Printf("Worker %v is waiting ..... ! \n", i+1)
		go func() {
			controllers.Worker(trans, accounts)
		}()
	}

	log.Fatal(http.ListenAndServe(":8000", router))

}
func initaliseHandlers(router *mux.Router) {
	router.HandleFunc("/create", controllers.CreateAccount).Methods("POST")
	router.HandleFunc("/creates", controllers.CreateAccountFromCSV).Methods("POST")
	router.HandleFunc("/transfer", controllers.AccountTransfer).Methods("PUT")
	router.HandleFunc("/get", controllers.GetAllAccount).Methods("GET")
	router.HandleFunc("/get/{id}", controllers.GetAccountByID).Methods("GET")
	router.HandleFunc("/update", controllers.UpdateAccountByID).Methods("PUT")
	router.HandleFunc("/delete/{id}", controllers.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/delete", controllers.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/withdraw", controllers.AccountWithdraw).Methods("PUT")
	router.HandleFunc("/deposit", controllers.AccountDeposit).Methods("PUT")
	router.HandleFunc("/transfer", controllers.AccountTransfer).Methods("PUT")
	router.HandleFunc("/transfers", controllers.AccountTransferFromCSV).Methods("PUT")
	router.HandleFunc("/transfers_CC", controllers.AccountTransfer_CC).Methods("PUT")

}
