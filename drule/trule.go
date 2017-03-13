// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"sync"
	"time"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 创建TRule，事务统治者，需要伊俄hadstore的本地存储
func NewTRule(local_store *hardstore.HardStore) (t *TRule, err error) {
	trans := &tranService{
		role_cache:  make(map[string]*roleCache),
		lock:        new(sync.RWMutex),
		local_store: local_store,
	}
	t = &TRule{
		local_store:        local_store,
		transaction:        make(map[string]*Transaction),
		tran_service:       trans,
		tran_lock:          new(sync.RWMutex),
		tran_commit_signal: make(chan *tranCommitSignal),
	}
	go t.tranSignalHandle()
	return
}

// 处理事务的信号
func (t *TRule) tranSignalHandle() {
	for {
		signal := <-t.tran_commit_signal
		switch signal.ask {
		case TRAN_COMMIT_ASK_COMMIT:
			t.handleCommitSignal(signal)
		case TRAN_COMMIT_ASK_ROLLBACK:
			t.handleRollbackSignal(signal)
		default:
		}
	}
}

// 处理commit信号
func (t *TRule) handleCommitSignal(signal *tranCommitSignal) {
	fmt.Println("Tran log, 正在执行 ", signal.tran_id)
	// 给事务加锁
	t.tran_lock.Lock()
	defer t.tran_lock.Unlock()

	// 构造返回
	returnHan := tranReturnHandle{}

	//找到这个事务
	tran, find := t.transaction[signal.tran_id]
	if find == false {
		// 如果找不到
		returnHan.Status = TRAN_RETURN_HANDLE_ERROR
		returnHan.Error = fmt.Errorf("Can't find the transaction %v.", signal.tran_id)
		signal.return_handle <- returnHan
		return
	}
	// 开始大量的执行
	// 给这个事务本身加锁
	tran.lock.Lock()
	defer tran.lock.Unlock()
	// 遍历这里面所有的缓存角色
	for roleid, rolec := range tran.tran_cache {
		// 给这个角色加锁
		rolec.lock.Lock()

		// 检查等待队列
		wait_count := len(rolec.wait_line)
		if wait_count == 0 {
			fmt.Println("Tran log, 写入 ", rolec.role.ReturnId())
			// 如果没有等待队列
			// 将这个角色保存或删除
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_NO {
				t.local_store.StoreRole(rolec.role)
			} else if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				t.local_store.DeleteRole(rolec.role.ReturnId())
			}
			// 将这个角色从缓存中移除
			t.tran_service.role_cache[roleid] = nil
			delete(t.tran_service.role_cache, roleid)
		} else {
			fmt.Println("Tran log, 不写入 ", rolec.role.ReturnId())
			// 替换本尊或删除
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_NO {
				rolec.role_store.body, rolec.role_store.rela, rolec.role_store.vers, _ = hardstore.EncodeRole(rolec.role)
			} else if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				t.local_store.DeleteRole(rolec.role.ReturnId())
				rolec.role = nil
				rolec.be_delete = TRAN_ROLE_BE_DELETE_COMMIT
			}
			// 删除占用标记
			rolec.tran_id = ""
			// 得到第一个等待的队列
			wait_first := rolec.wait_line[0]
			// 构造新的waitline
			if wait_count == 1 {
				rolec.wait_line = make([]*tranAskGetRole, 0)
			} else {
				new_wait_line := make([]*tranAskGetRole, 0)
				new_wait_line = rolec.wait_line[1:]
				rolec.wait_line = new_wait_line
			}
			// 修改占用标记
			rolec.tran_id = wait_first.tran_id
			// 修改占用时间
			rolec.tran_time = time.Now()
			// 发送允许的信息，接收者要自行判断是否被删除
			wait_first.approved <- true
			// 给这个角色解锁
			rolec.lock.Unlock()
		}
	}
	// 发送回执
	returnHan.Status = TRAN_RETURN_HANDLE_OK
	signal.return_handle <- returnHan
	// 删除这个事务
	tran.tran_service = nil
	tran.tran_cache = nil
	tran.tran_commit_signal = nil
	tran.be_delete = true
	t.transaction[signal.tran_id] = nil
	delete(t.transaction, signal.tran_id)
}

