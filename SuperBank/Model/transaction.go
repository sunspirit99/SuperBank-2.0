package Model

import (
	"gorm.io/gorm"
)

type TransactionORM struct {
	Transaction
	db *gorm.DB
}

func NewTransactionORM() *TransactionORM {
	return &TransactionORM{db: Connector}
}

func (tx TransactionORM) CreateTx() (Transaction, error) {
	var tran Transaction
	err := tx.db.Create(&tran).Error
	return tran, err
}

func (tx TransactionORM) UpdateTxID(status int, txid string, trace string) (Transaction, error) {
	var tran Transaction
	err := tx.db.Exec("UPDATE transactions SET status = ?, tx_id = ? WHERE trace = ?", status, txid, trace).Error // status = 2 : Failed
	return tran, err
}
