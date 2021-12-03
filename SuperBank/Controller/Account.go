package Controller

import (
	cfg "SuperBank/Config"
	model "SuperBank/Model"
	view "SuperBank/View"
	"SuperBank/database"
	"SuperBank/fabric"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

var c = cfg.GetConfig()

var (
	channelName  string = c.GetString("channelName")
	contractName string = c.GetString("contractName")

	maxWorker      = c.GetInt("maxWorker")
	maxTransaction = c.GetInt("maxTransaction")
	maxAccount     = c.GetInt("maxAccount")

	processing = c.GetInt("processing")
	success    = c.GetInt("success")
	fail       = c.GetInt("fail")
)

var trans chan model.Transaction
var accounts chan []model.Account
var gw *gateway.Gateway

// GetAllAccount get all Account data
func GetAllAccount(w http.ResponseWriter, r *http.Request) {

	var Accounts []model.Account

	error := database.Connector.Find(&Accounts).Error // // Return all Accounts exists in table of database
	if error != nil {
		fmt.Println("Query error !")
		return
	}
	view.Response(w, Accounts)
	// Returns a list of Accounts in JSON format
}

// GetAccountByID get Account with specific ID
func GetAccountByID(w http.ResponseWriter, r *http.Request) {

	var Account model.Account
	vars := mux.Vars(r)
	key := vars["id"] // Get id from URL path

	error := database.Connector.First(&Account, key).Error // Return Account with id = key
	if error != nil {
		fmt.Println("Query error")
		return
	}

	view.Response(w, Account)
	// Returns the found Account in JSON format
}

//CreateAccount creates an Account
func CreateAccount(w http.ResponseWriter, r *http.Request) {

	var db = database.Connector
	var Account model.Account
	w.Header().Set("Content-Type", "application/json")
	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body

	json.Unmarshal(requestBody, &Account) // Convert from JSON format to Account Format

	gw = fabric.InitGateway()

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Println("Failed to get network")
	}

	id := Account.Id
	name := Account.Name
	address := Account.Address
	phonenumber := Account.PhoneNumber
	balance := fmt.Sprintf("%f", Account.Balance)
	status := strconv.Itoa(Account.Status)

	contract := network.GetContract(contractName)

	response, err := contract.SubmitTransaction("CreateAccount", id, name, address, phonenumber, balance, status)
	if err != nil {
		log.Fatalln("Failed to create account in fabric network !", err)
	}

	fmt.Println("Response from Fabric : ", string(response.Responses[0].Response.Payload))

	Account.Createtime = time.Now().String()
	error := db.Create(Account).Error // Create a record in database
	if error != nil {
		log.Fatalln("Failed to create account in mysql !")
	}

	view.Response(w, Account)

	// return the created Account in JSON format
}

// UpdateAccountByID updates Account with respective ID, if ID does not exist, create a new Account
func UpdateAccountByID(w http.ResponseWriter, r *http.Request) {

	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body
	var Account model.Account
	json.Unmarshal(requestBody, &Account) // Convert from JSON format to Account Format

	error := database.Connector.Save(&Account).Error // Update database
	if error != nil {
		fmt.Println("Query Error !")
	}

	fmt.Println("Updated Account successfully !")

	view.Response(w, Account)

	// Return the updated Account in JSON format
}

// DeleteAccountByID delete an Account with specific ID
func DeleteAccountByID(w http.ResponseWriter, r *http.Request) {

	var Account model.Account
	vars := mux.Vars(r) // Get id from URL path
	key := vars["id"]
	if len(vars) == 0 {
		panic("Enter an ID !")
	}

	err := database.Connector.First(&Account, key).Error // Return Account with id = key
	if err != nil {
		fmt.Println("ID doesn't exist")
		return
	}

	database.Connector.Where("id = ?", key).Delete(&Account) // Delete Account with id = key
	fmt.Println("[ID :", key, "] has been successfully deleted !")

	view.NoContent(w)

}

