package entity

import "time"

// Account Object
type Account struct {
	Id          string  `json:id gorm:"primary_key"`
	Name        string  `json:name`
	Address     string  `json:address`
	PhoneNumber string  `json:phonenumber`
	Balance     float32 `json:balance`
	Status      int     `json:status`
	Createtime  string  `json:createtime`
}

// Transaction Object
type Transaction struct {
	Trace      string     `json:trace gorm:"PRIMARY_KEY"`
	TxID       string     `json:txid`
	From       string     `json:"from"`
	To         string     `json:"to"`
	Amount     float32    `json:"amount"`
	Status     int        `json:"status"`
	Createtime *time.Time `json:"createtime"`
}
