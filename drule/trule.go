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

// 创建TRule，事务统治者，需要依赖hadstore的本地存储
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

// 创建事务，并依据输入的角色ID进行准备，获取这些角色的写权限
func (t *TRule) prepareForDRule(unid string, roleids []string) (err error) {
	// 调用Begin()
	err = t.beginForDRule(unid)
	if err != nil {
		return err
	}
	// 调用tran中的prepare()
	err = t.transaction[unid].prepare(roleids)
	if err != nil {
		return err
	}
	return
}

// 由DRule来控制创建事务
func (t *TRule) beginForDRule(unid string) (err error) {
	t.tran_lock.Lock()
	defer t.tran_lock.Unlock()
	tran := &Transaction{
		unid:               unid,
		tran_cache:         make(map[string]*roleCache),
		tran_service:       t.tran_service,
		tran_time:          time.Now(),
		lock:               new(sync.RWMutex),
		tran_commit_signal: t.tran_commit_signal,
		be_delete:          false,
	}
	_, find := t.transaction[unid]
	if find == true {
		err = fmt.Errorf("The transaction is already exist, can't recreate : %v .", unid)
		return
	}
	t.transaction[unid] = tran
	fmt.Println("Zr Log, New Tran: ", unid)
	return
}

// 由DRule来控制的获取到事务
func (t *TRule) getTransactionForDRule(unid string) (tran *Transaction, err error) {
	var find bool
	tran, find = t.transaction[unid]
	if find == false {
		err = fmt.Errorf("Can not find transaction : %v .", unid)
		return
	}
	return
}

/* 下面是rolesio.RolesInOutManager接口的实现 */

// 往永久存储写入一个角色，直接调用底层HardStore存储（也就是说，直接使用这个是不安全的）
func (t *TRule) StoreRole(role roles.Roleer) (err error) {
	err = t.local_store.StoreRole(role)
	if err != nil {
		err = fmt.Errorf("drule[TRule]StoreRole: %v", err)
	}
	return
}

// 从永久存储读出一个角色，直接调用底层HardStore存储（也就是说，直接使用这个是不安全的）
func (t *TRule) ReadRole(id string) (role roles.Roleer, err error) {
	role, err = t.local_store.ReadRole(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadRole: %v", err)
	}
	return
}

// 从永久存储删除一个角色，直接调用底层HardStore存储（也就是说，直接使用这个是不安全的）
func (t *TRule) DeleteRole(id string) (err error) {
	err = t.local_store.DeleteRole(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]DeleteRole: %v", err)
	}
	return
}

