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
