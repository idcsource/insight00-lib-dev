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

	"github.com/idcsource/insight00-lib/roles"
)

// 初始化一个空的角色缓存
func initRoleC(area, id string) (rolec *roleCache) {
	rolec = &roleCache{
		area:           area,
		id:             id,
		exist:          false,
		forwrite:       false,
		be_delete:      TRAN_ROLE_BE_DELETE_NO,
		be_change:      false,
		tran_time:      time.Now(),
		wait_line:      make([]*cacheAskRole, 0),
		wait_line_lock: new(sync.RWMutex),
		op_lock:        new(sync.RWMutex),
	}
	//go rolec.listen()
	return
}

// 排队信号监听
//func (r *roleCache) listen() {
//	for {
//		wait_sig := <-r.wait_line_sig
//		switch wait_sig.optype {
//		case ROLE_CACHE_ASK_GET:
//			// 请求获取
//			r.askToGet(wait_sig)
//		case ROLE_CACHE_ASK_RELEASE:
//			// 请求释放
//			r.askToRelease(wait_sig)
//		}
//	}
//}

// 处理请求获取的信号
func (r *roleCache) askToGet(wait *cacheAskRole) (approved bool) {
	r.wait_line_lock.Lock()
	defer r.wait_line_lock.Unlock()
	towait := false
	if wait.forwrite == true {
		if r.tran_id == "" {
			r.tran_id = wait.tran_id
			r.tran_time = time.Now()
			r.forwrite = true
		} else {
			towait = true
		}
	} else {
		if r.tran_id != "" || r.forwrite == true {
			towait = true
		} else {
			r.tran_time = time.Now()
		}
	}
	if towait == false {
		// 如果可以给予就发送
		approved = true
	} else {
		// 否则就加入等待，并返回false
		r.wait_line = append(r.wait_line, wait)
		approved = false
	}
	return
}

// 处理请求释放的信号，等待队列空就返回true，否则返回false
func (r *roleCache) askToRelease() (approved bool) {
	r.wait_line_lock.Lock()
	defer r.wait_line_lock.Unlock()

	r.tran_id = ""
	r.forwrite = false

	waitlen := len(r.wait_line)
	if waitlen == 0 {
		// 队列空，就发true给发出释放信号的家伙
		approved = true
	} else {
		thenext := r.wait_line[0]
		// 队列没空，就发false给发出释放信号的家伙
		approved = false
		if thenext.forwrite == true {
			r.tran_id = thenext.tran_id
			r.forwrite = true
		} else {
			r.tran_id = ""
			r.forwrite = false
		}
		r.tran_time = time.Now()

		if waitlen == 1 {
			r.wait_line = make([]*cacheAskRole, 0)
		} else {
			new_wait_line := r.wait_line[1:]
			r.wait_line = new_wait_line
		}
		thenext.approved <- true
	}
	return
}

// 设置角色
func (r *roleCache) setRole(role *roles.RoleMiddleData) {
	r.role = role
}

// 加入某角色缓存的等待队列
func (r *roleCache) addWait(wait *cacheAskRole) {
	r.wait_line = append(r.wait_line, wait)
}
