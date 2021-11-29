package entity

import "time"

// Account Object
type Account struct {
	Id          uint64     `json:id gorm:"primary_key"`
	Name        string     `json:name`
	Address     string     `json:address`
	PhoneNumber string     `json:phonenumber`
	Balance     float32    `json:balance`
	Status      int        `json:status`
	Createtime  *time.Time `json:createtime`
}

// Transaction Object
type Transaction struct {
	From   uint64  `json:"from"`
	To     uint64  `json:"to"`
	Amount float32 `json:"amount"`
}
