// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this souDMODE_MASTERrce code is governed by GNU LGPL v3 license

package drcm

import (
	"sync"
	"fmt"
	
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 从永久存储读出一个角色
func (z *ZrStorage) ReadRole (id string) (role roles.Roleer, err error) {
	// 查看缓存，如果缓存里有则从缓存里直接调用。
	rolec, find := z.rolesCache[id];
	if find == true {
		return rolec.role, nil;
	}
	
	connmode, conn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 如果是本地，就调用配套的hardstore的方法
		role, err = z.local_store.ReadRole(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadRole: %v",err);
			return nil, err;
		}
		z.rolesCache[id] = oneRoleCache{
			lock : new(sync.RWMutex),
			role : role,
		};
		return role, nil;
	} else {
		// 如果没有，因为是读取，所以就随即从一个slave中调用
		conncount := len(conn);
		connrandom := random.GetRandNum(conncount - 1);
		role, err = z.readRole(conn[connrandom], id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadRole: %v",err);
			return nil, err;
		}
		z.rolesCache[id] = oneRoleCache{
			lock : new(sync.RWMutex),
			role : role,
		};
		return role, nil;
	}
}

// 从slave读取一个角色
func (z *ZrStorage) readRole (conn *slaveIn, id string) (role roles.Roleer, err error) {
	
}

// 查看连接是哪个，id为角色的id，connmode来自CONN_IS_*
func (z *ZrStorage) findConn (id string) (connmode uint8, conn []*slaveIn) {
	// 如果模式为own，则直接返回本地
	if z.dmode == DMODE_OWN {
		connmode = CONN_IS_LOCAL;
		return;
	}
	
	// 找到第一个首字母。
	theChar := string(id[0]);
	// slave池中有没有
	conn, find := z.slaves[theChar];
	if find == false {
		// 如果在slave池里没有找到，那么就默认为本地存储
		connmode = CONN_IS_LOCAL;
		return;
	} else {
		connmode = CONN_IS_SLAVE;
		return;
	}
}

// 向slave发送前导状态，也就是身份验证码和要操作的状态，并获取slave是否可以继续传输的要求
func (z *ZrStorage) sendPrefixStat (conn *slaveIn, operate int) (status uint8, err error) {
	thestat := PrefixStat{
		Operate : operate,
		Code : conn.code,
	};
	statbyte, err := nst.StructGobBytes(thestat);
	if err != nil {
		return;
	}
}

// 查看是否被标记删除，标记删除则返回true。
func (z *ZrStorage) checkDel (id string) bool {
	del := z.checkDelById(id);
	if del == true {
		return del;
	}
	rolec, find := z.rolesCache[id];
	if find == false {
		return true;
	}else{
		del = rolec.role.ReturnDelete();
		return del;
	}
}

func (z *ZrStorage) checkDelById (id string) bool {
	for _, v := range z.deleteCache {
		if v == id {
			return true;
		}
	}
	return false;
}
