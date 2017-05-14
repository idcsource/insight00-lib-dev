// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package trule

import (
	"fmt"
	"sync"
	"time"

	"github.com/idcsource/Insight-0-0-lib/drule2/hardstorage"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 创建TRule，事务统治者，需要依赖hadstore的本地存储
func NewTRule(local_store *hardstorage.HardStorage) (t *TRule, err error) {
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
		tran_timeout:       TRAN_TIME_OUT,
		tran_timeout_check: TRAN_TIME_OUT_CHECK,
		max_transaction:    TRAN_MAX_COUNT,
		pausing_signal:     make(chan bool),
		paused_signal:      make(chan bool),
		work_status:        TRULE_RUN_PAUSED,
		tran_wait:          &sync.WaitGroup{},
	}
	go t.tranSignalHandle()
	go t.tranTimeOutMonitor()
	go t.pauseSignalHandle()
	return
}

// 启动
func (t *TRule) Start() {
	t.work_status = TRULE_RUN_RUNNING
}

// 暂停
func (t *TRule) Pause() {
	t.work_status = TRULE_RUN_PAUSEING
	t.pausing_signal <- true
	// 开始等paused_signal
	<-t.paused_signal
	t.work_status = TRULE_RUN_PAUSED
	return
}

// 查看工作状态
func (t *TRule) WorkStatus() (status uint8) {
	return t.work_status
}

// 获取当前事务数
func (t *TRule) TransactionCount() (count int) {
	return t.count_transaction
}

func (t *TRule) pauseSignalHandle() {
	for {
		// 等待暂停中信号
		<-t.pausing_signal
		// 等待waiting的信号
		t.tran_wait.Wait()
		// 发送已经暂停信号
		t.paused_signal <- true
	}
}

// 处理事务超时的监控
func (t *TRule) tranTimeOutMonitor() {
	for {
		time.Sleep(time.Duration(t.tran_timeout_check) * time.Second)
		t.tranTimeOutMonitorToDo()
	}
}

// 实际是监测的tserver中的角色
func (t *TRule) tranTimeOutMonitorToDo() {
	t.tran_service.lock.Lock()
	t.tran_lock.Lock()
	defer t.tran_service.lock.Unlock()
	defer t.tran_lock.Unlock()

	for key, rolec := range t.tran_service.role_cache {
		del := t.tranTimeOutMonitorToOneRoleC(rolec)
		if del == true {
			delete(t.tran_service.role_cache, key)
		}
	}
}

