// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"fmt"
	"time"

	"github.com/idcsource/insight00-lib/random"
)

func initTransactionOp(spotCache chan *spotCacheSig) (top *transactionOp) {
	top = &transactionOp{
		signal:            make(chan *transactionSig, TRANSACTION_SIGNAL_CHANNEL_LEN),
		transaction:       make(map[string]*Transaction),
		spotCache:         spotCache,
		count_transaction: 0,
		tran_timeout:      TRANSACTION_CLEAN_TIME_OUT,
		closed:            2,
		closesig:          make(chan bool),
	}
	return
}

func (top *transactionOp) ReturnSig() (sig chan *transactionSig) {
	return top.signal
}

// 启动
func (top *transactionOp) Start() (err error) {
	if top.closed != 2 {
		err = fmt.Errorf("The Transaction Operator not be completed closed.")
		return
	}
	top.closed = 0
	go top.listen()
	return
}

// 关闭
func (top *transactionOp) Stop() {
	top.closed = 1
	top.closesig <- true
}

// 监听信号
func (top *transactionOp) listen() {
	for {
		if top.closed == 1 && top.count_transaction == 0 {
			top.closed = 2
			return
		}
		select {
		case t_sig := <-top.signal:
			top.doSignal(t_sig)
		case <-top.closesig:
			if top.count_transaction == 0 {
				return
			}
		}
	}
}

// 处理信号
func (top *transactionOp) doSignal(sig *transactionSig) {
	switch sig.ask {
	case TRANSACTION_ASK_BEGIN:
		top.toBegin(sig)
	case TRANSACTION_ASK_COMMIT:
		top.toCommit(sig)
	case TRANSACTION_ASK_ROLLBACK:
		top.toRollback(sig)
	case TRANSACTION_ASK_CLEAN:

	}
}

// 处理begin信号
func (top *transactionOp) toBegin(sig *transactionSig) {
	unid := random.Unid(1, time.Now().String())
	tran := initTransaction(unid, top.signal, top.spotCache)
	top.transaction[unid] = tran
	top.count_transaction++
	go top.gotoBegin(sig, tran)
}

// 将begin的最后一部分进入协程
func (top *transactionOp) gotoBegin(sig *transactionSig, tran *Transaction) {
	re := &transactionReturn{
		status: TRAN_RETURN_HANDLE_OK,
		tran:   tran,
	}
	sig.re <- re
}

// 处理commit信号
func (top *transactionOp) toCommit(sig *transactionSig) {
	tran, have := top.transaction[sig.id]
	if have == false {
		re := &transactionReturn{
			status: TRAN_RETURN_HANDLE_ERROR,
			err:    fmt.Errorf("The transaction not exist."),
		}
		sig.re <- re
		return
	}
	tran.be_delete = true
	delete(top.transaction, sig.id)
	top.count_transaction--
	go top.gotoCommit(sig, tran)
}

// 协程中的commit
func (top *transactionOp) gotoCommit(sig *transactionSig, tran *Transaction) {
	// 大量TODO
	for area, sp := range tran.spot_cache {
		for id, _ := range sp {
			signal := &spotCacheSig{
				ask:  SPOT_CACHE_ASK_STORE,
				area: area,
				id:   id,
				//	re:   make(chan *spotCacheReturn),
			}
			top.spotCache <- signal
			//<-signal.re
			//			re := <-signal.re
			//			if re.status == SPOT_CACHE_RETURN_HANDLE_ERROR {
			//				top.log.ErrLog(re.err)
			//			}
		}
	}

	tran.spot_cache = nil

	re := &transactionReturn{
		status: TRAN_RETURN_HANDLE_OK,
		tran:   tran,
	}
	sig.re <- re
}

// 处理rollback信号
func (top *transactionOp) toRollback(sig *transactionSig) {
	tran, have := top.transaction[sig.id]
	if have == false {
		re := &transactionReturn{
			status: TRAN_RETURN_HANDLE_ERROR,
			err:    fmt.Errorf("The transaction not exist."),
		}
		sig.re <- re
		return
	}
	tran.be_delete = true
	delete(top.transaction, sig.id)
	top.count_transaction--
	go top.gotoRollback(sig, tran)
}

// 协程中的rollback
func (top *transactionOp) gotoRollback(sig *transactionSig, tran *Transaction) {

}
