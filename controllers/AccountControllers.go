package controllers

import (
	"SuperBank/database"
	"SuperBank/entity"
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

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const minBalance float32 = 1 // Minimum balance of an Account
const minCost float32 = 1    // Minimum amount to trade in a transaction
const batchsize int = 1000   // the number of records in a single write to the database
const loop int = 10

var trans chan entity.Transaction
var accounts chan []entity.Account

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

	w.Header().Set("Content-Type", "application/json")
	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body

	var Account entity.Account
	json.Unmarshal(requestBody, &Account) // Convert from JSON format to Account Format

	error := database.Connector.Create(Account).Error // Create a record in database
	if error != nil {
		fmt.Println("Failed to create account !")
	} else {

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Account)
		w.WriteHeader(http.StatusCreated)
	}

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

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		fmt.Println("Unreadable !!!")
	}

	err1 := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if err1 != nil {
		fmt.Println("Error")
	}

	sess := database.Connector.Session((&gorm.Session{PrepareStmt: true}))

	error := sess.First(&Account, tx.From).Error
	if error != nil {
		panic("Query error !")
	}

	if Account.Balance < minBalance {
		// Current balance is less than minimum balance
		fmt.Println("You dont have enough money to withdraw !")
		return
	}
	if Account.Balance-tx.Amount < minBalance {
		// Make sure the balance after the transaction is not less than the minimum balance
		fmt.Println("The maximum amount that can be withdrawn is", Account.Balance-minBalance, "!")
		return
	}
	if tx.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (1000)
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	}

	Withdraw(&Account, tx.Amount) // Change the balance of this Account
	fmt.Println("You have successfully withdrew", tx.Amount, "from your account !")

	sess.Model(&Account).Update("balance", Account.Balance) // Update database
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountDeposit deposit money into account through a transaction
func AccountDeposit(w http.ResponseWriter, r *http.Request) {
	var tx entity.Transaction
	var Account entity.Account

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		fmt.Println("Unreadble ")
	}

	err1 := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if err1 != nil {
		fmt.Println("Error")
	}
	// fmt.Println(tx)

	sess := database.Connector.Session((&gorm.Session{PrepareStmt: true}))

	error := sess.First(&Account, tx.From).Error
	// Get Account information to deposit money

	if error != nil {
		panic("Query error !!!")
	}
	if tx.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (5000)
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	}

	// Deposit(&Account, tx.Amount) // Change the balance of this Account
	// fmt.Println(Account)
	fmt.Println("You have successfully deposited", tx.Amount, "to your account !")

	// database.Connector.Save(&Account) // Update database
	// sess.Model(&Account).Update("balance", Account.Balance)
	sess.Exec("UPDATE accounts SET balance = ? WHERE id = ?", gorm.Expr("balance + ?", tx.Amount), tx.From)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Account)
	// Returns the new state of the Account when the transaction is done
}

