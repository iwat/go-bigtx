// go-bigtx - Simple two phase commits implementation on MongoDB with Golang

// Copyright (c) 2016 Chaiwat Shuetrakoonpaiboon. All rights reserved.
//
// Use of this source code is governed by a MIT license that can be found in
// the LICENSE file.

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
