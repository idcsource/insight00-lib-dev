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
		return fmt.Errorf("trule[Transaction]StoreSpot: This transaction has been deleted.")
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
		return fmt.Errorf("trule[Transaction]DeleteSpot: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, exist, err := t.getSpot(area, id, true)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("trule[Transaction]DeleteSpot: The Spot not exist.")
	}
	spotc.be_delete = TRAN_SPOT_BE_DELETE_YES
	spotc.be_change = true
	return nil
}

func (t *Transaction) WriteFather(area, id, father string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("trule[Transaction]WriteFather: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, exist, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFather:  %v", err)
		return
	}
	if exist == false {
		err = fmt.Errorf("trule[Transaction]WriteFather: The Spot not exist.")
		return
	}
	//spotc.spot_lock.Lock()
	//defer spotc.spot_lock.Unlock()
	spotc.spot.SetFather(father)
	spotc.be_change = true
	return
}

func (t *Transaction) ReadFather(area, id string) (father string, err error) {
	if t.be_delete == true {
		return "", fmt.Errorf("trule[Transaction]ReadFather: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFather: %v", err)
		return
	}
	father = spotc.spot.GetFather()
	return
}

func (t *Transaction) ResetFather(area, id string) (err error) {
	return t.WriteFather(area, id, "")
}

func (t *Transaction) ReadChildren(area, id string) (children []string, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadChildren: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadChildren: %v", err)
		return
	}
	children = spotc.spot.GetChildren()
	return
}

func (t *Transaction) WriteChildren(area, id string, children []string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteChildren: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteChildren: %v", err)
		return
	}
	spotc.spot_lock.Lock()
	defer spotc.spot_lock.Unlock()
	spotc.spot.SetChildren(children)
	spotc.be_change = true
	return
}

func (t *Transaction) ResetChildren(area, id string) (err error) {
	children := make([]string, 0)
	return t.WriteChildren(area, id, children)
}

func (t *Transaction) WriteChild(area, id, child string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("trule[Transaction]WriteChild: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteChild: %v", err)
		return
	}
	spotc.spot_lock.Lock()
	defer spotc.spot_lock.Unlock()
	spotc.spot.AddChild(child)
	spotc.be_change = true
	return
}

func (t *Transaction) DeleteChild(area, id, child string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("trule[Transaction]DeleteChild: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]DeleteChild: %v", err)
		return
	}
	spotc.spot_lock.Lock()
	defer spotc.spot_lock.Unlock()
	spotc.spot.DeleteChild(child)
	spotc.be_change = true
	return
}

func (t *Transaction) ExistChild(area, id, child string) (have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ExistChild: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ExistChild: %v", err)
		return
	}
	have = spotc.spot.ExistChild(child)
	return
}

func (t *Transaction) ReadFriends(area, id string) (friends map[string]spots.Status, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadFriends: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriends: %v", err)
		return
	}
	friends = spotc.spot.GetFriends()
	return
}

func (t *Transaction) WriteFriends(area, id string, friends map[string]spots.Status) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteFriends: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriends: %v", err)
		return
	}
	spotc.spot.SetFriends(friends)
	spotc.be_change = true
	return
}

func (t *Transaction) ResetFriends(area, id string) (err error) {
	friends := make(map[string]spots.Status)
	return t.WriteFriends(area, id, friends)
}

