// go-bigtx - Simple two phase commits implementation on MongoDB with Golang

// Copyright (c) 2016 Chaiwat Shuetrakoonpaiboon. All rights reserved.
//
// Use of this source code is governed by a MIT license that can be found in
// the LICENSE file.

package bigtx

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func connectTest(t *testing.T) {
	session, err := mgo.Dial("mongodb://192.168.99.100/bigtx_test")
	if err != nil {
		t.Fatal("Got connection error to MongoDB:", err)
	}

	rootSession = session
	rootSession.DB("").DropDatabase()
}

func TestCashIn(t *testing.T) {
	connectTest(t)
	defer Disconnect()

	CreateAccount("A0001", AccountSideDebit)
	CreateAccount("A0002", AccountSideDebit)
	CreateAccount("L0001", AccountSideCredit)

	txn, err := BeginTransaction(
		"TX0001",
		map[string]int64{"A0001": 2000, "A0002": 20},
		map[string]int64{"L0001": 2020},
		"CIN")
	if err != nil {
		t.Fatal(err)
	}

	if txn == "" {
		t.Fatal("Expected txn to be a valid string")
	}

	err = CommitTransaction()
	if err != nil {
		t.Fatal(err)
	}

	a1, err := ReadBalance("A0001")
	if err != nil {
		t.Fatal(err)
	}
	if a1 != 2000 {
		t.Fatal("Expected A0001 to be 2000, got", a1)
	}

	a2, err := ReadBalance("A0002")
	if err != nil {
		t.Fatal(err)
	}
	if a2 != 20 {
		t.Fatal("Expected A0002 to be 20, got", a2)
	}

	l1, err := ReadBalance("L0001")
	if err != nil {
		t.Fatal(err)
	}
	if l1 != -2020 {
		t.Fatal("Expected L0001 to be 2020, got", l1)
	}
}

func TestUnbalance(t *testing.T) {
	connectTest(t)
	defer Disconnect()

	CreateAccount("A0001", AccountSideDebit)
	CreateAccount("A0002", AccountSideDebit)
	CreateAccount("L0001", AccountSideCredit)

	_, err := BeginTransaction(
		"TX0001",
		map[string]int64{"A0001": 2000, "A0002": 20},
		map[string]int64{"L0001": 2000},
		"CIN")
	if err == nil {
		t.Fatal("Expected UnbalancedErr, got nil")
	}
	if !IsUnbalancedErr(err) {
		t.Fatal("Expected UnbalancedErr, got", err)
	}
}

func TestRecover(t *testing.T) {
	connectTest(t)
	defer Disconnect()

	CreateAccount("A0001", AccountSideDebit)
	CreateAccount("A0002", AccountSideDebit)
	CreateAccount("L0001", AccountSideCredit)
	CreateAccount("L0002", AccountSideCredit)
	tx1 := Transaction{
		ID:   bson.NewObjectId(),
		Date: time.Now().Add(-60 * time.Second),
		Changes: map[string]int64{
			"A0001": 50,
			"A0002": 50,
			"L0001": -100,
		},
		State: TxPending,
	}
	tx2 := Transaction{
		ID:   bson.NewObjectId(),
		Date: time.Now().Add(-60 * time.Second),
		Changes: map[string]int64{
			"A0002": 100,
			"L0001": -20,
			"L0002": -80,
		},
		State: TxApplied,
	}
	session := rootSession.Copy()
	err := session.DB("").C("transactions").Insert(&tx1)
	if err != nil {
		t.Fatal(err)
	}
	err = session.DB("").C("transactions").Insert(&tx2)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	err = Recover()
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	err = Recover()
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
}