func (t *TRule) tranTimeOutMonitorToOneRoleC(rolec *roleCache) (del bool) {
	del = false
	rolec.lock.Lock()
	defer rolec.lock.Unlock()
	wait_count := len(rolec.wait_line)
	if wait_count == 0 {
		// 如果没有在排队的，超时就延长10倍
		if rolec.tran_time.Unix()+(t.tran_timeout) > time.Now().Unix() {
			// 找到这个事务
			tran, find := t.transaction[rolec.tran_id]
			if find == false {
				// 如果事务已经不存在了怎么办，得强制释放
				rolec.tran_id = ""
				del = true
			} else {
				if tran.tran_time.Unix()+(t.tran_timeout) > time.Now().Unix() {
					// 强制回滚
					tran.Rollback()
				}
			}
		}
	} else {
		// 有排队的
		if rolec.tran_time.Unix()+t.tran_timeout > time.Now().Unix() {
			// 找到这个事务
			tran, find := t.transaction[rolec.tran_id]
			if find == false {
				// 如果事务已经不存在了怎么办，得强制释放占用呀
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
			} else {
				if tran.tran_time.Unix()+t.tran_timeout > time.Now().Unix() {
					// 强制回滚
					tran.Rollback()
				}
			}
		}
	}
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
	for cacheid, rolec := range tran.tran_cache {
		// 给这个角色加锁
		rolec.lock.Lock()

		// 检查等待队列
		wait_count := len(rolec.wait_line)
		if wait_count == 0 {
			fmt.Println("Tran log, 写入 ", rolec.role.ReturnId())
			// 如果没有等待队列
			// 将这个角色保存或删除
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_NO {
				if rolec.be_change == true {
					t.local_store.RoleStoreMiddleData(rolec.area, *rolec.role)
					rolec.be_change = false
				}
				rolec.tran_id = ""
			} else if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				t.local_store.RoleDelete(rolec.area, rolec.role.Version.Id)
				// 将这个角色从缓存中移除
				t.tran_service.role_cache[cacheid] = nil
				delete(t.tran_service.role_cache, cacheid)
			}
			rolec.lock.Unlock()
		} else {
			alreadyhave := true
			// 替换本尊或删除
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_NO {
				rolec.role_store = *rolec.role
				if rolec.be_change == true {
					t.local_store.RoleStoreMiddleData(rolec.area, *rolec.role)
					rolec.be_change = false
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
				wait_first.approved <- alreadyhave
				// 给这个角色解锁
				rolec.lock.Unlock()
			} else if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				// 如果删除了怎么办
				t.local_store.RoleDelete(rolec.area, rolec.role.Version.Id)
				rolec.role = nil
				rolec.tran_id = ""
				rolec.be_delete = TRAN_ROLE_BE_DELETE_COMMIT
				alreadyhave = false
				// 向所有排队发送已经被删除的状态
				for _, wait := range rolec.wait_line {
					wait.approved <- alreadyhave
				}
				rolec.lock.Unlock()
				delete(t.tran_service.role_cache, cacheid)
			}
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
	t.count_transaction--
	t.tran_wait.Done()
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
				midn := rolec.role_store
				rolec.role = &midn
			}
			// 重置删除标记
			if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES {
				rolec.be_delete = TRAN_ROLE_BE_DELETE_NO
			}
			if rolec.be_change == true {
				rolec.be_change = false
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
	t.count_transaction--
	t.tran_wait.Done()
}

// 创建事务
func (t *TRule) Begin() (tran *Transaction, err error) {
	if t.work_status != TRULE_RUN_RUNNING {
		err = fmt.Errorf("trule[TRule]Begin: The TRule is paused.")
		return
	}
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
	t.count_transaction++
	t.tran_wait.Add(1)
	return
}

// 创建事务，Begin的别名
func (t *TRule) Transcation() (tan *Transaction, err error) {
	return t.Begin()
}

// 创建事务，并依据输入的角色ID进行准备，获取这些角色的写权限
func (t *TRule) Prepare(area string, roleids ...string) (tran *Transaction, err error) {
	// 调用Begin()
	tran, _ = t.Begin()
	// 调用tran中的prepare()
	err = tran.prepare(area, roleids)
	if err != nil {
		err = fmt.Errorf("trule[TRule]Prepare: %v", err)
		tran.Rollback()
	}
	return
}

// 是否存在这个角色
func (t *TRule) ExistRole(area, id string) (have bool) {
	have = t.local_store.RoleExist(area, id)
	return
}

/* 下面是rolesio.RolesInOutManager接口的实现 */

// 往永久存储写入一个角色
func (t *TRule) StoreRole(area string, role roles.Roleer) (err error) {
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("trule[TRule]StoreRole: %v", err)
		return
	}
	err = t.local_store.RoleStoreMiddleData(area, mid)
	if err != nil {
		err = fmt.Errorf("trule[TRule]StoreRole: %v", err)
	}
	return
}

func (t *TRule) StoreRoleFromMiddleData(area string, mid roles.RoleMiddleData) (err error) {
	err = t.local_store.RoleStoreMiddleData(area, mid)
	if err != nil {
		err = fmt.Errorf("trule[TRule]StoreRole: %v", err)
	}
	return
}

// 从永久存储读出一个角色
func (t *TRule) ReadRole(area, id string, role roles.Roleer) (err error) {
	mid, err := t.local_store.RoleReadMiddleData(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]RoleRead: %v", err)
		return
	}
	// 转码
	err = roles.DecodeMiddleToRole(mid, role)
	if err != nil {
		err = fmt.Errorf("trule[TRule]RoleRead: %v", err)
	}
	return
}

