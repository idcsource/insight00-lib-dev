// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"sync"
	"time"
)

func initTransaction(id string, tranSig chan *transactionSig, rolecache chan *roleCacheSig) (t *Transaction) {
	t = &Transaction{
		id:              id,
		tran_cache:      make(map[string]*roleCache),
		tran_cache_lock: new(sync.RWMutex),
		tran_time:       time.Now(),
		role_cache:      rolecache,
		tran_sig:        tranSig,
		be_delete:       false,
	}
	return
}
