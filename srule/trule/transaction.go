// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"fmt"
	"sync"
	"time"

	"github.com/idcsource/insight00-lib/spots"
)

// 初始化一个transaction
func initTransaction(id string, tranSig chan *transactionSig, spotcache_sig chan *spotCacheSig) (t *Transaction) {
	t = &Transaction{
		id:              id,
		spot_cache:      make(map[string]map[string]*spotCache),
		spot_cache_name: make(map[string]map[string]bool),
		spot_cache_lock: new(sync.RWMutex),
		tran_time:       time.Now(),
		spot_cache_sig:  spotcache_sig,
		tran_sig:        tranSig,
		be_delete:       false,
	}
	return
}

// If the Spot exist.
func (t *Transaction) ExistSpot(area, id string) (exist bool, err error) {
	if t.be_delete == true {
		return false, fmt.Errorf("trule[Transaction]ExistSpot: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	_, exist, err = t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ExistSpot: %v", err)
	}
	return
}

// Read a Spot
func (t *Transaction) ReadSpot(area, id string) (spot *spots.Spots, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadSpot: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, exist, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transacion]ReadSpot: %v", err)
		return
	}
	if exist == false {
		err = fmt.Errorf("trule[Transacion]ReadSpot: The Spot not exist.")
		return
	}
	spot = spotc.spot
	return
}

func (t *Transaction) ReadSpotWithBody(area, id string, body spots.DataBodyer) (spot *spots.Spots, err error) {
	spot, err = t.ReadSpot(area, id)
	if err != nil {
		return
	}
	err = spot.BtoDataBody(body)
	return
}

// Store Spot
func (t *Transaction) StoreSpot(area string, spot *spots.Spots) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]StoreSpot: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	//	t.lock.RLock()
	//	defer t.lock.RUnlock()
	spotid := spot.GetId()
	spotc, err := t.getSpotForNew(area, spotid)
	if err != nil {
		return
	}
	spot.DataBodyToB()
	spotc.spot = spot
	spotc.be_change = true
	return nil
}

// Delete a Spot
func (t *Transaction) DeleteSpot(area, id string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]DeleteSpot: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, exist, err := t.getSpot(area, id, true)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("drule[Transaction]DeleteSpot: The Spot not exist.")
	}
	spotc.be_delete = TRAN_SPOT_BE_DELETE_YES
	spotc.be_change = true
	return nil
}

func (t *Transaction) WriteFather(area, id, father string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]WriteFather: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, exist, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("drule[Transaction]WriteFather:  %v", err)
		return
	}
	if exist == false {
		err = fmt.Errorf("drule[Transaction]WriteFather: The Spot not exist.")
		return
	}
	spotc.spot_lock.Lock()
	defer spotc.spot_lock.Unlock()
	spotc.spot.SetFather(father)
	spotc.be_change = true
	return
}

// 获取一个角色，forwrite是true就是为了写
func (t *Transaction) getSpot(area, id string, forwrite bool) (spotc *spotCache, exist bool, err error) {
	// 构建信号
	spot_cache_sig := &spotCacheSig{
		ask:      SPOT_CACHE_ASK_GET,
		area:     area,
		id:       id,
		tranid:   t.id,
		forwrite: forwrite,
		ask_time: time.Now(),
		re:       make(chan *spotCacheReturn),
	}
	// 发送信号
	t.spot_cache_sig <- spot_cache_sig
	// 等待返回
	sigre := <-spot_cache_sig.re
	if sigre.status != SPOT_CACHE_RETURN_HANDLE_OK {
		err = sigre.err
		return
	}
	spotc = sigre.spot
	exist = sigre.exist
	// 如果是为了写，就加入自己的缓存
	if forwrite == true {
		t.spot_cache_lock.Lock()
		if _, have := t.spot_cache[area]; have == false {
			t.spot_cache[area] = make(map[string]*spotCache)
			t.spot_cache_name[area] = make(map[string]bool)
		}
		t.spot_cache[area][id] = spotc
		t.spot_cache_name[area][id] = true
		t.spot_cache_lock.Unlock()
	}
	return
}

// 获取一个角色（有可能是空的，这通常是用来新建一个角色），forwrite是true就是为了写
func (t *Transaction) getSpotForNew(area, id string) (spotc *spotCache, err error) {
	// 构建信号
	spot_cache_sig := &spotCacheSig{
		ask:      SPOT_CACHE_ASK_WRITE,
		area:     area,
		id:       id,
		tranid:   t.id,
		forwrite: true,
		ask_time: time.Now(),
		re:       make(chan *spotCacheReturn),
	}
	// 发送信号
	t.spot_cache_sig <- spot_cache_sig
	// 等待返回
	sigre := <-spot_cache_sig.re
	if sigre.status != SPOT_CACHE_RETURN_HANDLE_OK {
		err = sigre.err
		return
	}
	spotc = sigre.spot
	// 加入自己的缓存
	t.spot_cache_lock.Lock()
	if _, have := t.spot_cache[area]; have == false {
		t.spot_cache[area] = make(map[string]*spotCache)
		t.spot_cache_name[area] = make(map[string]bool)
	}
	t.spot_cache[area][id] = spotc
	t.spot_cache_name[area][id] = true
	t.spot_cache_lock.Unlock()
	return
}
