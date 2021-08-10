package controllers

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"rest-go-demo/database"
	"rest-go-demo/entity"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const minBalance float64 = 5000 // Minimum balance of an user
const minCost float64 = 1000    // Minimum amount to trade in a transaction
const batchsize int = 1000      // the number of records in a single write to the database

// GetAllUser get all user data
func GetAllUser(w http.ResponseWriter, r *http.Request) {

	var users []entity.User

	error := database.Connector.Find(&users).Error // // Return all users exists in table of database
	if error != nil {
		fmt.Println("Query error !")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
	// Returns a list of users in JSON format
}

// GetUserByID get user with specific ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {

	var user entity.User
	vars := mux.Vars(r)
	key := vars["id"] // Get id from URL path

	error := database.Connector.First(&user, key).Error // Return user with id = key
	if error != nil {
		fmt.Println("Query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	// Returns the found user in JSON format
}

//CreateUser creates an user
func CreateUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body

	var user entity.User
	json.Unmarshal(requestBody, &user) // Convert from JSON format to User Format

	t := time.Now()
	user.Created_time = fmt.Sprintf("%v", t.Format("2020-01-02 15:04:05")) // Update Current time
	user.Modified_time = ""

	error := database.Connector.Create(user).Error // Create a record in database
	if error != nil {
		fmt.Println("Query Error !")
	}

	fmt.Printf("\n Created an account complete at %v", user.Created_time)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusCreated)

	// return the created user in JSON format
}

// CreateUserFromCSV creates a list of user in CSV file
func CreateUserFromCSV(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Processing !!!")
	start1 := time.Now()
	var users = LoadUsersCSV()
	end1 := time.Since(start1) // Total time to read data from CSV file

	start2 := time.Now()
	error := database.Connector.Statement.CreateInBatches(users, batchsize).Error
	// Load multiple records (const batchsize = 1000) into 1 batch and then write to database

	if error != nil {
		fmt.Println("Query Error !")
	}

	end2 := time.Since(start2) // Total time to insert data to Database
	fmt.Printf("\n Created 100.000 accounts complete at %v", time.Now())
	fmt.Printf("\n Time to read data from CSV file is : %v \n Time to write to DB is : %v \n", end1, end2)
	w.WriteHeader(http.StatusOK)

}

// UpdateUserByID updates user with respective ID, if ID does not exist, create a new user
func UpdateUserByID(w http.ResponseWriter, r *http.Request) {

	requestBody, _ := ioutil.ReadAll(r.Body) // read JSON data from Body
	var user entity.User
	json.Unmarshal(requestBody, &user) // Convert from JSON format to User Format

	t := time.Now()
	error := database.Connector.Save(&user).Error // Update database
	if error != nil {
		fmt.Println("Query Error !")
	}

	user.Modified_time = fmt.Sprintf(t.Format("2020-01-02 15:04:05")) // Update current time
	fmt.Printf("\n Updating complete at %v", user.Modified_time)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

	fmt.Println("Updated user successfully !")
	// Return the updated user in JSON format
}

// DeleteUserByID delete an user with specific ID
func DeleteUserByID(w http.ResponseWriter, r *http.Request) {

	var user entity.User
	vars := mux.Vars(r) // Get id from URL path
	key := vars["id"]
	if len(vars) == 0 {
		panic("Enter an ID !")
	}

	err := database.Connector.First(&user, key).Error // Return user with id = key
	if err != nil {
		fmt.Println("ID doesn't exist")
		return
	}

	database.Connector.Where("id = ?", key).Delete(&user) // Delete user with id = key
	fmt.Println("[ID :", key, "] has been successfully deleted !")
	w.WriteHeader(http.StatusNoContent)

}

// UserWithdraw withdraw money from account through a transaction
func UserWithdraw(w http.ResponseWriter, r *http.Request) {

	var user entity.User
	var cb entity.ChangeBalance

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		fmt.Println("Unreadable !!!")
	}

	err1 := json.Unmarshal(requestBody, &cb) // Convert from JSON format to User Format
	if err1 != nil {
		fmt.Println("Error")
	}

	error := database.Connector.Where(`id =? AND name=?`, cb.ID, cb.Name).First(&user).Error
	// Get user information to withdraw money

	if error != nil {
		panic("Query error !!!")
	}
	if user.Balance < minBalance {
		// Current balance is less than minimum balance
		fmt.Println("You dont have enough money to withdraw !")
		return
	}
	if user.Balance-cb.Amount < minBalance {
		// Make sure the balance after the transaction is not less than the minimum balance
		fmt.Println("The maximum amount that can be withdrawn is", user.Balance-minBalance, "!")
		return
	}
	if cb.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (1000)
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	}

	Withdraw(&user, cb.Amount) // Change the balance of this user
	fmt.Println("You have successfully withdrew", cb.Amount, "from your account !")

	database.Connector.Save(&user) // Update database
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	// Returns the new state of the user when the transaction is done
}

// UserDeposit deposit money into account through a transaction
func UserDeposit(w http.ResponseWriter, r *http.Request) {
	var cb entity.ChangeBalance
	var user entity.User

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		fmt.Println("Unreadble ")
	}

	err1 := json.Unmarshal(requestBody, &cb) // Convert from JSON format to User Format
	if err1 != nil {
		fmt.Println("Error")
	}

	error := database.Connector.Where(`id =? AND name=?`, cb.ID, cb.Name).First(&user).Error
	// Get user information to deposit money

	if error != nil {
		panic("Query error !!!")
	}
	if cb.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (5000)
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	}

	Deposit(&user, cb.Amount) // Change the balance of this user
	fmt.Println("You have successfully deposited", cb.Amount, "to your account !")

	database.Connector.Save(&user) // Update database
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	// Returns the new state of the user when the transaction is done
}