// AccountWithdraw withdraw money from account through a transaction
func AccountWithdraw(w http.ResponseWriter, r *http.Request) {

	var Account model.Account
	var tx model.Transaction
	var db = database.Connector

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if error != nil {
		fmt.Println("Error :", error)
	}

	tx.Trace = uuid.New().String()
	// fmt.Print(tx.Trace)
	now := time.Now()
	tx.Createtime = &now
	tx.Status = processing // processing

	error = db.Create(tx).Error // Create a record in database
	if error != nil {
		log.Fatalln("Failed to create transaction in DB !")
	}

	gw := fabric.InitGateway()
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Println("Failed to get network !")
	}
	contract := network.GetContract(contractName)

	from := tx.To
	amount := fmt.Sprintf("%f", tx.Amount)

	response, err := contract.SubmitTransaction("AccountWithdraw", from, amount)
	if err != nil {
		fmt.Println("Failed to send proposal !")
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", fail, tx.Trace) // status = 2 : Failed
	} else {
		payload := response.Responses[0].ProposalResponse.Response.Payload
		err = json.Unmarshal(payload, &Account)
		if err != nil {
			fmt.Println("err :", err)
		}

		fmt.Println("You have successfully withdrew", tx.Amount, "to your account !")
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Account.Balance, Account.Id)
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", success, tx.Trace) // status = 1 : Done
	}

	view.Response(w, Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountDeposit deposit money into account through a transaction
func AccountDeposit(w http.ResponseWriter, r *http.Request) {
	var tx model.Transaction
	var Account model.Account
	var db = database.Connector

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if error != nil {
		fmt.Println("Error :", error)
	}

	tx.Trace = uuid.New().String()
	// fmt.Print(tx.Trace)
	now := time.Now()
	tx.Createtime = &now
	tx.Status = processing // processing

	error = db.Create(tx).Error // Create a record in database
	if error != nil {
		log.Fatalln("Failed to create transaction in DB !")
	}

	gw := fabric.InitGateway()
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Println("Failed to get network !")
	}
	contract := network.GetContract(contractName)

	from := tx.To
	amount := fmt.Sprintf("%f", tx.Amount)
	response, err := contract.SubmitTransaction("AccountDeposit", from, amount)
	if err != nil {
		fmt.Println("Failed to send proposal !")
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", fail, tx.Trace) // status = 2 : Failed
	} else {
		payload := response.Responses[0].ProposalResponse.Response.Payload
		err = json.Unmarshal(payload, &Account)
		if err != nil {
			fmt.Println("err :", err)
		}

		fmt.Println("You have successfully deposited", tx.Amount, "to your account !")
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Account.Balance, Account.Id)
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", success, tx.Trace) // status = 1 : Done
	}

	view.Response(w, Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountTransfer transfer money between 2 accounts through a transaction
func AccountTransfer(w http.ResponseWriter, r *http.Request) {
	var Accounts []model.Account
	var tx model.Transaction
	var db = database.Connector

	start := time.Now()
	gw = fabric.InitGateway()

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if error != nil {
		fmt.Println("Error :", error)
	}

	db.Raw("SELECT * from accounts WHERE id = ? or id = ?", tx.From, tx.To).Scan(&Accounts)
	fmt.Println("Before :", Accounts)

	tx.Trace = uuid.New().String()
	// fmt.Print(tx.Trace)
	now := time.Now()
	tx.Createtime = &now
	tx.Status = processing // processing

	error = db.Create(tx).Error // Create a record in database
	if error != nil {
		log.Fatalln("Failed to create transaction in DB !")
	}

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Println("Failed to get network from gateway !")
	}

	contract := network.GetContract(contractName)

	from := tx.From
	to := tx.To
	amount := strconv.Itoa(int(tx.Amount))

	response, err := contract.SubmitTransaction("AccountTransfer", from, to, amount)
	if err != nil {
		fmt.Println("Failed to submit transaction ! Can't update the balance !")
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", fail, tx.Trace) // status = 2 : Failed
	} else {
		txid := string(response.TransactionID)
		tx.TxID = txid
		fmt.Println("txid :", txid)
		// fmt.Println(string(response.Responses[0].ProposalResponse.Payload))
		payload := response.Responses[0].ProposalResponse.Response.Payload

		// fmt.Println(string(payload))

		err := json.Unmarshal(payload, &Accounts)
		if err != nil {
			fmt.Println("err :", err)
		}
		fmt.Println("After :", Accounts)

		db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", success, txid, tx.Trace) // status = 1 : Done
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[0].Balance, tx.From)
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[1].Balance, tx.To)
	}

	view.Response(w, tx)

	fmt.Println("Execution time :", time.Since(start))

}

func AccountTransfer_CC(w http.ResponseWriter, r *http.Request) {

	var Accounts []model.Account
	var tx model.Transaction

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if error != nil {
		fmt.Printf("can't unmarshal data ! %+v \n", requestBody)
	}

	go func() {
		fmt.Println("goroutine is processing !")
		trans <- tx
	}()

	Accounts = <-accounts

	fmt.Println("After :", Accounts)

	view.Response(w, Accounts)

	// Returns the new state of 2 Accounts when the transaction is done
}

func InitWorker() {
	trans = make(chan model.Transaction, maxTransaction)
	accounts = make(chan []model.Account, maxAccount)
	for i := 0; i < maxWorker; i++ {
		fmt.Printf("Worker %v is waiting ..... ! \n", i+1)
		go func() {
			Worker(trans, accounts)
		}()
	}
}

func Worker(trans <-chan model.Transaction, accounts chan<- []model.Account) {
	var Accounts []model.Account
	var db = database.Connector

	gw := fabric.InitGateway()
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Println("Failed to get network !")
	}

	contract := network.GetContract(contractName)

	fmt.Println("Worker is working !")
	for tx := range trans {
		fmt.Println("Still processing _________")
		if tx.From == tx.To {
			log.Panicf("Please enter correct recipient ID ! %v and %v", tx.From, tx.To)
		}

		db.Raw("SELECT * from accounts WHERE id = ? or id = ?", tx.From, tx.To).Scan(&Accounts)
		fmt.Println("Before :", Accounts)

		tx.Trace = uuid.New().String()
		// fmt.Print(tx.Trace)
		now := time.Now()
		tx.Createtime = &now
		tx.Status = processing // processing

		error := db.Create(tx).Error // Create a record in database
		if error != nil {
			log.Fatalln("Failed to create transaction in DB !")
		}

		from := tx.From
		to := tx.To
		amount := fmt.Sprintf("%f", tx.Amount)

		response, err := contract.SubmitTransaction("AccountTransfer", from, to, amount)
		if err != nil {
			fmt.Println("Failed to submit transaction ! Can't update the balance !")
			db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", fail, tx.Trace) // status = 2 : Failed
		} else {
			txid := string(response.TransactionID)
			tx.TxID = txid
			fmt.Println("txid :", txid)

			payload := response.Responses[0].ProposalResponse.Response.Payload

			// fmt.Println(string(payload))

			err := json.Unmarshal(payload, &Accounts)
			if err != nil {
				fmt.Println("err :", err)
			}
			// fmt.Println(Accounts)

			db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", success, txid, tx.Trace) // status = 1 : Done
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[0].Balance, tx.From)
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[1].Balance, tx.To)
		}
		// time.Sleep(time.Second)
		accounts <- Accounts
	}
}
