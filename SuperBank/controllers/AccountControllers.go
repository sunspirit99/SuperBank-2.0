package controllers

import (
	"SuperBank/database"
	"SuperBank/entity"
	"SuperBank/fabric"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const minBalance float32 = 1 // Minimum balance of an Account
const minCost float32 = 1    // Minimum amount to trade in a transaction
const batchsize int = 1000   // the number of records in a single write to the database
// const loop int = 10

var channelName string = "mychannel"
var contractName string = "basic"

var trans chan entity.Transaction
var accounts chan []entity.Account
var gw *gateway.Gateway

// GetAllAccount get all Account data
func GetAllAccount(w http.ResponseWriter, r *http.Request) {

	var Accounts []entity.Account

	error := database.Connector.Find(&Accounts).Error // // Return all Accounts exists in table of database
	if error != nil {
		fmt.Println("Query error !")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Accounts)
	// Returns a list of Accounts in JSON format
}

// GetAccountByID get Account with specific ID
func GetAccountByID(w http.ResponseWriter, r *http.Request) {

	var Account entity.Account
	vars := mux.Vars(r)
	key := vars["id"] // Get id from URL path

	error := database.Connector.First(&Account, key).Error // Return Account with id = key
	if error != nil {
		fmt.Println("Query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)
	// Returns the found Account in JSON format
}

//CreateAccount creates an Account
func CreateAccount(w http.ResponseWriter, r *http.Request) {

	var db = database.Connector
	var Account entity.Account
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Account)
	w.WriteHeader(http.StatusCreated)

	// return the created Account in JSON format
}

// CreateAccountFromCSV creates a list of Account in CSV file
func CreateAccountFromCSV(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Processing !!!")
	start1 := time.Now()
	var Accounts = LoadAccountsCSV()
	end1 := time.Since(start1) // Total time to read data from CSV file

	start2 := time.Now()

	error := database.Connector.Statement.CreateInBatches(Accounts, batchsize).Error
	// Load multiple records (const batchsize = 1000) into 1 batch and then write to database

	if error != nil {
		fmt.Println("Query Error !")
	}

	end2 := time.Since(start2) // Total time to insert data to Database
	fmt.Printf("\n Created 10.000 accounts complete at %v", time.Now())
	fmt.Printf("\n Time to read data from CSV file is : %v \n Time to write to DB is : %v \n", end1, end2)
	w.WriteHeader(http.StatusOK)

}

// UpdateAccountByID updates Account with respective ID, if ID does not exist, create a new Account
func UpdateAccountByID(w http.ResponseWriter, r *http.Request) {

	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body
	var Account entity.Account
	json.Unmarshal(requestBody, &Account) // Convert from JSON format to Account Format

	error := database.Connector.Save(&Account).Error // Update database
	if error != nil {
		fmt.Println("Query Error !")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)

	fmt.Println("Updated Account successfully !")
	// Return the updated Account in JSON format
}

// DeleteAccountByID delete an Account with specific ID
func DeleteAccountByID(w http.ResponseWriter, r *http.Request) {

	var Account entity.Account
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
	w.WriteHeader(http.StatusNoContent)

}

// AccountWithdraw withdraw money from account through a transaction
func AccountWithdraw(w http.ResponseWriter, r *http.Request) {

	var Account entity.Account
	var tx entity.Transaction
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
	tx.Status = 0 // processing

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
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 2, tx.Trace) // status = 2 : Failed
	} else {
		payload := response.Responses[0].ProposalResponse.Response.Payload
		err = json.Unmarshal(payload, &Account)
		if err != nil {
			fmt.Println("err :", err)
		}

		fmt.Println("You have successfully withdrew", tx.Amount, "to your account !")
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Account.Balance, Account.Id)
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 1, tx.Trace) // status = 1 : Done
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountDeposit deposit money into account through a transaction
func AccountDeposit(w http.ResponseWriter, r *http.Request) {
	var tx entity.Transaction
	var Account entity.Account
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
	tx.Status = 0 // processing

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
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 2, tx.Trace) // status = 2 : Failed
	} else {
		payload := response.Responses[0].ProposalResponse.Response.Payload
		err = json.Unmarshal(payload, &Account)
		if err != nil {
			fmt.Println("err :", err)
		}

		fmt.Println("You have successfully deposited", tx.Amount, "to your account !")
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Account.Balance, Account.Id)
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 1, tx.Trace) // status = 1 : Done
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountTransfer transfer money between 2 accounts through a transaction
func AccountTransfer(w http.ResponseWriter, r *http.Request) {
	var Accounts []entity.Account
	var tx entity.Transaction
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
	tx.Status = 0 // processing

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
		db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 2, tx.Trace) // status = 2 : Failed
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

		db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", 1, txid, tx.Trace) // status = 1 : Done
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[0].Balance, tx.From)
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[1].Balance, tx.To)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)

	fmt.Println("Execution time :", time.Since(start))

}

func AccountTransfer_CC(w http.ResponseWriter, r *http.Request) {

	var Accounts []entity.Account
	var tx entity.Transaction

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
	// fmt.Println(Accounts)

	fmt.Printf("Post-Balance : [id=%v] = %v , [id=%v] = %v \n", tx.From, Accounts[0].Balance, tx.To, Accounts[1].Balance)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)

	// Returns the new state of 2 Accounts when the transaction is done
}

