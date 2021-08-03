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

const min float64 = 50000

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
	key := vars["ID"]

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

	error := database.Connector.Create(user).Error
	if error != nil {
		fmt.Println("ban chua nhap du lieu")
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

	//nhap thieu du lieu
	//nhap rong

}

//UpdatePersonByID updates user with respective ID
func UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	var user entity.User
	json.Unmarshal(requestBody, &user)
	error := database.Connector.Save(&user).Error
	if error != nil {
		fmt.Println("Error")
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
	if vars != nil {
		fmt.Println("Enter id ")
	}

	key := vars["id"]

	var user entity.User
	id, _ := strconv.ParseInt(key, 10, 64)
	err := database.Connector.Where("id = ?", id).Delete(&user).Error
	if err != nil {
		fmt.Println("ID doesn't exist")
	}
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
		fmt.Print("error")
	}

	var user entity.User
	database.Connector.First(&user, cb.ID)

	Withdraw(&user, cb.Amount)

	t := time.Now()                                      //set thoi gian hien tai
	user.Modified_time = t.Format("2020-01-02 15:04:05") //truyen vao

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
	database.Connector.First(&user, cb.ID)

	Deposit(&user, cb.Amount)

	t := time.Now()                                      //set thoi gian hien tai
	user.Modified_time = t.Format("2020-01-02 15:04:05") // truyen vao

	database.Connector.Save(&user)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)

}

//take out two 2 ids 1 is the sender's id and 2 is the target id of the recipient for the transfer
func UserTransfer(w http.ResponseWriter, r *http.Request) {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Enter all required information")
	}
	var cb entity.ChangeBalance

	json.Unmarshal(requestBody, &cb)

	var users []entity.User
	trans := []int64{cb.ID, cb.TargetId} //id la tk chuyen , targetId la tk nhan
	database.Connector.Find(&users, trans)
	if users[0].Balance < min {
		fmt.Println("You dont have enough money to transfer")
	} else if cb.Amount < users[0].Balance {
		Withdraw(&users[0], cb.Amount) // rut tien tu tk = amount
		Deposit(&users[1], cb.Amount)  //chuyen tien sang tk = amount
		fmt.Println("Transfer : Successful ", cb.Amount)
	} else {
		fmt.Println(" The amount to be transferred must be greater than the balance")
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
	if num < 0 {
		fmt.Println("Invalid input !")
	} else if user.Balance < min {
		fmt.Println("You dont have enough money !")
	} else if user.Balance < num {
		fmt.Println("the amount you need to withdraw must be less than your balance !")
	} else {
		user.Balance = user.Balance - num
		fmt.Println("Success")

	}
}

//check deposit amount and deposit function
func Deposit(user *entity.User, num float64) {
	if num < 0 {
		fmt.Print("Invalid input !")
	} else if num < min {
		fmt.Println("Minimum amount to deposit is more than ", min)

	} else {
		user.Balance = user.Balance + num
		fmt.Println("Success")
	}

}