func (t *Transaction) WriteFriendIntStatus(area, id, friend string, bindbit int, value int64) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
		return
	}
	err = spotc.spot.SetFriendIntStatus(id, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteFriendFloatStatus(area, id, friend string, bindbit int, value float64) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
		return
	}
	err = spotc.spot.SetFriendFloatStatus(id, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteFriendComplexStatus(area, id, friend string, bindbit int, value complex128) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
		return
	}
	err = spotc.spot.SetFriendComplexStatus(id, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteFriendStringStatus(area, id, friend string, bindbit int, value string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
		return
	}
	err = spotc.spot.SetFriendStringStatus(id, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteFriendStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) ReadFriendIntStatus(area, id, friend string, bindbit int) (value int64, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetFriendIntStatus(friend, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadFriendFloatStatus(area, id, friend string, bindbit int) (value float64, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetFriendFloatStatus(friend, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadFriendComplexStatus(area, id, friend string, bindbit int) (value complex128, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetFriendComplexStatus(friend, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadFriendStringStatus(area, id, friend string, bindbit int) (value string, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetFriendStringStatus(friend, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadFriendStatus: %v", err)
	}
	return
}

func (t *Transaction) DeleteFriend(area, id, friend string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]DeleteFriend: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]DeleteFriend: %v", err)
		return
	}
	spotc.spot.DeleteFriend(friend)
	spotc.be_change = true
	return
}

func (t *Transaction) CreateContext(area, id, contextname string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]CreateContext: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]CreateContext: %v", err)
		return
	}
	err = spotc.spot.NewContext(contextname)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]CreateContext: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) ExistContext(area, id, contextname string) (have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ExistContext: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ExistContext: %v", err)
		return
	}
	have = spotc.spot.ExistContext(contextname)
	return
}

func (t *Transaction) DropContext(area, id, contextname string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]DropContext: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]DropContext: %v", err)
		return
	}
	spotc.spot.DelContext(contextname)
	spotc.be_change = true
	return
}

func (t *Transaction) ReadContext(area, id, contextname string) (context spots.Context, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContext: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContext: %v", err)
		return
	}
	context, have = spotc.spot.GetContext(contextname)
	return
}

func (t *Transaction) WriteContext(area, id, contextname string, context spots.Context) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContext: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContext: %v", err)
		return
	}
	spotc.spot.SetContext(contextname, context)
	spotc.be_change = true
	return
}

func (t *Transaction) DeleteContextBind(area, id, contextname string, upordown spots.ContextUpDown, bindspot string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]DeleteContextBind: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]DeleteContextBind: %v", err)
		return
	}
	if upordown == spots.CONTEXT_UP {
		spotc.spot.DelContextUp(contextname, bindspot)
	} else if upordown == spots.CONTEXT_DOWN {
		spotc.spot.DelContextDown(contextname, bindspot)
	} else {
		err = fmt.Errorf("trule[Transaction]DeleteContextBind: Must CONTEXT_UP or CONTEXT_DOWN.")
	}
	spotc.be_change = true
	return
}

func (t *Transaction) ReadContextsName(area, id string) (names []string, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContextsName: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextsName: %v", err)
		return
	}
	names = spotc.spot.GetContextsName()
	return
}

func (t *Transaction) WriteContextIntStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int, value int64) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
		return
	}
	err = spotc.spot.SetContextIntStatus(contextname, upordown, bindspotid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteContextFloatStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int, value float64) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
		return
	}
	err = spotc.spot.SetContextFloatStatus(contextname, upordown, bindspotid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteContextComplexStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int, value complex128) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
		return
	}
	err = spotc.spot.SetContextComplexStatus(contextname, upordown, bindspotid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteContextStringStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int, value string) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
		return
	}
	err = spotc.spot.SetContextStringStatus(contextname, upordown, bindspotid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContextStatus: %v", err)
	}
	spotc.be_change = true
	return
}

func (t *Transaction) ReadContextIntStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int) (value int64, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetContextIntStatus(contextname, upordown, bindspotid, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadContextFloatStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int) (value float64, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetContextFloatStatus(contextname, upordown, bindspotid, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadContextComplexStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int) (value complex128, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetContextComplexStatus(contextname, upordown, bindspotid, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
	}
	return
}

func (t *Transaction) ReadContextStringStatus(area, id, contextname string, upordown spots.ContextUpDown, bindspotid string, bindbit int) (value string, have bool, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
		return
	}
	value, have, err = spotc.spot.GetContextStringStatus(contextname, upordown, bindspotid, bindbit)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContextStatus: %v", err)
	}
	return
}

