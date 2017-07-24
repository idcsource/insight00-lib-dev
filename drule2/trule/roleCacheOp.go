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

	"github.com/idcsource/insight00-lib/drule2/hardstorage"
	"github.com/idcsource/insight00-lib/ilogs"
)

// 初始化roleCacheOp
func initRoleCacheOp(local_store *hardstorage.HardStorage, log *ilogs.Logs) (rco *roleCacheOp) {
	rco = &roleCacheOp{
		local_store: local_store,
		signal:      make(chan *roleCacheSig, ROLE_CACHE_SIGNAL_CHANNEL_LEN),
		cache:       make(map[string]map[string]*roleCache),
		clean_count: 0,
		log:         log,
		closed:      true,
		closesig:    make(chan bool),
	}
	return
}

// Start
func (rco *roleCacheOp) Start() {
	rco.closed = false
	go rco.listen()
}

// Stop
func (rco *roleCacheOp) Stop() {
	rco.closesig <- true
	rco.consCleanSig()
	rco.closed = true
}

// listen the role cache operate signal
func (rco *roleCacheOp) listen() {
	for {
		if rco.closed == true {
			return
		}
		select {
		case signal := <-rco.signal:
			rco.doSignal(signal)
		case closesig := <-rco.closesig:
			if closesig == true {
				return
			}
		}
	}
}

// 构建清理的信号
func (rco *roleCacheOp) consCleanSig() {
	signal := &roleCacheSig{
		ask: ROLE_CACHE_ASK_CLEAN,
		re:  make(chan *roleCacheReturn),
	}
	rco.signal <- signal
	re := <-signal.re
	if re.status == ROLE_CACHE_RETURN_HANDLE_ERROR {
		rco.log.ErrLog(re.err)
	}
	rco.clean_count = 0
}

// do signal
func (rco *roleCacheOp) doSignal(signal *roleCacheSig) {
	switch signal.ask {
	case ROLE_CACHE_ASK_GET:
		// ask get a role
		rco.askGetRole(signal, false)
	case ROLE_CACHE_ASK_WRITE:
		// ask write a role，两者的区别只在对角色不存在的错误处理上
		rco.askGetRole(signal, true)
	case ROLE_CACHE_ASK_STORE:
		// ask to store the role's change, and to next wait, for transaction commit
		rco.askStoreRole(signal)
	case ROLE_CACHE_ASK_RESET:
		// ask to reset the role, and to next wait, for transaction rollback
		rco.askResetRole(signal)
	case ROLE_CACHE_ASK_CLEAN:
		rco.askCleanRoles(signal)
	default:
		rco.log.ErrLog("Role Cache Signal's ask does not exist.")
	}
}

// ask to reset the role, and to next wait, for transaction rollback
func (rco *roleCacheOp) askResetRole(signal *roleCacheSig) {
	rolec := rco.cache[signal.area][signal.id]
	go rco.resetTheRole(signal, rolec)
}

// gro to reset role
func (rco *roleCacheOp) resetTheRole(signal *roleCacheSig, rolec *roleCache) {
	// 加锁
	rolec.op_lock.Lock()
	// 延时解锁
	defer rolec.op_lock.Unlock()
	// 修正到初始状态
	if rolec.be_change == true {
		if rolec.exist == false {
			rolec.role = nil
		} else {
			var err error
			rolec.role, _, err = rco.local_store.RoleReadMiddleData(rolec.area, rolec.id)
			if err != nil {
				rco.log.ErrLog(err)
			}
		}
		rolec.be_change = false
	}
	rolec.be_delete = TRAN_ROLE_BE_DELETE_NO
	// 释放当前
	rolec.askToRelease()
}

// ask to store the role's change, and to next wait
func (rco *roleCacheOp) askStoreRole(signal *roleCacheSig) {
	rolec := rco.cache[signal.area][signal.id]
	go rco.storeTheRole(signal, rolec)
}

