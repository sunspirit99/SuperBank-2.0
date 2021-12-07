package Type

import (
	model "SuperBank/Model"

	"gorm.io/gorm"
)

type TransactionORM struct {
	model.Transaction
	db *gorm.DB
}

func NewTransactionORM() *TransactionORM {
	return &TransactionORM{}
}

func (tx TransactionORM) CreateTx() (model.Transaction, error) {
	var tran model.Transaction
	err := tx.db.Create(&tran).Error
	return tran, err
}

func (tx TransactionORM) UpdateTxID(status int, txid string, trace string) (model.Transaction, error) {
	var tran model.Transaction
	err := tx.db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", status, txid, trace).Error // status = 2 : Failed
	return tran, err
}