// UserTransfer transfer money between 2 accounts through a transaction
func UserTransfer(w http.ResponseWriter, r *http.Request) {
	var users []entity.User
	var user1 entity.User
	var user2 entity.User
	var cb entity.ChangeBalance

	requestBody, err := ioutil.ReadAll(r.Body) // read JSON data from Body
	if err != nil {
		panic("Enter all required information !!!")
	}

	error := json.Unmarshal(requestBody, &cb) // Convert from JSON format to User Format
	if error != nil {
		fmt.Println("Error !")
	}

	if cb.ID == cb.TargetId {
		panic("Please enter correct recipient ID !")
	}

	err1 := database.Connector.Where(`id =? AND name=?`, cb.ID, cb.Name).First(&user1).Error
	err2 := database.Connector.Where(`id =? `, cb.TargetId).First(&user2).Error
	if err1 != nil || err2 != nil {
		panic("Query error !")
	}

	users = append(users, user1, user2) // Contains 2 users participating in the transaction

	if users[0].Balance < minBalance {
		// Current balance is less than minimum balance
		fmt.Println("You dont have enough money to transfer !")
		return
	}
	if users[0].Balance-cb.Amount < minBalance {
		// Make sure the balance after the transaction is not less than the minimum balance
		fmt.Println("The maximum amount that can be transferred is", users[0].Balance-minBalance, "!")
		return
	}
	if cb.Amount < minCost {
		// Make sure the current trading amount is large the minimum trading amount (1000)
		fmt.Println("The minimum amount that can be transferred is", minCost, "!")
		return
	}

	Withdraw(&users[0], cb.Amount)
	Deposit(&users[1], cb.Amount) // Change the balance of these 2 users
	fmt.Println("You [ID :", cb.ID, "] have successfully transferred", cb.Amount, "to [ID :", cb.TargetId, "] !")

	t := time.Now()
	users[0].Modified_time = t.Format("2006-01-02 15:04:05")
	users[1].Modified_time = t.Format("2006-01-02 15:04:05") // Update current time

	database.Connector.Save(&users[0])
	database.Connector.Save(&users[1]) // Update database

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
	// Returns the new state of 2 users when the transaction is done
}

func UserTransferFromCSV(w http.ResponseWriter, r *http.Request) {

	trans := LoadTransactionsCSV() // LoadTransactionsCSV return a list of user from CSV file

	for _, tran := range trans {
		var users []entity.User
		var user1 entity.User
		var user2 entity.User

		if tran.ID == tran.TargetId {
			panic("Please enter correct recipient ID !")
		}

		err1 := database.Connector.Where(`id =? AND name=?`, tran.ID, tran.Name).First(&user1).Error
		err2 := database.Connector.Where(`id =? `, tran.TargetId).First(&user2).Error
		if err1 != nil || err2 != nil {
			fmt.Println("Please enter correct information !")
			continue
		}

		users = append(users, user1, user2) // Contains 2 users participating in a transaction

		if users[0].Balance < minBalance {
			fmt.Println("You dont have enough money to transfer !")
			return
		}
		if users[0].Balance-tran.Amount < minBalance {
			fmt.Println("The maximum amount that can be transferred is", users[0].Balance-minBalance, "!")
			return
		}
		if tran.Amount < minCost {
			fmt.Println("The minimum amount that can be transferred is", minCost, "!")
			return
		}

		Withdraw(&users[0], tran.Amount)
		Deposit(&users[1], tran.Amount) // Change the balance of these 2 users
		fmt.Println("you [ID :", tran.ID, "] have successfully transferred", tran.Amount, "to [ID :", tran.TargetId, "] !")

		t := time.Now()
		users[0].Modified_time = t.Format("2006-01-02 15:04:05")
		users[1].Modified_time = t.Format("2006-01-02 15:04:05") // Update current time

		database.Connector.Save(&users[0])
		database.Connector.Save(&users[1]) // Update database

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
	w.WriteHeader(http.StatusOK)
	// Returns the new state of users when transactions are processed
}

// Update balance when there's a qualified withdrawal request
func Withdraw(user *entity.User, num float64) {
	user.Balance = user.Balance - num
}

// Update balance when there's a qualified deposit request
func Deposit(user *entity.User, num float64) {
	user.Balance = user.Balance + num
}

// LoadUsersCSV reads users from CSV file
func LoadUsersCSV() []entity.User {

	var users []entity.User
	file, _ := os.Open("users-100k.csv")
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
		balance, err := strconv.ParseFloat(line[2], 64)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		users = append(users, entity.User{
			ID:            id,
			Name:          line[1],
			Balance:       balance,
			Created_time:  line[3],
			Modified_time: line[4],
		})
	}
	return users
	// Return a list of Users
}

// LoadTransactionsCSV reads transactions from CSV file
func LoadTransactionsCSV() []entity.ChangeBalance {
	var trans []entity.ChangeBalance
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

		id, err := strconv.ParseInt(line[0], 0, 64)
		balance, err := strconv.ParseFloat(line[2], 64)
		targetId, err := strconv.ParseInt(line[3], 0, 64)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		trans = append(trans, entity.ChangeBalance{
			ID:       id,
			Name:     line[1],
			Amount:   balance,
			TargetId: targetId,
		})
	}
	return trans
	// return a list of transactions
}