// 处理rollback信号
func (t *TRule) handleRollbackSignal(signal *tranCommitSignal) {
	// 给事务加锁
	t.tran_lock.Lock()
	defer t.tran_lock.Unlock()

	// 构造返回
	returnHan := tranReturnHandle{}
	//找到这个事务
	tran, find := t.transaction[signal.tran_id]
	if find == false {
		// 如果找不到
		returnHan.Status = TRAN_RETURN_HANDLE_ERROR
		returnHan.Error = fmt.Errorf("Can't find the transaction %v.", signal.tran_id)
		signal.return_handle <- returnHan
		return
	}
	// 开始大量的执行
	// 给这个事务本身加锁
	tran.lock.Lock()
	defer tran.lock.Unlock()
	// 遍历这里面所有的缓存角色
	for roleid, rolec := range tran.tran_cache {
		// 给这个角色加锁
		rolec.lock.Lock()

		// 检查等待队列
		wait_count := len(rolec.wait_line)
		if wait_count == 0 {
			// 如果没有等待队列，将这个角色从总缓存中移除
			t.tran_service.role_cache[roleid] = nil
			delete(t.tran_service.role_cache, roleid)
		} else {
			// 如果有排队的
			// 用本尊替换实体
			if rolec.be_delete != TRAN_ROLE_BE_DELETE_COMMIT {
				rolec.role, _ = hardstore.DecodeRole(rolec.role_store.body, rolec.role_store.rela, rolec.role_store.vers)
			}
			// 重置删除标记
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				rolec.be_delete = TRAN_ROLE_BE_DELETE_NO
			}
			// 删除占用标记
			rolec.tran_id = ""
			// 得到第一个等待的队列
			wait_first := rolec.wait_line[0]
			// 构造新的waitline
			if wait_count == 1 {
				rolec.wait_line = make([]*tranAskGetRole, 0)
			} else {
				new_wait_line := make([]*tranAskGetRole, 0)
				new_wait_line = rolec.wait_line[1:]
				rolec.wait_line = new_wait_line
			}
			// 修改占用标记
			rolec.tran_id = wait_first.tran_id
			// 发送允许的信息
			wait_first.approved <- true
			// 给这个角色解锁
			rolec.lock.Unlock()
		}
	}
	// 发送回执
	returnHan.Status = TRAN_RETURN_HANDLE_OK
	signal.return_handle <- returnHan
	// 删除这个事务
	tran.tran_service = nil
	tran.tran_cache = nil
	tran.tran_commit_signal = nil
	t.transaction[signal.tran_id] = nil
	tran.be_delete = true
	delete(t.transaction, signal.tran_id)
}

// 往永久存储写入一个角色
func (t *TRule) StoreRole(role roles.Roleer) (err error) {
	err = t.local_store.StoreRole(role)
	return
}

// 从永久存储读出一个角色
func (t *TRule) ReadRole(id string) (role roles.Roleer, err error) {
	role, err = t.local_store.ReadRole(id)
	return
}

// 创建事务
func (t *TRule) Begin() (tran *Transaction) {
	t.tran_lock.Lock()
	defer t.tran_lock.Unlock()
	unid := random.GetRand(40)
	tran = &Transaction{
		unid:               unid,
		tran_cache:         make(map[string]*roleCache),
		tran_service:       t.tran_service,
		tran_time:          time.Now(),
		lock:               new(sync.RWMutex),
		tran_commit_signal: t.tran_commit_signal,
		be_delete:          false,
	}
	t.transaction[unid] = tran
	fmt.Println("Zr Log, New Tran: ", unid)
	return
}

// 创建事务，Begin的别名
func (t *TRule) Transcation() (tan *Transaction) {
	return t.Begin()
}

// 创建事务，并依据输入的角色ID进行准备，获取这些角色的写权限
func (t *TRule) Prepare(roleids ...string) (tran *Transaction, err error) {
	// 调用Begin()
	tran = t.Begin()
	// 调用tran中的prepare()
	err = tran.prepare(roleids)
	if err != nil {
		err = fmt.Errorf("drule[TRule]Prepare: %v", err)
	}
	return
}
