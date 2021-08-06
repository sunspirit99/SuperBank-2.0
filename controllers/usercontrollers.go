package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"rest-go-demo/database"
	"rest-go-demo/entity"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const minBalance float64 = 5000
const minCost float64 = 1000

//GetAllPerson get all user data
func GetAllUser(w http.ResponseWriter, r *http.Request) {
	var users []entity.User
	error := database.Connector.Find(&users).Error
	if error != nil {
		fmt.Println("Error")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
	//truong hop rong
}

//GetPersonByID returns user with specific ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]

	var user entity.User
	error := database.Connector.First(&user, key).Error
	if error != nil {
		fmt.Println("Error")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	//truong hop id rong
	//truong hop id k co trong db
}

//CreatePerson creates user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requestBody, _ := ioutil.ReadAll(r.Body)
	var user entity.User
	json.Unmarshal(requestBody, &user)
	t := time.Now()
	error := database.Connector.Create(user).Error
	if error != nil {
		fmt.Println("Fill your correct info to continue")
	} else {
		user.Created_time = fmt.Sprintf("%v", t.Format("2020-01-02 15:04:05"))
		user.Modified_time = ""
		fmt.Printf("\n Created an account complete at %v", user.Created_time)
	}

	//nhap thieu du lieu
	//nhap rong

}

//UpdatePersonByID updates user with respective ID
func UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	var user entity.User
	json.Unmarshal(requestBody, &user)
	t := time.Now()
	error := database.Connector.Save(&user).Error
	if error != nil {
		fmt.Println("Error")
	} else {
		user.Modified_time = fmt.Sprintf(t.Format("2020-01-02 15:04:05"))
		fmt.Printf("\n Updating complete at %v", user.Modified_time)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	fmt.Println("Update person success")
	//nhap thieu du lieu
	//nhap sai du lieu

}

//DeletePersonByID delete's user with specific ID
func DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if len(vars) == 0 {
		fmt.Println("Enter an ID !")
	}

	key := vars["id"]

	var user entity.User
	id, _ := strconv.ParseInt(key, 10, 64)

	err := database.Connector.First(&user, key).Error
	if err != nil {
		fmt.Println("ID doesn't exist")
		return
	}
	database.Connector.Where("id = ?", id).Delete(&user)
	fmt.Println("[ID :", key, "] has been successfully deleted !")
	w.WriteHeader(http.StatusNoContent)

}

//take out the change balance id enter the amount to withdraw
func UserWithdraw(w http.ResponseWriter, r *http.Request) {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Unreadable !!!")
	}

	var cb entity.ChangeBalance

	err1 := json.Unmarshal(requestBody, &cb)
	if err1 != nil {
		fmt.Print("Error")
	}

	var user entity.User
	database.Connector.Where(`id =? AND name=?`, cb.ID, cb.Name).First(&user)
	if user.Balance < minBalance {
		fmt.Println("You dont have enough money to withdraw !")
		return
	} else if user.Balance-cb.Amount < minBalance {
		fmt.Println("The maximum amount that can be withdrawn is", user.Balance-minBalance, "!")
		return
	} else if cb.Amount < minCost {
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	} else {
		Withdraw(&user, cb.Amount)
		fmt.Println("you have successfully withdrew", cb.Amount, "from your account !")
	}

	//t := time.Now()                                      //set thoi gian hien tai
	//user.Modified_time = t.Format("2020-01-02 15:04:05") //truyen vao

	database.Connector.Save(&user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

}

//take out the change balance id enter the amount to deposit
func UserDeposit(w http.ResponseWriter, r *http.Request) {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Unreadble ")

	}
	var cb entity.ChangeBalance
	err1 := json.Unmarshal(requestBody, &cb)
	if err1 != nil {
		fmt.Print("error")
	}
	var user entity.User
	database.Connector.Where(`id =? AND name=?`, cb.ID, cb.Name).First(&user)

	if cb.Amount < minCost {
		fmt.Println("The minimum amount to perform a transaction is", minCost, "!")
		return
	} else {
		Deposit(&user, cb.Amount)
		fmt.Println("you have successfully deposited", cb.Amount, "to your account !")
	}

	//t := time.Now()                                      //set thoi gian hien tai
	//user.Modified_time = t.Format("2020-01-02 15:04:05") // truyen vao

	database.Connector.Save(&user)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

}

//take out two 2 ids 1 is the sender's id and 2 is the target id of the recipient for the transfer
func UserTransfer(w http.ResponseWriter, r *http.Request) {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Enter all required information !!!")
	}
	var cb entity.ChangeBalance

	json.Unmarshal(requestBody, &cb)

	var users []entity.User
	trans := []int64{cb.ID, cb.TargetId} //id la tk chuyen , targetId la tk nhan
	database.Connector.Find(&users, trans)
	if users[0].Balance < minBalance {
		fmt.Println("You dont have enough money to transfer !")
		return
	} else if users[0].Balance-cb.Amount < minBalance {
		fmt.Println("The maximum amount that can be transferred is", users[0].Balance-minBalance, "!")
		return
	} else if cb.Amount < minCost {
		fmt.Println("The minimum amount that can be transferred is", minCost, "!")
		return
	} else {
		Withdraw(&users[0], cb.Amount)
		Deposit(&users[1], cb.Amount)
		fmt.Println("you have successfully transferred", cb.Amount, "to id :", cb.TargetId, "!")
	}
	t := time.Now()
	users[0].Modified_time = t.Format("2006-01-02 15:04:05")
	users[1].Modified_time = t.Format("2006-01-02 15:04:05")

	database.Connector.Save(&users[0]) //save vao db
	database.Connector.Save(&users[1])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)

}

// check withdraw amount and withdrawal function
func Withdraw(user *entity.User, num float64) {
	user.Balance = user.Balance - num
}

//check deposit amount and deposit function
func Deposit(user *entity.User, num float64) {
	user.Balance = user.Balance + num
}
