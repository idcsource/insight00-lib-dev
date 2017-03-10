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

// 处理事务的信号量
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

// 处理commit信号量
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
			// 将这个角色保存
			t.local_store.StoreRole(rolec.role)
			// 将这个角色从缓存中移除
			tran.tran_cache[roleid] = nil
			delete(tran.tran_cache, roleid)
		} else {
			fmt.Println("Tran log, 不写入 ", rolec.role.ReturnId())
			// 替换本尊
			rolec.role_store.body, rolec.role_store.rela, rolec.role_store.vers, _ = hardstore.EncodeRole(rolec.role)
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
	// 删除这个事务
	delete(t.transaction, signal.tran_id)
}

// 处理rollback信号量
func (t *TRule) handleRollbackSignal(signal *tranCommitSignal) {

}

func (t *TRule) StoreRole(role roles.Roleer) (err error) {
	err = t.local_store.StoreRole(role)
	return
}

func (t *TRule) ReadRole(id string) (role roles.Roleer, err error) {
	role, err = t.local_store.ReadRole(id)
	return
}

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
	}
	t.transaction[unid] = tran
	fmt.Println("Zr Log, New Tran: ", unid)
	return
}
