// go-bigtx - Simple two phase commits implementation on MongoDB with Golang

// Copyright (c) 2016 Chaiwat Shuetrakoonpaiboon. All rights reserved.
//
// Use of this source code is governed by a MIT license that can be found in
// the LICENSE file.

package bigtx

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
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

func CreateAccount(acID string, side AccountSide) error {
	session := rootSession.Copy()
	defer session.Close()

	_, err := session.DB("").C("accounts").UpsertId(acID, bson.M{"$setOnInsert": bson.M{"bal": 0, "side": side}})
	return err
}

func ReadBalance(account string) (int64, error) {
	session := rootSession.Copy()
	defer session.Close()

	acct := Account{}
	err := session.DB("").C("accounts").FindId(account).One(&acct)
	if err != nil {
		return 0, fmt.Errorf("bigtx: find error: %v", err)
	}

	return acct.Balance, nil
}