func (t *Transaction) WriteContexts(area, id string, contexts map[string]spots.Context) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteContexts: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteContexts: %v", err)
		return
	}
	spotc.spot.SetContexts(contexts)
	spotc.be_change = true
	return
}

func (t *Transaction) ReadContexts(area, id string) (contexts map[string]spots.Context, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadContexts: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadContexts: %v", err)
		return
	}
	contexts = spotc.spot.GetContexts()
	return
}

func (t *Transaction) ResetContexts(area, id string) (err error) {
	contexts := make(map[string]spots.Context)
	return t.WriteContexts(area, id, contexts)
}

func (t *Transaction) WriteData(area, id, name string, data interface{}) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteData: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteData: %v", err)
		return
	}
	if spotc.spot.Body == nil {
		err = fmt.Errorf("trule[Transaction]WriteData: The data body not exist.")
		return
	}
	err = spotc.spot.Body.Set(name, data)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteData: %v", err)
		return
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteDataToBbody(area, id string, prototype BbodyMarshaler, name string, data interface{}) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteDataToBbody: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteDataToBbody: %v", err)
		return
	}
	bvalue, err := prototype.BbodyMarshel(name, data)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteDataToBbody: %v", err)
		return
	}
	err = spotc.spot.SetBbody(name, bvalue)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteDataToBbody: %v", err)
		return
	}
	spotc.be_change = true
	return
}

func (t *Transaction) WriteDataBytes(area, id string, name string, data []byte) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]WriteDataBytes: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	spotc, _, err := t.getSpot(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteDataBytes: %v", err)
		return
	}
	err = spotc.spot.SetBbody(name, data)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]WriteDataBytes: %v", err)
		return
	}
	spotc.be_change = true
	return
}

func (t *Transaction) ReadData(area, id, name string, data interface{}) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadData: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()

	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadData: %v", err)
		return
	}
	if spotc.spot.Body == nil {
		err = fmt.Errorf("trule[Transaction]WriteData: The data body not exist.")
		return
	}
	err = spotc.spot.Body.Get(name, data)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadData: %v", err)
	}
	return
}

func (t *Transaction) ReadDataFromBbody(area, id string, prototype BbodyMarshaler, name string, data interface{}) (err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadDataFromBbody: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()

	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadDataFromBbody: %v", err)
		return
	}
	bvalue, err := spotc.spot.GetBbody(name)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadDataFromBbody: %v", err)
		return
	}
	err = prototype.BbodyUnmarshaler(name, bvalue, data)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadDataFromBbody: %v", err)
		return
	}
	return
}

func (t *Transaction) ReadDataBytes(area, id string, name string) (data []byte, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("trule[Transaction]ReadDataBytes: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()

	spotc, _, err := t.getSpot(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadDataBytes: %v", err)
		return
	}
	data, err = spotc.spot.GetBbody(name)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ReadDataBytes: %v", err)
		return
	}
	return
}

func (t *Transaction) Commit() (err error) {
	transig := &transactionSig{
		ask: TRANSACTION_ASK_COMMIT,
		id:  t.id,
		re:  make(chan *transactionReturn),
	}
	t.tran_sig <- transig
	treturn := <-transig.re
	if treturn.status != TRAN_RETURN_HANDLE_OK {
		err = treturn.err
		return
	}
	return
}

func (t *Transaction) Rollback() (err error) {
	transig := &transactionSig{
		ask: TRANSACTION_ASK_ROLLBACK,
		id:  t.id,
		re:  make(chan *transactionReturn),
	}
	t.tran_sig <- transig
	treturn := <-transig.re
	if treturn.status != TRAN_RETURN_HANDLE_OK {
		err = treturn.err
		return
	}
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
