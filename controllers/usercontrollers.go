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

//GetAllPerson get all user data
func GetAllUser(w http.ResponseWriter, r *http.Request) {
	var users []entity.User
	database.Connector.Find(&users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

//GetPersonByID returns user with specific ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["ID"]

	var user entity.User
	database.Connector.First(&user, key)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

//CreatePerson creates user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	var user entity.User
	json.Unmarshal(requestBody, &user)

	database.Connector.Create(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

//UpdatePersonByID updates user with respective ID
func UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	var user entity.User
	json.Unmarshal(requestBody, &user)
	database.Connector.Save(&user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

//DeletePersonByID delete's user with specific ID
func DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	var user entity.User
	id, _ := strconv.ParseInt(key, 10, 64)
	database.Connector.Where("id = ?", id).Delete(&user)
	w.WriteHeader(http.StatusNoContent)
}

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

func UserTransfer(w http.ResponseWriter, r *http.Request) {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Unreadable !!!")
	}
	var cb entity.ChangeBalance

	json.Unmarshal(requestBody, &cb)

	var users []entity.User
	trans := []string{cb.ID, cb.TargetId} //id la tk chuyen , targetId la tk nhan
	database.Connector.Find(&users, trans)
	if cb.Amount > 5000 {
		Withdraw(&users[0], cb.Amount) // rut tien tu tk = amount
		Deposit(&users[1], cb.Amount)  //chuyen tien sang tk = amount
	} else {
		fmt.Print("Amount > 5000")
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

func Withdraw(user *entity.User, num float64) {
	if num < 0 {
		fmt.Println("Invalid input !")
	} else if user.Balance < 5000 {
		fmt.Println("You dont have enough money !")
	} else if user.Balance < num {
		fmt.Println("the amount you need to withdraw must be less than your balance !")
	} else {
		user.Balance = user.Balance - num
	}
}
func Deposit(user *entity.User, num float64) {
	if num < 0 {
		fmt.Print("Invalid input !")
	} else if num < 5000 {
		fmt.Print("Min > 5000")

	} else {
		user.Balance = user.Balance + num

	}

}