// 从永久存储读出角色的MiddleData格式
func (t *TRule) ReadRoleMiddleData(area, id string) (mid roles.RoleMiddleData, err error) {
	mid, err = t.local_store.RoleReadMiddleData(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]RoleReadMiddleData: %v", err)
	}
	return
}

// 从永久存储删除一个角色，直接调用底层HardStore存储（也就是说，直接使用这个是不安全的）
func (t *TRule) DeleteRole(area, id string) (err error) {
	err = t.local_store.RoleDelete(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]DeleteRole: %v", err)
	}
	return
}

// 删除区域
func (t *TRule) AreaDelete(area string) (err error) {
	err = t.local_store.AreaDelete(area)
	if err != nil {
		err = fmt.Errorf("trule[TRule]AreaDelete: %v", err)
	}
	return
}

// 初始化区域
func (t *TRule) AreaInit(area string) (err error) {
	err = t.local_store.AreaInit(area)
	if err != nil {
		err = fmt.Errorf("trule[TRule]AreaInit: %v", err)
	}
	return
}

// 区域是否存在
func (t *TRule) AreaExist(area string) (have bool) {
	return t.local_store.AreaExist(area)
}

// 区域列表
func (t *TRule) AreaList() (list []string, err error) {
	return t.local_store.AreaList()
}

// 区域改名
func (t *TRule) AreaReName(oldname, newname string) (err error) {
	err = t.local_store.AreaReName(oldname, newname)
	if err != nil {
		err = fmt.Errorf("trule[TRule]AreaReName: %v", err)
	}
	return
}

