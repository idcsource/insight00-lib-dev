// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"sync"

	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/srule/hardstorage"
)

func NewTRule(local_store *hardstorage.HardStorage, log *ilogs.Logs) (t *TRule) {
	t = &TRule{
		local_store:        local_store,
		log:                log,
		spot_cache_sig:     make(chan *spotCacheSig),
		transaction_signal: make(chan *transactionSig),
		pausing_signal:     make(chan bool),
		paused_signal:      make(chan bool),
		work_status:        TRULE_RUN_PAUSED,
		tran_wait:          new(sync.WaitGroup),
	}
	return
}

func (t *TRule) Start() {

}

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
