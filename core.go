// go-bigtx - Simple two phase commits implementation on MongoDB with Golang

// Copyright (c) 2016 Chaiwat Shuetrakoonpaiboon. All rights reserved.
//
// Use of this source code is governed by a MIT license that can be found in
// the LICENSE file.

package bigtx

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rootSession *mgo.Session

func Connect() {
	session, err := mgo.Dial("mongodb://localhost/bigtx")
	if err != nil {
		log.Fatal("Got connection error to MongoDB:", err)
	}

	rootSession = session
}

func Disconnect() {
	if rootSession != nil {
		rootSession.Close()
		rootSession = nil
	}
}

func BeginTransaction(txn string, debit map[string]int64, credit map[string]int64, note string) (string, error) {
	// Validate transaction
	debitSum := int64(0)
	creditSum := int64(0)
	for _, amt := range debit {
		debitSum += amt
	}
	for acct, amt := range credit {
		if _, ok := debit[acct]; ok {
			return "", errDuplicatedAcct
		}
		creditSum += amt
	}
	if debitSum != creditSum {
		return "", errUnbalanced
	}

	session := rootSession.Copy()
	defer session.Close()

	// Insert transaction
	tx := Transaction{
		ID:      bson.NewObjectId(),
		Date:    time.Now(),
		Changes: make(map[string]int64, len(debit)+len(credit)),
		State:   TxInitial,
	}
	for acct, amt := range debit {
		tx.Changes[acct] = amt
	}
	for acct, amt := range credit {
		tx.Changes[acct] = -amt
	}
	err := session.DB("").C("transactions").Insert(tx)
	if err != nil {
		return "", fmt.Errorf("bigtx: insert error: %v", err)
	}

	return string(tx.ID), nil
}

func CommitTransaction() error {
	session := rootSession.Copy()
	defer session.Close()

	tx := Transaction{}

	// Update transaction state to pending.
	query := bson.M{"stat": TxInitial}
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"stat": TxPending}, "$currentDate": bson.M{"date": true}},
		ReturnNew: true,
	}
	nfo, err := session.DB("").C("transactions").Find(query).Apply(change, &tx)
	if err != nil {
		return fmt.Errorf("bigtx: findAndModify initial tx error: %v", err)
	}
	if nfo.Updated == 0 {
		return nil
	}

	return Apply(session, &tx)
}

func Apply(session *mgo.Session, tx *Transaction) error {
	// Apply the transaction to all accounts.
	for acID, amt := range tx.Changes {
		query := bson.M{"_id": acID, "txs": bson.M{"$ne": tx.ID}}
		update := bson.M{"$inc": bson.M{"bal": amt}, "$push": bson.M{"txs": tx.ID}}
		session.DB("").C("accounts").Update(query, update)
	}

	// Update transaction state to applied.
	query := bson.M{"_id": tx.ID, "stat": TxPending}
	update := bson.M{"$set": bson.M{"stat": TxApplied}, "$currentDate": bson.M{"date": true}}
	err := session.DB("").C("transactions").Update(query, update)
	if err != nil {
		return fmt.Errorf("bigtx: update state pending -> apply error: %v", err)
	}

	return MarkDone(session, tx)
}

func MarkDone(session *mgo.Session, tx *Transaction) error {
	// Update all accounts’ list of pending transactions.
	for acID, _ := range tx.Changes {
		query := bson.M{"_id": acID, "txs": tx.ID}
		update := bson.M{"$pull": bson.M{"txs": tx.ID}}
		session.DB("").C("accounts").Update(query, update)
	}

	// Update transaction state to done.
	query := bson.M{"_id": tx.ID, "stat": TxApplied}
	update := bson.M{"$set": bson.M{"stat": TxDone}, "$currentDate": bson.M{"date": true}}
	err := session.DB("").C("transactions").Update(query, update)
	if err != nil {
		return fmt.Errorf("bigtx: update state apply -> done error: %v", err)
	}

	return nil
}

// Recover resumes the transaction in process that crashed
func Recover() error {
	session := rootSession.Copy()
	defer session.Close()

	tx := Transaction{}
	query := bson.M{
		"stat": TxPending,
		"date": bson.M{"$lt": time.Now().Add(-30 * time.Second)},
	}
	err := session.DB("").C("transactions").Find(query).One(&tx)
	if err != nil && err != mgo.ErrNotFound {
		return fmt.Errorf("bigtx: find for recover: %v", err)
	}

	if err == nil {
		err = Apply(session, &tx)
		if err != nil {
			return err
		}
	}

	query = bson.M{
		"stat": TxApplied,
		"date": bson.M{"$lt": time.Now().Add(-30 * time.Second)},
	}
	err = session.DB("").C("transactions").Find(query).One(&tx)
	if err != nil && err != mgo.ErrNotFound {
		return fmt.Errorf("bigtx: find for recover: %v", err)
	}

	if err == nil {
		return MarkDone(session, &tx)
	}

	return nil
}

var errUnbalanced = errors.New("bigtx: debit/credit not balance")

func IsUnbalancedErr(err error) bool {
	return err == errUnbalanced
}

var errDuplicatedAcct = errors.New("bigtx: duplicated account")

func IsDuplicatedAcctErr(err error) bool {
	return err == errDuplicatedAcct
}