// 设置父角色
func (t *TRule) WriteFather(area, id, father string) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteFather(area, id, father)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取父角色
func (t *TRule) ReadFather(area, id string) (father string, err error) {
	tran, _ := t.Begin()
	father, err = tran.ReadFather(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置父角色
func (t *TRule) ResetFather(area, id string) (err error) {
	tran, _ := t.Begin()
	err = tran.ResetFather(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ResetFather: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 读取所有角色的子角色
func (t *TRule) ReadChildren(area, id string) (children []string, err error) {
	tran, _ := t.Begin()
	children, err = tran.ReadChildren(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入所有角色的子角色
func (t *TRule) WriteChildren(area, id string, children []string) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteChildren(area, id, children)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除所有子角色
func (t *TRule) ResetChildren(area, id string) (err error) {
	tran, _ := t.Begin()
	err = tran.ResetChildren(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ResetChildren: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入一个子角色
func (t *TRule) WriteChild(area, id, child string) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteChild(area, id, child)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个子角色
func (t *TRule) DeleteChild(area, id, child string) (err error) {
	tran, _ := t.Begin()
	err = tran.DeleteChild(area, id, child)
	if err != nil {
		err = fmt.Errorf("trule[TRule]DeleteChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 查询是否有角色
func (t *TRule) ExistChild(area, id, child string) (have bool, err error) {
	tran, _ := t.Begin()
	have, err = tran.ExistChild(area, id, child)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ExistChild: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 读取所有朋友关系
func (t *TRule) ReadFriends(area, id string) (friends map[string]roles.Status, err error) {
	tran, _ := t.Begin()
	friends, err = tran.ReadFriends(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写入所有朋友关系
func (t *TRule) WriteFriends(area, id string, friends map[string]roles.Status) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteFriends(area, id, friends)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置朋友关系
func (t *TRule) ResetFriends(area, id string) (err error) {
	tran, _ := t.Begin()
	err = tran.ResetFriends(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ResetFriends: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 写一个朋友关系并绑定
func (t *TRule) WriteFriend(area, id, friend string, bind int64) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteFriend(area, id, friend, bind)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteFriend: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个朋友关系
func (t *TRule) DeleteFriend(area, id, friend string) (err error) {
	tran, _ := t.Begin()
	err = tran.DeleteFriend(area, id, friend)
	if err != nil {
		err = fmt.Errorf("trule[TRule]DeleteFriend: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 创建一个空的上下文
func (t *TRule) CreateContext(area, id, contextname string) (err error) {
	tran, _ := t.Begin()
	err = tran.CreateContext(area, id, contextname)
	if err != nil {
		err = fmt.Errorf("trule[TRule]CreateContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 是否有这个上下文
func (t *TRule) ExistContext(area, id, contextname string) (have bool, err error) {
	tran, _ := t.Begin()
	have, err = tran.ExistContext(area, id, contextname)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ExistContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个上下文
func (t *TRule) DropContext(area, id, contextname string) (err error) {
	tran, _ := t.Begin()
	err = tran.DropContext(area, id, contextname)
	if err != nil {
		err = fmt.Errorf("trule[TRule]DropContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回某个上下文全部的信息
func (t *TRule) ReadContext(area, id, contextname string) (context roles.Context, have bool, err error) {
	tran, _ := t.Begin()
	context, have, err = tran.ReadContext(area, id, contextname)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadContext: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 删除一个上下文绑定
func (t *TRule) DeleteContextBind(area, id, contextname string, upordown roles.ContextUpDown, bindrole string) (err error) {
	tran, _ := t.Begin()
	err = tran.DeleteContextBind(area, id, contextname, upordown, bindrole)
	if err != nil {
		err = fmt.Errorf("trule[TRule]DeleteContextBind: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回某个上下文的同样绑定值的所有
func (t *TRule) ReadContextSameBind(area, id, contextname string, upordown roles.ContextUpDown, bind int64) (rolesid []string, have bool, err error) {
	tran, _ := t.Begin()
	rolesid, have, err = tran.ReadContextSameBind(area, id, contextname, upordown, bind)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadContextSameBind: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 返回所有上下文组的名称
func (t *TRule) ReadContextsName(area, id string) (names []string, err error) {
	tran, _ := t.Begin()
	names, err = tran.ReadContextsName(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadContextsName: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设置朋友的状态属性
func (t *TRule) WriteFriendStatus(area, id, friends string, bindbit int, value interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteFriendStatus(area, id, friends, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteFriendStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取朋友的状态属性
func (t *TRule) ReadFriendStatus(area, id, friends string, bindbit int, value interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.ReadFriendStatus(area, id, friends, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadFriendStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设置上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
func (t *TRule) WriteContextStatus(area, id, contextname string, upordown roles.ContextUpDown, bindroleid string, bindbit int, value interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteContextStatus(area, id, contextname, upordown, bindroleid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteContextStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
func (t *TRule) ReadContextStatus(area, id, contextname string, upordown roles.ContextUpDown, bindroleid string, bindbit int, value interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.ReadContextStatus(area, id, contextname, upordown, bindroleid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadContextStatus: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 设定上下文
func (t *TRule) WriteContexts(area, id string, context map[string]roles.Context) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteContexts(area, id, context)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 获取上下文
func (t *TRule) ReadContexts(area, id string) (contexts map[string]roles.Context, err error) {
	tran, _ := t.Begin()
	contexts, err = tran.ReadContexts(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 重置上下文
func (t *TRule) ResetContexts(area, id string) (err error) {
	tran, _ := t.Begin()
	err = tran.ResetContexts(area, id)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ResetContexts: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 把data的数据装入role的name值下，如果找不到name，则返回错误。
func (t *TRule) WriteData(area, id, name string, data interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteData(area, id, name, data)
	if err != nil {
		err = fmt.Errorf("trule[TRule]WriteData: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

func (t *TRule) WriteDataFromByte(area, id, name string, data []byte) (err error) {
	tran, _ := t.Begin()
	err = tran.WriteDataFromByte(area, id, name, data)
	if err != nil {
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

// 从角色中知道name的数据名并返回其数据。
func (t *TRule) ReadData(area, id, name string, data interface{}) (err error) {
	tran, _ := t.Begin()
	err = tran.ReadData(area, id, name, data)
	if err != nil {
		err = fmt.Errorf("trule[TRule]ReadData: %v", err)
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}

func (t *TRule) ReadDataToByte(area, id, name string) (data []byte, err error) {
	tran, _ := t.Begin()
	data, err = tran.ReadDataToByte(area, id, name)
	if err != nil {
		tran.Rollback()
		return
	}
	tran.Commit()
	return
}
