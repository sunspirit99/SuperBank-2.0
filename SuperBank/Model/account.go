package Model

import (
	"fmt"

	"gorm.io/gorm"
)

type AccountORM struct {
	Account
	db *gorm.DB
}

func NewAccountORM() *AccountORM {
	return &AccountORM{db: Connector}
}

func (acc AccountORM) FindAccountByTX(tx Transaction) []Account {
	var Accounts []Account
	acc.db.Raw("SELECT * from accounts WHERE id = ? or id = ?", tx.From, tx.To).Scan(&Accounts)
	return Accounts
}

func (acc AccountORM) FindAllAccount() ([]Account, error) {
	var accounts []Account
	fmt.Println(acc.db)
	err := acc.db.Find(&accounts).Error
	return accounts, err
}

func (acc AccountORM) FindAccountByID(id string) (Account, error) {
	var account Account
	err := acc.db.First(&account, id).Error
	return account, err
}

func (acc AccountORM) CreateAccount(account Account) (Account, error) {
	err := acc.db.Create(&account).Error
	return account, err
}

func (acc AccountORM) SaveAccount(account Account) (Account, error) {
	err := acc.db.Save(&account).Error
	return account, err
}

func (acc AccountORM) DeleteAccount(id string) error {
	var account Account

	err := acc.db.First(&account, id).Error // Return Account with id = key
	if err != nil {
		return err
	}

	return acc.db.Where("id = ?", id).Delete(&account).Error // Delete Account with id = key
}

func (acc AccountORM) UpdateBalance(balance float32, id string) (Account, error) {
	var account Account
	err := acc.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", balance, id).Error
	return account, err
}
