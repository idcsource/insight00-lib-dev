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

	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 读取角色，lockmode为TRAN_LOCK_MODE_*
func (t *tranService) getRole(tran_id, area, id string, lockmode uint8) (rolec *roleCache, err error) {
	t.lock.Lock()
	//defer t.lock.Unlock()
	cache_id := area + id
	var find bool
	rolec, find = t.role_cache[cache_id]

	if find == false {
		// 找不到就从硬盘上读取角色
		mid, err := t.local_store.RoleReadMiddleData(area, id)
		if err != nil {
			t.lock.Unlock()
			return nil, err
		}
		// 将tran_id定为自己
		rolec = &roleCache{
			area:       area,
			role:       &mid,
			role_store: mid,
			be_delete:  TRAN_ROLE_BE_DELETE_NO,
			tran_time:  time.Now(),
			wait_line:  make([]*tranAskGetRole, 0),
			lock:       new(sync.RWMutex),
		}
		// 如果lockmode为写模式，则将tran_id设置为这个tran
		if lockmode == TRAN_LOCK_MODE_WRITE {
			rolec.tran_id = tran_id
		}
		t.role_cache[cache_id] = rolec
		t.lock.Unlock()
		return rolec, nil
	} else {
		// 如果找到了就麻烦了
		// 看删除
		if rolec.be_delete != TRAN_ROLE_BE_DELETE_NO {
			err = fmt.Errorf("The Role already be delete.")
			return nil, err
		}
		//看tran_id是否被指定
		if rolec.tran_id == "" {
			// 没有指定就简单了
			if lockmode == TRAN_LOCK_MODE_WRITE {
				rolec.tran_id = tran_id
			}
			t.lock.Unlock()
			return
		} else {
			// 指定了该怎么办呢，如果是读那就直接返回读就是了
			if lockmode == TRAN_LOCK_MODE_READ {
				t.lock.Unlock()
				return
			} else {
				// 如果不是读，那就要获得写的权限，那就释放锁，排队等待
				// 构造等待队列
				wait := &tranAskGetRole{
					tran_id:  tran_id,
					ask_time: time.Now(),
					approved: make(chan bool),
				}
				// 加入等待队列
				rolec.addWait(wait)
				// 主动解锁
				t.lock.Unlock()
				// 等待回音
				fmt.Println("Tran log ", tran_id, "等待", id)
				ifhave := <-wait.approved
				if ifhave == true {
					fmt.Println("Tran log ", tran_id, "等到了", id)
					// 如果等到了回音，在收到回音的时候，已经得到了被独占的设定，所以直接返回就可以了
					return rolec, nil
				} else {
					err = fmt.Errorf("The Role already be delete.")
					return nil, err
				}

			}
		}
	}
}

// 加入（写入）一个角色
func (t *tranService) addRole(tran_id, area string, mid roles.RoleMiddleData) (rolec *roleCache, err error) {
	t.lock.Lock()
	// 先看能找到吗
	id := mid.Version.Id
	cache_id := area + id

	var find bool
	rolec, find = t.role_cache[cache_id]

	if find == false {
		// 找不到
		if err != nil {
			t.lock.RUnlock()
			return nil, err
		}
		rolec = &roleCache{
			area:       area,
			role:       &mid,
			role_store: mid,
			be_delete:  TRAN_ROLE_BE_DELETE_NO,
			tran_time:  time.Now(),
			wait_line:  make([]*tranAskGetRole, 0),
			lock:       new(sync.RWMutex),
		}
		rolec.tran_id = tran_id
		t.role_cache[cache_id] = rolec
		t.lock.Unlock()
		return rolec, nil
	} else {
		// 找到
		// 构造等待队列
		wait := &tranAskGetRole{
			tran_id:  tran_id,
			ask_time: time.Now(),
			approved: make(chan bool),
		}
		// 加入等待队列
		rolec.addWait(wait)
		// 主动解锁
		t.lock.Unlock()
		// 等待回音
		fmt.Println("Tran log ", tran_id, "等待", id)
		<-wait.approved
		fmt.Println("Tran log ", tran_id, "等到了", id)
		// 如果等到了回音，在收到回音的时候，已经得到了被独占的设定，所以就把角色的主体改了吧
		rolec.role = &mid
		// 如果被确认删除了还就很麻烦的
		if rolec.be_delete != TRAN_ROLE_BE_DELETE_NO {
			rolec.role_store = mid
			rolec.be_delete = TRAN_ROLE_BE_DELETE_NO
		}
		return rolec, nil
	}
}

// 加入某角色缓存的等待队列
func (r *roleCache) addWait(wait *tranAskGetRole) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.wait_line = append(r.wait_line, wait)
}
