package Router

import (
	view "SuperBank/View"

	"github.com/gorilla/mux"
)

func InitaliseHandlers(router *mux.Router) {

	router.HandleFunc("/create", view.CreateAccount).Methods("POST")
	router.HandleFunc("/transfer", view.AccountTransfer).Methods("PUT")
	router.HandleFunc("/get", view.GetAllAccount).Methods("GET")
	router.HandleFunc("/get/{id}", view.GetAccountByID).Methods("GET")
	router.HandleFunc("/update", view.UpdateAccountByID).Methods("PUT")
	router.HandleFunc("/delete/{id}", view.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/delete", view.DeleteAccountByID).Methods("DELETE")
	router.HandleFunc("/withdraw", view.AccountWithdraw).Methods("PUT")
	router.HandleFunc("/deposit", view.AccountDeposit).Methods("PUT")
	router.HandleFunc("/transfer", view.AccountTransfer).Methods("PUT")
	router.HandleFunc("/transfers_CC", view.AccountTransfer_CC).Methods("PUT")

}