// AccountTransfer transfer money between 2 accounts through a transaction
func AccountTransfer(w http.ResponseWriter, r *http.Request) {
	var Accounts []entity.Account
	// var Account1 entity.Account
	// var Account2 entity.Account
	var tx entity.Transaction
	var db = database.Connector

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &tx) // Convert from JSON format to Account Format
	if error != nil {
		fmt.Println("Error !")
	}

	if tx.From == tx.To {
		panic("Please enter correct recipient ID !")
	}

	// sess := db.Session((&gorm.Session{PrepareStmt: true}))

	// err = sess.Where("id =? or id =?", tx.From, tx.To).Find(&Accounts).Error
	err = db.Raw("SELECT balance FROM accounts where id = ? or id = ?", tx.From, tx.To).Limit(2).Scan(&Accounts).Error
	// err2 := sess.First(&Account2, tx.To).Error
	fmt.Println("Balance of Founded Accounts :", Accounts[0].Balance, Accounts[1].Balance)

	if err != nil {
		panic("Cannot find the record by id !")
	}

	if Accounts[0].Balance < minBalance {
		// Current balance is less than minimum balance
		fmt.Println("You dont have enough money to transfer !")
		return
	}
	if Accounts[0].Balance-tx.Amount < minBalance {
		// Make sure the balance after the transaction is not less than the minimum balance
		fmt.Println("The maximum amount that can be transferred is", Accounts[0].Balance-minBalance, "!")
		return
	}
	if tx.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (1000)
		fmt.Println("The minimum amount that can be transferred is", minCost, "!")
		return
	}

	fmt.Printf("Pre-Balance : [id=%v] = %v , [id=%v] = %v \n", tx.From, Accounts[0].Balance, tx.To, Accounts[1].Balance)
	// Accounts[0].Balance -= tx.Amount
	// Accounts[1].Balance += tx.Amount
	// Withdraw(&Account1, tx.Amount)
	// Deposit(&Account2, tx.Amount) // Change the balance of these 2 Accounts
	db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", gorm.Expr("balance - ?", tx.Amount), tx.From)
	db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", gorm.Expr("balance + ?", tx.Amount), tx.To)
	// result := sess.Save(&Accounts)
	// if result.Error != nil {
	// 	fmt.Println("Failed to save the records :", error)
	// }
	fmt.Printf("Post-Balance : [id=%v] = %v , [id=%v] = %v \n", tx.From, Accounts[0].Balance, tx.To, Accounts[1].Balance)

	// fmt.Println("You [ID :", tx.From, "] have successfully transferred", tx.Amount, "to [ID :", tx.To, "] !")

	// fmt.Println("Affected update :", result.RowsAffected, tx.From, tx.To)
	// fmt.Println("Error Update :", result1.Error, result2.Error)
	// sess.Save(&Account2)
	// sess.Model(&entity.Account{}).Update("balance", Account2.Balance) // Update database

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Accounts)
	// Returns the new state of 2 Accounts when the transaction is done
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

	for tx := range trans {
		if tx.From == tx.To {
			log.Panicf("Please enter correct recipient ID ! %d and %d", tx.From, tx.To)
		}

		err := db.Raw("SELECT balance FROM accounts where id = ? or id = ?", tx.From, tx.To).Limit(2).Scan(&Accounts).Error
		// err2 := sess.First(&Account2, tx.To).Error
		fmt.Println("Balance of Founded Accounts :", Accounts[0].Balance, Accounts[1].Balance)

		if err != nil {
			panic("Cannot find the record by id !")
		}

		if Accounts[0].Balance < minBalance {
			// Current balance is less than minimum balance
			fmt.Println("You dont have enough money to transfer !")
			return
		}
		if Accounts[0].Balance-tx.Amount < minBalance {
			// Make sure the balance after the transaction is not less than the minimum balance
			fmt.Println("The maximum amount that can be transferred is", Accounts[0].Balance-minBalance, "!")
			return
		}
		if tx.Amount < minCost {
			// Make sure the current trading amount is large the minimum trading amount (1000)
			fmt.Println("The minimum amount that can be transferred is", minCost, "!")
			return
		}

		fmt.Printf("Pre-Balance : [id=%v] = %v , [id=%v] = %v \n", tx.From, Accounts[0].Balance, tx.To, Accounts[1].Balance)

		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", gorm.Expr("balance - ?", tx.Amount), tx.From)
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", gorm.Expr("balance + ?", tx.Amount), tx.To)

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
		id, err := strconv.ParseInt(line[0], 0, 64)
		balance, err := strconv.ParseFloat(line[4], 32)
		status, err := strconv.ParseInt(line[5], 0, 64)
		t := time.Now()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		Accounts = append(Accounts, entity.Account{
			Id:          uint64(id),
			Name:        line[1],
			Address:     line[2],
			PhoneNumber: line[3],
			Balance:     float32(balance),
			Status:      int(status),
			Createtime:  &t,
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

		from, err := strconv.ParseInt(line[0], 0, 64)
		// amount, err := strconv.ParseFloat(line[2], 64)
		to, err := strconv.ParseInt(line[3], 0, 64)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		trans = append(trans, entity.Transaction{
			From:   uint64(from),
			Amount: 1000000,
			To:     uint64(to),
		})
	}
	return trans
	// return a list of transactions
}
