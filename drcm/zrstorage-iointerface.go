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
	// 如果启用了缓存，则启用全局的读锁。
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
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
	} else {
		// 如果没有，因为是读取，所以就随即从一个slave中调用
		conncount := len(conn);
		connrandom := random.GetRandNum(conncount - 1);
		role, err = z.readRole(id, conn[connrandom]);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadRole: %v",err);
			return nil, err;
		}
	}
	// 如果开启了缓存，则存入缓存，并使其检查缓存
	if z.cacheMax >= 0 {
		z.rolesCache[id] = oneRoleCache{
			lock : new(sync.RWMutex),
			role : role,
		};
		z.rolesCount++;
	}
	if z.cacheMax > 0 {
		z.checkCacheNum();
	}
	return role, nil;
}

// 从slave读取一个角色
func (z *ZrStorage) readRole (id string, slave *slaveIn) (role roles.Roleer, err error) {
	slavereceipt, err := z.sendPrefixStat(slave, OPERATE_READ_ROLE);
	if err != nil {
		return nil, err;
	}
	// 如果获取到的DATA_ALL_OK则说明认证已经通过
	if slavereceipt.DataStat != DATA_ALL_OK {
		return nil, slavereceipt.Error;
	}
	// 发送想要的id，并接收slave的返回
	sreb, err := slave.tcpconn.SendAndReturn([]byte(id));
	if err != nil {
		return nil, err;
	}
	// 解码返回值
	slavereceipt, err = z.decodeSlaveReceipt(sreb);
	if err != nil {
		return nil, err;
	}
	// 如果回执状态不是DATA_WILL_SEND，因为我们希望slave是应该把role发送给我们的
	if slavereceipt.DataStat != DATA_WILL_SEND {
		return nil, slavereceipt.Error;
	}
	// 请求对方发送数据，使用DATA_PLEASE状态，并接收角色的byte流，这是一个RoleSendAndReceive的值。
	dataplace := nst.Uint8ToBytes(DATA_PLEASE);
	rdata, err := slave.tcpconn.SendAndReturn(dataplace);
	if err != nil {
		return nil, err;
	}
	// 解码RoleSendAndReceive。
	rolegetstruct := RoleSendAndReceive{};
	err = nst.BytesGobStruct(rdata, &rolegetstruct);
	if err != nil {
		return nil, err;
	}
	// 合成出role来
	role, err = z.local_store.DecodeRole(rolegetstruct.RoleBody, rolegetstruct.RoleRela);
	return role, err;
}

// 往永久存储写一个角色
func (z *ZrStorage) StoreRole (role roles.Roleer) (err error) {
	id := role.ReturnId();
	// 如果启用了缓存，则启用全局的读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
		// 如果缓存里没有则加入缓存
		_, find := z.rolesCache[id];
		if find == false {
			z.rolesCache[id] = oneRoleCache{
				lock : new(sync.RWMutex),
				role : role,
			};
			z.rolesCount++;
		}
		// 如果缓存有个数要求，那么就检查个数要求
		if z.cacheMax > 0 {
			z.checkCacheNum();
		}
	}
	// 检查这个角色应该保存在哪里
	connmode, slaveconn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 如果是本地保存
		err = z.local_store.StoreRole(role);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]StoreRole: %v",err);
		}
		return err;
	} else {
		// 如果是slave保存
		err = z.storeRole(role, slaveconn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]StoreRole: %v",err);
		}
		return err;
	}
}

// 将角色保存到slave中，因为是保存所以需要将所有镜像同时保存
func (z *ZrStorage) storeRole (role roles.Roleer, conns []*slaveIn) (err error) {
	// 将角色编码，并生成传输所需要的RoleSendAndReceive格式，并最终编码成为[]byte
	roleb, relab, err := z.local_store.EncodeRole(role);
	if err != nil {
		return err;
	}
	roleS := RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
	};
	roleS_b, err := nst.StructGobBytes(roleS);
	if err != nil {
		return err;
	}
	// 遍历slave的连接，如果slave出现错误就输出，继续下一个结点
	var errstring string;
	for _, onec := range conns {
		//发送前导
		slavereceipt, err := z.sendPrefixStat(onec, OPERATE_WRITE_ROLE);
		if err != nil {
			errstring += fmt.Sprint(err, " | ");
			continue;
		}
		// 如果slave请求发送数据
		if slavereceipt.DataStat == DATA_PLEASE {
			srb, err := onec.tcpconn.SendAndReturn(roleS_b);
			if err != nil {
				errstring += fmt.Sprint(err, " | ");
				continue;
			}
			sr, err := z.decodeSlaveReceipt(srb);
			if err != nil {
				errstring += fmt.Sprint(err, " | ");
				continue;
			}
			if sr.DataStat != DATA_ALL_OK {
				errstring += fmt.Sprint(sr.Error, " | ");
				continue;
			}
		} else {
			errstring += fmt.Sprint(slavereceipt.Error, " | ");
			continue;
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring);
	}
	return nil;
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

// 从[]byte解码SlaveReceipt
func (z *ZrStorage) decodeSlaveReceipt (b []byte) (receipt SlaveReceipt, err error) {
	receipt = SlaveReceipt{};
	err = nst.BytesGobStruct(b, &receipt);
	return;
}

// 向slave发送前导状态，也就是身份验证码和要操作的状态，并获取slave是否可以继续传输的要求
func (z *ZrStorage) sendPrefixStat (slavein *slaveIn, operate int) (receipt SlaveReceipt, err error) {
	thestat := PrefixStat{
		Operate : operate,
		Code : slavein.code,
	};
	statbyte, err := nst.StructGobBytes(thestat);
	if err != nil {
		return;
	}
	rdata, err := slavein.tcpconn.SendAndReturn(statbyte);
	if err != nil {
		return;
	}
	receipt = SlaveReceipt{};
	err = nst.BytesGobStruct(rdata, &receipt);
	return;
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
