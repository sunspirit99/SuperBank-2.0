package Type

import (
	model "SuperBank/Model"
	"SuperBank/database"
	"fmt"

	"gorm.io/gorm"
)

type AccountORM struct {
	model.Account
	db *gorm.DB
}

func NewAccountORM() *AccountORM {
	return &AccountORM{db: database.Connector}
}

func (acc AccountORM) FindAccountByTX(tx model.Transaction) []model.Account {
	var Accounts []model.Account
	acc.db.Raw("SELECT * from accounts WHERE id = ? or id = ?", tx.From, tx.To).Scan(&Accounts)
	return Accounts
}

func (acc AccountORM) FindAllAccount() ([]model.Account, error) {
	var accounts []model.Account
	fmt.Println(acc.db)
	err := acc.db.Find(&accounts).Error
	return accounts, err
}

func (acc AccountORM) FindAccountByID(id string) (model.Account, error) {
	var account model.Account
	err := acc.db.First(&account, id).Error
	return account, err
}

func (acc AccountORM) CreateAccount() (model.Account, error) {
	var account model.Account
	err := acc.db.Create(&account).Error
	return account, err
}

func (acc AccountORM) SaveAccount() (model.Account, error) {
	var account model.Account
	err := acc.db.Save(&account).Error
	return account, err
}

func (acc AccountORM) DeleteAccount(id string) error {
	var account model.Account

	err := acc.db.First(&account, id).Error // Return Account with id = key
	if err != nil {
		return err
	}

	return acc.db.Where("id = ?", id).Delete(&account).Error // Delete Account with id = key
}

func (acc AccountORM) UpdateBalance(balance float32, id string) (model.Account, error) {
	var account model.Account
	err := acc.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", balance, id).Error
	return account, err
}
