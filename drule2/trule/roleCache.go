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

// 初始化一个空的角色缓存
func initRoleC(area, id string) (rolec *roleCache) {
	rolec = &roleCache{
		area:           area,
		id:             id,
		exist:          false,
		be_delete:      TRAN_ROLE_BE_DELETE_NO,
		be_change:      false,
		tran_time:      time.Now(),
		wait_line:      make([]*tranAskGetRole, 0),
		wait_line_lock: new(sync.RWMutex),
		op_lock:        new(sync.RWMutex),
	}
	return
}

// 加入某角色缓存的等待队列
func (r *roleCache) addWait(wait *tranAskGetRole) {
	r.wait_line_lock.Lock()
	defer r.wait_line_lock.Unlock()
	r.wait_line = append(r.wait_line, wait)
}

// 查看能不能释放这个缓存
func (r *roleCache) canbeRelease() (can bool) {
	r.wait_line_lock.RLock()
	defer r.wait_line_lock.RUnlock()
	thelen := len(r.wait_line)
	if thelen != 0 {
		return false
	} else {
		return true
	}
}
