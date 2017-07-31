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

	"github.com/idcsource/insight00-lib/roles"
)

// 初始化一个transaction
func initTransaction(id string, tranSig chan *transactionSig, rolecache_sig chan *roleCacheSig) (t *Transaction) {
	t = &Transaction{
		id:              id,
		role_cache:      make(map[string]map[string]*roleCache),
		role_cache_name: make(map[string]map[string]bool),
		role_cache_lock: new(sync.RWMutex),
		tran_time:       time.Now(),
		role_cache_sig:  rolecache_sig,
		tran_sig:        tranSig,
		be_delete:       false,
	}
	return
}

// If the Role exist.
func (t *Transaction) ExistRole(area, id string) (exist bool, err error) {
	if t.be_delete == true {
		return false, fmt.Errorf("trule[Transaction]ExistRole: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	_, exist, err = t.getRole(area, id, false)
	if err != nil {
		err = fmt.Errorf("trule[Transaction]ExistRole: %v", err)
	}
	return
}

// Read a Role
func (t *Transaction) ReadRole(area, id string, role roles.Roleer) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("trule[Transaction]ReadRole: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	rolec, exist, err := t.getRole(area, id, true)
	if err != nil {
		err = fmt.Errorf("trule[Transacion]ReadRole: %v", err)
		return
	}
	if exist == false {
		err = fmt.Errorf("trule[Transacion]ReadRole: The Role not exist.")
		return
	}
	err = roles.DecodeMiddleToRole(rolec.role, role)
	if err != nil {
		err = fmt.Errorf("trule[Transacion]ReadRole: %v", err)
	}
	return
}

// Read Role's MiddleData
func (t *Transaction) ReadRoleMiddleData(area, id string) (mid *roles.RoleMiddleData, err error) {
	if t.be_delete == true {
		err = fmt.Errorf("drule[Transaction]ReadRoleMiddleData: This transaction has been deleted.")
		return
	}
	t.tran_time = time.Now()
	rolec, exist, err := t.getRole(area, id, true)
	if err != nil {
		err = fmt.Errorf("drule[Transacion]ReadRoleMiddleData: %v", err)
		return
	}
	if exist == false {
		err = fmt.Errorf("trule[Transacion]ReadRoleMiddleData: The Role not exist.")
		return
	}
	mid = rolec.role
	return
}

// Store Role
func (t *Transaction) StoreRole(area string, role roles.Roleer) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]StoreRole: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	//	t.lock.RLock()
	//	defer t.lock.RUnlock()
	roleid := role.ReturnId()
	rolec, err := t.getRoleForNew(area, roleid)
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		return fmt.Errorf("drule[Transaction]StoreRole: %v", err)
	}
	rolec.role = mid
	rolec.be_change = true
	return nil
}

// Store Role from middle data type
func (t *Transaction) StoreRoleFromMiddleData(area string, mid *roles.RoleMiddleData) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]StoreRoleFromMiddleData: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	//	t.lock.RLock()
	//	defer t.lock.RUnlock()
	roleid := mid.ReturnId()
	rolec, err := t.getRoleForNew(area, roleid)
	rolec.role = mid
	rolec.be_change = true
	return nil
}

// Delete a Role
func (t *Transaction) DeleteRole(area, id string) (err error) {
	if t.be_delete == true {
		return fmt.Errorf("drule[Transaction]DeleteRole: This transaction has been deleted.")
	}
	t.tran_time = time.Now()
	rolec, exist, err := t.getRole(area, id, true)
	if err != nil {
		return err
	}
	if exist == false {
		return fmt.Errorf("drule[Transaction]DeleteRole: The Role not exist.")
	}
	rolec.be_delete = TRAN_ROLE_BE_DELETE_YES
	rolec.be_change = true
	return nil
}

// 获取一个角色，forwrite是true就是为了写
func (t *Transaction) getRole(area, id string, forwrite bool) (rolec *roleCache, exist bool, err error) {
	// 构建信号
	role_cache_sig := &roleCacheSig{
		ask:      ROLE_CACHE_ASK_GET,
		area:     area,
		id:       id,
		tranid:   t.id,
		forwrite: forwrite,
		ask_time: time.Now(),
		re:       make(chan *roleCacheReturn),
	}
	// 发送信号
	t.role_cache_sig <- role_cache_sig
	// 等待返回
	sigre := <-role_cache_sig.re
	if sigre.status != ROLE_CACHE_RETURN_HANDLE_OK {
		err = sigre.err
		return
	}
	rolec = sigre.role
	exist = sigre.exist
	// 如果是为了写，就加入自己的缓存
	if forwrite == true {
		t.role_cache_lock.Lock()
		if _, have := t.role_cache[area]; have == false {
			t.role_cache[area] = make(map[string]*roleCache)
			t.role_cache_name[area] = make(map[string]bool)
		}
		t.role_cache[area][id] = rolec
		t.role_cache_name[area][id] = true
		t.role_cache_lock.Unlock()
	}
	return
}

// 获取一个角色（有可能是空的，这通常是用来新建一个角色），forwrite是true就是为了写
func (t *Transaction) getRoleForNew(area, id string) (rolec *roleCache, err error) {
	// 构建信号
	role_cache_sig := &roleCacheSig{
		ask:      ROLE_CACHE_ASK_WRITE,
		area:     area,
		id:       id,
		tranid:   t.id,
		forwrite: true,
		ask_time: time.Now(),
		re:       make(chan *roleCacheReturn),
	}
	// 发送信号
	t.role_cache_sig <- role_cache_sig
	// 等待返回
	sigre := <-role_cache_sig.re
	if sigre.status != ROLE_CACHE_RETURN_HANDLE_OK {
		err = sigre.err
		return
	}
	rolec = sigre.role
	// 加入自己的缓存
	t.role_cache_lock.Lock()
	if _, have := t.role_cache[area]; have == false {
		t.role_cache[area] = make(map[string]*roleCache)
		t.role_cache_name[area] = make(map[string]bool)
	}
	t.role_cache[area][id] = rolec
	t.role_cache_name[area][id] = true
	t.role_cache_lock.Unlock()
	return
}