// go store the role's change
func (rco *roleCacheOp) storeTheRole(signal *roleCacheSig, rolec *roleCache) {
	// 加锁
	rolec.op_lock.Lock()
	// 延时解锁
	defer rolec.op_lock.Unlock()
	// 保存
	if rolec.be_change == true && rolec.be_delete == TRAN_ROLE_BE_DELETE_NO {
		err := rco.local_store.RoleStoreMiddleData(rolec.area, rolec.role)
		if err != nil {
			rco.log.ErrLog(err)
		}
	}
	// 或真正删除
	if rolec.be_delete == TRAN_ROLE_BE_DELETE_YES || rolec.be_delete == TRAN_ROLE_BE_DELETE_COMMIT {
		err := rco.local_store.RoleDelete(rolec.area, rolec.id)
		if err != nil {
			rco.log.ErrLog(err)
		}
		rolec.be_delete = TRAN_ROLE_BE_DELETE_COMMIT
	}
	// 释放
	rolec.askToRelease()
}

// ask get a role
func (rco *roleCacheOp) askGetRole(signal *roleCacheSig, write bool) {
	// 在缓存中找到或分配位置
	_, havearea := rco.cache[signal.area]
	if havearea == false {
		rco.cache[signal.area] = make(map[string]*roleCache)
	}
	_, rolehave := rco.cache[signal.area][signal.id]
	if rolehave == false {
		rco.cache[signal.area][signal.id] = initRoleC(signal.area, signal.id)
	}
	// 交给协程去处理，但要先加锁，getRoleFromStorage中要释放这个锁
	go rco.getRoleFromStorage(signal, rco.cache[signal.area][signal.id], write)
}

// get role from hardstorage
func (rco *roleCacheOp) getRoleFromStorage(signal *roleCacheSig, rolec *roleCache, write bool) {
	rolec.op_lock.Lock()

	re := &roleCacheReturn{}
	if rolec.exist == false && rolec.role == nil {
		// 如果确定是没有
		therole, exist, err := rco.local_store.RoleReadMiddleData(signal.area, signal.id)
		if err != nil {
			if write == false {
				re.err = err
				re.status = ROLE_CACHE_RETURN_HANDLE_ERROR
			} else {
				re.status = ROLE_CACHE_RETURN_HANDLE_OK
				rco.clean_count++
			}
		} else if exist == false {
			if write == false {
				re.err = fmt.Errorf("The Role not exist.")
				re.status = ROLE_CACHE_RETURN_HANDLE_ERROR
			} else {
				re.status = ROLE_CACHE_RETURN_HANDLE_OK
				rco.clean_count++
			}
		} else {
			re.status = ROLE_CACHE_RETURN_HANDLE_OK
			rolec.role = therole
			rolec.exist = true
			re.role = rolec
			rco.clean_count++
		}
	} else {
		re.status = ROLE_CACHE_RETURN_HANDLE_OK
		re.role = rolec
	}
	// 解锁这个缓存
	rolec.op_lock.Unlock()
	// 尝试排队
	askrole := &cacheAskRole{
		tran_id:  signal.tranid,
		forwrite: signal.forwrite,
		approved: make(chan bool),
		ask_time: signal.ask_time,
	}
	approved := rolec.askToGet(askrole)
	if approved == false {
		// 进入排队的话就去监听等待
		<-askrole.approved
	}

	// 发送这个re
	signal.re <- re
}

// ask clean roles
func (rco *roleCacheOp) askCleanRoles(signal *roleCacheSig) {
	// 这里不能有协程了，也就是说在这个执行完所有的请求角色的都要等了
	tmpa := make(map[string][]string)
	for areaname, _ := range rco.cache {
		tmpa[areaname] = make([]string, 0)
		for rolename, rolec := range rco.cache[areaname] {
			if rolec.tran_id == "" && rolec.tran_time.Unix()+ROLE_CACHE_CLEAN_TIME_OUT < time.Now().Unix() {
				tmpa[areaname] = append(tmpa[areaname], rolename)
			}
		}
	}
	for areaname, _ := range tmpa {
		for _, rolename := range tmpa[areaname] {
			delete(rco.cache[areaname], rolename)
		}
	}
}
