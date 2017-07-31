// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	_ "github.com/idcsource/insight00-lib/drule2/types"
)

// Begin a transaction
func (t *TRule) Begin() (tran *Transaction, err error) {
	transig := &transactionSig{
		ask: TRANSACTION_ASK_BEGIN,
		re:  make(chan *transactionReturn),
	}
	t.transaction_signal <- transig
	treturn := <-transig.re
	if treturn.status != TRAN_RETURN_HANDLE_OK {
		err = treturn.err
		return
	}
	tran = treturn.tran
	return
}
