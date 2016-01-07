package bigtx

import (
	"time"
)

type AccountSide string

const (
	AccountSideDebit  AccountSide = "dbt"
	AccountSideCredit             = "crd"
)

type Account struct {
	ID                  string    `bson:"_id"`
	Name                string    `bson:"name"`
	Side                string    `bson:"side"`
	Balance             int64     `bson:"bal"`
	Date                time.Time `bson:"date"`
	PendingTransactions []string  `bson:"txs"`
}