// 设置父角色
func (t *TRule) WriteFather(id, father string) (err error) {
	tran := t.Begin()
	err = tran.WriteFather(id, father)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取父角色
func (t *TRule) ReadFather(id string) (father string, err error) {
	tran := t.Begin()
	father, err = tran.ReadFather(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置父角色
func (t *TRule) ResetFather(id string) (err error) {
	tran := t.Begin()
	err = tran.ResetFather(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ResetFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 读取所有角色的子角色
func (t *TRule) ReadChildren(id string) (children []string, err error) {
	tran := t.Begin()
	children, err = tran.ReadChildren(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入所有角色的子角色
func (t *TRule) WriteChildren(id string, children []string) (err error) {
	tran := t.Begin()
	err = tran.WriteChildren(id, children)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除所有子角色
func (t *TRule) ResetChildren(id string) (err error) {
	tran := t.Begin()
	err = tran.ResetChildren(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ResetChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入一个子角色
func (t *TRule) WriteChild(id, child string) (err error) {
	tran := t.Begin()
	err = tran.WriteChild(id, child)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个子角色
func (t *TRule) DeleteChild(id, child string) (err error) {
	tran := t.Begin()
	err = tran.DeleteChild(id, child)
	if err != nil {
		err = fmt.Errorf("drule[TRule]DeleteChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 查询是否有角色
func (t *TRule) ExistChild(id, child string) (have bool, err error) {
	tran := t.Begin()
	have, err = tran.ExistChild(id, child)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ExistChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 读取所有朋友关系
func (t *TRule) ReadFriends(id string) (friends map[string]roles.Status, err error) {
	tran := t.Begin()
	friends, err = tran.ReadFriends(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入所有朋友关系
func (t *TRule) WriteFriends(id string, friends map[string]roles.Status) (err error) {
	tran := t.Begin()
	err = tran.WriteFriends(id, friends)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置朋友关系
func (t *TRule) ResetFriends(id string) (err error) {
	tran := t.Begin()
	err = tran.ResetFriends(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ResetFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写一个朋友关系并绑定
func (t *TRule) WriteFriend(id, friend string, bind int64) (err error) {
	tran := t.Begin()
	err = tran.WriteFriend(id, friend, bind)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteFriend: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个朋友关系
func (t *TRule) DeleteFriend(id, friend string) (err error) {
	tran := t.Begin()
	err = tran.DeleteFriend(id, friend)
	if err != nil {
		err = fmt.Errorf("drule[TRule]DeleteFriend: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 创建一个空的上下文
func (t *TRule) CreateContext(id, contextname string) (err error) {
	tran := t.Begin()
	err = tran.CreateContext(id, contextname)
	if err != nil {
		err = fmt.Errorf("drule[TRule]CreateContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个上下文
func (t *TRule) DropContext(id, contextname string) (err error) {
	tran := t.Begin()
	err = tran.DropContext(id, contextname)
	if err != nil {
		err = fmt.Errorf("drule[TRule]DropContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回某个上下文全部的信息
func (t *TRule) ReadContext(id, contextname string) (context roles.Context, have bool, err error) {
	tran := t.Begin()
	context, have, err = tran.ReadContext(id, contextname)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个上下文绑定
func (t *TRule) DeleteContextBind(id, contextname string, upordown uint8, bindrole string) (err error) {
	tran := t.Begin()
	err = tran.DeleteContextBind(id, contextname, upordown, bindrole)
	if err != nil {
		err = fmt.Errorf("drule[TRule]DeleteContextBind: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回某个上下文的同样绑定值的所有
func (t *TRule) ReadContextSameBind(id, contextname string, upordown uint8, bind int64) (rolesid []string, have bool, err error) {
	tran := t.Begin()
	rolesid, have, err = tran.ReadContextSameBind(id, contextname, upordown, bind)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadContextSameBind: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回所有上下文组的名称
func (t *TRule) ReadContextsName(id string) (names []string, err error) {
	tran := t.Begin()
	names, err = tran.ReadContextsName(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadContextsName: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设置朋友的状态属性
func (t *TRule) WriteFriendStatus(id, friends string, bindbit int, value interface{}) (err error) {
	tran := t.Begin()
	err = tran.WriteFriendStatus(id, friends, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteFriendStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取朋友的状态属性
func (t *TRule) ReadFriendStatus(id, friends string, bindbit int, value interface{}) (err error) {
	tran := t.Begin()
	err = tran.ReadFriendStatus(id, friends, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadFriendStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设置上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
func (t *TRule) WriteContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	tran := t.Begin()
	err = tran.WriteContextStatus(id, contextname, upordown, bindroleid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteContextStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
func (t *TRule) ReadContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	tran := t.Begin()
	err = tran.ReadContextStatus(id, contextname, upordown, bindroleid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadContextStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设定上下文
func (t *TRule) WriteContexts(id string, context map[string]roles.Context) (err error) {
	tran := t.Begin()
	err = tran.WriteContexts(id, context)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取上下文
func (t *TRule) ReadContexts(id string) (contexts map[string]roles.Context, err error) {
	tran := t.Begin()
	contexts, err = tran.ReadContexts(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置上下文
func (t *TRule) ResetContexts(id string) (err error) {
	tran := t.Begin()
	err = tran.ResetContexts(id)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ResetContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 把data的数据装入role的name值下，如果找不到name，则返回错误。
func (t *TRule) WriteData(id, name string, data interface{}) (err error) {
	tran := t.Begin()
	err = tran.WriteData(id, name, data)
	if err != nil {
		err = fmt.Errorf("drule[TRule]WriteData: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

func (t *TRule) writeDataFromByte(id, name string, data []byte) (err error) {
	tran := t.Begin()
	err = tran.writeDataFromByte(id, name, data)
	if err != nil {
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 从角色中知道name的数据名并返回其数据。
func (t *TRule) ReadData(id, name string, data interface{}) (err error) {
	tran := t.Begin()
	err = tran.ReadData(id, name, data)
	if err != nil {
		err = fmt.Errorf("drule[TRule]ReadData: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

func (t *TRule) readDataToByte(role_data *Net_RoleData_Data) (err error) {
	tran := t.Begin()
	err = tran.readDataToByte(role_data)
	if err != nil {
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 运行时保存
func (t *TRule) ToStore() (err error) {
	err = fmt.Errorf("drule[TRule]ToStore: Transaction does not provide this method.")
	return
}
