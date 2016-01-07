// go-bigtx - Simple two phase commits implementation on MongoDB with Golang

// Copyright (c) 2016 Chaiwat Shuetrakoonpaiboon. All rights reserved.
//
// Use of this source code is governed by a MIT license that can be found in
// the LICENSE file.

package bigtx

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TxState string

const (
	TxInitial   TxState = "init"
	TxPending           = "pend"
	TxApplied           = "appl"
	TxDone              = "done"
	TxCanceling         = "cing"
	TxCanceled          = "canc"
)

type Transaction struct {
	ID          bson.ObjectId    `bson:"_id"`
	Date        time.Time        `bson:"date"`
	Changes     map[string]int64 `bson:"chg"`
	State       TxState          `bson:"stat"`
	Ref1        string           `bson:"ref1,omitempty"`
	Ref2        string           `bson:"ref2,omitempty"`
	Description string           `bson:"dscr"`
}