func Worker(trans <-chan entity.Transaction, accounts chan<- []entity.Account) {
	var Accounts []entity.Account
	var db = database.Connector

	fmt.Println("Worker is working !")
	for tx := range trans {
		if tx.From == tx.To {
			log.Panicf("Please enter correct recipient ID ! %v and %v", tx.From, tx.To)
		}

		gw := fabric.InitGateway()
		network, err := gw.GetNetwork(channelName)
		if err != nil {
			fmt.Println("Failed to get network !")
		}

		contract := network.GetContract(contractName)

		from := tx.From
		to := tx.To
		amount := fmt.Sprintf("%f", tx.Amount)
		response, err := contract.SubmitTransaction("AccountTransfer", from, to, amount)
		if err != nil {
			fmt.Println("Failed to submit transaction ! Can't update the balance !")
			db.Exec("UPDATE transactions SET status = ? WHERE trace = ?", 2, tx.Trace) // status = 2 : Failed
		} else {
			txid := string(response.TransactionID)
			// fmt.Println(len(response.Responses))
			// fmt.Println(string(response.Responses[0].ProposalResponse.Payload))
			payload := response.Responses[0].ProposalResponse.Response.Payload

			// fmt.Println(string(payload))

			err := json.Unmarshal(payload, &Accounts)
			if err != nil {
				fmt.Println("err :", err)
			}
			fmt.Println(Accounts)

			db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", 1, txid, tx.Trace) // status = 1 : Done
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[0].Balance, tx.From)
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", Accounts[1].Balance, tx.To)
		}

		// time.Sleep(time.Second)
		accounts <- Accounts

	}
}

func AccountTransferFromCSV(w http.ResponseWriter, r *http.Request) {

	trans := LoadTransactionsCSV() // LoadTransactionsCSV return a list of Account from CSV file

	for _, tran := range trans {
		var Accounts []entity.Account
		var Account1 entity.Account
		var Account2 entity.Account

		if tran.From == tran.To {
			panic("Please enter correct recipient ID !")
		}

		err1 := database.Connector.Where(`id =? `, tran.From).Select("balance").First(&Account1).Error
		err2 := database.Connector.Where(`id =? `, tran.To).Select("balance").First(&Account2).Error
		if err1 != nil || err2 != nil {
			fmt.Println("Please enter correct information !")
			continue
		}

		Accounts = append(Accounts, Account1, Account2) // Contains 2 Accounts participating in a transaction

		if Accounts[0].Balance < minBalance {
			fmt.Println("You dont have enough money to transfer !")
			return
		}
		if Accounts[0].Balance-tran.Amount < minBalance {
			fmt.Println("The maximum amount that can be transferred is", Accounts[0].Balance-minBalance, "!")
			return
		}
		if tran.Amount < minCost {
			fmt.Println("The minimum amount that can be transferred is", minCost, "!")
			return
		}

		Withdraw(&Accounts[0], tran.Amount)
		Deposit(&Accounts[1], tran.Amount) // Change the balance of these 2 Accounts
		fmt.Println("you [ID :", tran.From, "] have successfully transferred", tran.Amount, "to [ID :", tran.To, "] !")

		database.Connector.Save(&Accounts[0])
		database.Connector.Save(&Accounts[1]) // Update database

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Accounts)
	}
	w.WriteHeader(http.StatusOK)
	// Returns the new state of Accounts when transactions are processed
}

// Update balance when there's a qualified withdrawal request
func Withdraw(Account *entity.Account, num float32) {
	Account.Balance = Account.Balance - num
}

// Update balance when there's a qualified deposit request
func Deposit(Account *entity.Account, num float32) {
	Account.Balance = Account.Balance + num
	// fmt.Println(num)
}

// LoadAccountsCSV reads Accounts from CSV file
func LoadAccountsCSV() []entity.Account {

	var Accounts []entity.Account
	file, _ := os.Open("10kAccount.csv")
	reader := csv.NewReader(bufio.NewReader(file))

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		balance, err := strconv.ParseFloat(line[4], 32)
		status, err := strconv.ParseInt(line[5], 0, 64)
		// t := time.Now()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		Accounts = append(Accounts, entity.Account{
			Id:          line[0],
			Name:        line[1],
			Address:     line[2],
			PhoneNumber: line[3],
			Balance:     float32(balance),
			Status:      int(status),
			Createtime:  "",
		})
	}
	return Accounts
	// Return a list of Accounts
}

// LoadTransactionsCSV reads transactions from CSV file
func LoadTransactionsCSV() []entity.Transaction {
	var trans []entity.Transaction
	file, _ := os.Open("transactions.csv")
	reader := csv.NewReader(bufio.NewReader(file))

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// from, err := strconv.ParseInt(line[0], 0, 64)
		// amount, err := strconv.ParseFloat(line[2], 64)
		// to, err := strconv.ParseInt(line[3], 0, 64)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		trans = append(trans, entity.Transaction{
			From:   line[0],
			Amount: 1000000,
			To:     line[3],
		})
	}
	return trans
	// return a list of transactions
}
