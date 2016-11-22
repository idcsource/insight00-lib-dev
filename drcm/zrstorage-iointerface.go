// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

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
		// 如果没有，因为是读取，所以就随机从一个slave中调用
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
		z.rolesCache[id] = &oneRoleCache{
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

// 本地的角色小读取，不带锁，但会加入缓存，并会检查缓存，返回的是oneRoleCache
func (z *ZrStorage) readRole_small (id string) (rolec *oneRoleCache, err error) {
	// 如果开启了缓存，就去找缓存
	if z.cacheMax >= 0 {
		if _, find := z.rolesCache[id]; find == true {
			return z.rolesCache[id], nil;
		}
	}
	role, err := z.local_store.ReadRole(id);
	if err != nil {
		err = fmt.Errorf("drcm[ZrStorage]readRole_small: %v",err);
		return nil, err;
	}
	// 如果开启了缓存，则存入缓存，并使其检查缓存
	if z.cacheMax >= 0 {
		z.rolesCache[id] = &oneRoleCache{
			lock : new(sync.RWMutex),
			role : role,
		};
		z.rolesCount++;
	}
	if z.cacheMax > 0 {
		z.checkCacheNum();
	}
	return z.rolesCache[id], nil;
}

// 从slave读取一个角色
//
//	--> OPERATE_READ_ROLE (前导)
//	<-- DATA_ALL_OK (slave回执)
//	--> 角色ID
//	<-- DATA_WILL_SEND (slave回执)
//	--> DATA_PLEASE (uint8)
//	<-- Net_RoleSendAndReceive (结构体)
func (z *ZrStorage) readRole (id string, slave *slaveIn) (role roles.Roleer, err error) {
	cprocess := slave.tcpconn.OpenProgress();
	defer cprocess.Close();
	slavereceipt, err := z.sendPrefixStat(cprocess, slave.code, OPERATE_READ_ROLE);
	if err != nil {
		return nil, err;
	}
	// 如果获取到的DATA_ALL_OK则说明认证已经通过
	if slavereceipt.DataStat != DATA_ALL_OK {
		return nil, slavereceipt.Error;
	}
	// 发送想要的id，并接收slave的返回
	sreb, err := cprocess.SendAndReturn([]byte(id));
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
	// 请求对方发送数据，使用DATA_PLEASE状态，并接收角色的byte流，这是一个Net_RoleSendAndReceive的值。
	dataplace := nst.Uint8ToBytes(DATA_PLEASE);
	rdata, err := cprocess.SendAndReturn(dataplace);
	if err != nil {
		return nil, err;
	}
	// 解码Net_RoleSendAndReceive。
	rolegetstruct := Net_RoleSendAndReceive{};
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
			z.rolesCache[id] = &oneRoleCache{
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
	// 将角色编码，并生成传输所需要的Net_RoleSendAndReceive格式，并最终编码成为[]byte
	roleb, relab, err := z.local_store.EncodeRole(role);
	if err != nil {
		return err;
	}
	roleS := Net_RoleSendAndReceive{
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
		err = z.storeRole_one(roleS_b, onec);
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring);
	}
	return nil;
}

// 存储的一个slave连接
//
//	--> OPERATE_WRITE_ROLE (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleSendAndReceive (结构体)
//	<-- DATA_ALL_OK (salve回执)
func (z *ZrStorage) storeRole_one (roleS_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	//发送前导
	slavereceipt, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_WRITE_ROLE);
	if err != nil {
		return err;
	}
	// 如果slave请求发送数据
	if slavereceipt.DataStat == DATA_PLEASE {
		srb, err := cprocess.SendAndReturn(roleS_b);
		if err != nil {
			return err;
		}
		sr, err := z.decodeSlaveReceipt(srb);
		if err != nil {
			return err;
		}
		if sr.DataStat != DATA_ALL_OK {
			return sr.Error;
		}
		return nil;
	} else {
		return slavereceipt.Error;
	}
}


// 删除一个角色
func (z *ZrStorage) DeleteRole (id string) (err error) {
	// 如果启用了缓存，则启用全局的读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
		// 检查自己的缓存里有没有这个家伙，如果有先删除之，因为是删除，所以就不触发缓存个数检查了
		_, find := z.rolesCache[id];
		if find == true {
			delete(z.rolesCache, id);
			z.rolesCount--;
		}
	}
	// 检查这个角色应该保存在哪里
	connmode, slaveconn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 如果这个角色在本地，那么就调用本地的删除之
		err = z.local_store.DeleteRole(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]DeleteRole: %v",err);
			return;
		}
	} else {
		// 如果是slave
		err = z.deleteRole(id, slaveconn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]DeleteRole: %v",err);
		}
		return err;
	}
	return nil;
}

// 向slave要求删除一个角色，需要将所有镜像同时删除，slave上不存在也是返回正常的
func (z *ZrStorage) deleteRole (id string, conns []*slaveIn) (err error) {
	// 遍历slave的连接，如果slave出现错误就输出，继续下一个结点
	var errstring string;
	for _, onec := range conns {
		err = z.deleteRole_one(id, onec);
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring);
	}
	return nil;
}

// 删除的一个slave链接
//
//	--> OPERATE_DEL_ROLE (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> 角色ID
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) deleteRole_one (id string, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	//发送前导,OPERATE_DEL_ROLE
	slavereceipt, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_DEL_ROLE);
	if err != nil {
		return err;
	}
	// 如果slave请求发送数据
	if slavereceipt.DataStat == DATA_PLEASE {
		// 将id编码后发出去
		srb, err := cprocess.SendAndReturn([]byte(id));
		if err != nil {
			return err;
		}
		// 解码返回值
		sr, err := z.decodeSlaveReceipt(srb);
		if err != nil {
			return err;
		}
		if sr.DataStat != DATA_ALL_OK {
			return sr.Error;
		}
		return nil;
	} else {
		return slavereceipt.Error;
	}
}

// 设置父角色
func (z *ZrStorage) WriteFather (id, father string) (err error) {
	// 如果启用了缓存，则启用全局的读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 是否为本地
	connmode, slaveconn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 如果为本地
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteFather: %v", err);
			return err;
		}
		rolec.lock.Lock();
		defer rolec.lock.Unlock();
		rolec.role.SetFather(father);
		return nil;
	} else {
		// 如果是slave
		err = z.writeFather(id, father, slaveconn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteRole: %v",err);
		}
		return err;
	}
}

// 发送slave设置角色的父角色
//
//	--> OPERATE_SET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleFatherChange (结构)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) writeFather (id, father string, conns []*slaveIn) (err error) {
	// 构造要发送的信息
	sd := Net_RoleFatherChange{Id: id, Father: father};
	sdb, err := nst.StructGobBytes(sd);
	if err != nil {
		return err;
	}
	// 遍历slave的连接，如果slave出现错误就输出，继续下一个结点
	var errstring string;
	for _, onec := range conns {
		err = z.writeFather_one(sdb, onec);
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring);
	}
	return nil;
}

// 发送slave设置角色的父角色——一个slave的
func (z *ZrStorage) writeFather_one (sdb []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	//发送前导，OPERATE_SET_FATHER
	slavereceipt, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_SET_FATHER);
	if err != nil {
		return err;
	}
	if slavereceipt.DataStat == DATA_PLEASE {
		sre, err := cprocess.SendAndReturn(sdb);
		if err != nil {
			return err;
		}
		sr, err := z.decodeSlaveReceipt(sre);
		if err != nil {
			return err;
		}
		if sr.DataStat != DATA_ALL_OK {
			return sr.Error;
		}
		return nil;
	} else {
		return slavereceipt.Error;
	}
}

// 获取父角色的ID
func (z *ZrStorage) ReadFather (id string) (father string, err error) {
	// 如果启用了缓存，则启用全局的读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 是否为本地
	connmode, slaveconn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 如果是本地，则用ReadRole来读取
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadFather: %v", err);
			return "", err;
		}
		// 给这个角色加读锁
		rolec.lock.RLock();
		defer rolec.lock.RUnlock();
		father = rolec.role.GetFather();
		return father, nil;
	} else {
		// 如果不是本地，因为是读取，所以就随机从一个slave中调用
		conncount := len(slaveconn);
		connrandom := random.GetRandNum(conncount - 1);
		father, err = z.readFather(id, slaveconn[connrandom]);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadFather: %v", err);
		}
		return;
	}
}

// 一个slave的返回父亲
//
//	分配连接进程
//	--> OPERATE_GET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色id的byte)
//	<-- DATA_WILL_SEND (slave回执)
//	--> DATA_PLEASE (uint8)
//	<-- father's id (父角色id的byte)
func (z *ZrStorage) readFather (id string, conn *slaveIn) (father string, err error) {
	// 分配连接进程
	cprocess := conn.tcpconn.OpenProgress();
	defer cprocess.Close();
	// 发送前导词，OPERATE_GET_FATHER
	slavereceipt, err := z.sendPrefixStat(cprocess, conn.code, OPERATE_GET_FATHER);
	if err != nil {
		return "", err;
	}
	if slavereceipt.DataStat == DATA_PLEASE {
		// 将自己的id发送出去
		srb, err := cprocess.SendAndReturn([]byte(id));
		if err != nil {
			return "", err;
		}
		sr, err := z.decodeSlaveReceipt(srb);
		if err != nil {
			return "", err;
		}
		if sr.DataStat == DATA_WILL_SEND {
			// 发送DATA_PLEASE并接收返回数据，fatherid
			dataplace := nst.Uint8ToBytes(DATA_PLEASE);
			father_b, err := cprocess.SendAndReturn(dataplace);
			if err != nil {
				return "", err;
			}
			father = string(father_b);
			return father, nil;
		} else {
			return "", sr.Error;
		}
	} else {
		return "", slavereceipt.Error;
	}
}

// 重置父角色，这里只是调用WriteFather
func (z *ZrStorage) ResetFather (id string) error {
	return z.WriteFather(id, "");
}

// 读取角色的所有子角色名
func (z *ZrStorage) ReadChildren (id string) (children []string, err error) {
	// 如果启用了缓存，则启用全局读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 是否为本地
	connmode, conn := z.findConn(id);
	if connmode == CONN_IS_LOCAL {
		// 本地的解决方案
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadChildren: %v", err);
			return children, err;
		}
		// 给这个角色加读锁
		rolec.lock.RLock();
		defer rolec.lock.RUnlock();
		// 从角色获得信息
		children = rolec.role.GetChildren();
		return children, nil;
	} else {
		// 如果需要来slave
		conn_count := len(conn);
		conn_random := random.GetRandNum(conn_count - 1);
		children, err = z.readChildren(id, conn[conn_random]);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]ReadChildren: %v", err);
			return children, err;
		}else{
			return children, nil;
		}
	}
}

// 从slave读出一个角色的children
//
//	分配连接进程
//	--> OPERATE_GET_CHILDREN (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色的id)
//	<-- DATA_WILL_SEND (slave回执)
//	--> DATA_PLEASE (uint8)
//	<-- children's id ([]string)
func (z *ZrStorage) readChildren (id string, conn *slaveIn) (children []string, err error) {
	// 分配连接进程
	cprocess := conn.tcpconn.OpenProgress();
	defer cprocess.Close();
	// 发送前导 OPERATE_GET_CHILDREN
	sr, err := z.sendPrefixStat(cprocess, conn.code, OPERATE_GET_CHILDREN);
	if err != nil {
		return;
	}
	if sr.DataStat == DATA_PLEASE {
		// 发送要查询的id
		sr2_b, err := cprocess.SendAndReturn([]byte(id));
		if err != nil {
			return children, err;
		}
		// 解码slave返回
		sr2, err := z.decodeSlaveReceipt(sr2_b);
		if err != nil {
			return children, err;
		}
		if sr2.DataStat == DATA_WILL_SEND {
			// 请slave发送
			data_please := nst.Uint8ToBytes(DATA_PLEASE);
			children_b, err := cprocess.SendAndReturn(data_please);
			if err != nil {
				return children, err;
			}
			children = make([]string, 0);
			err = nst.BytesGobStruct(children_b, &children);
			return children, err;
		} else {
			err = sr2.Error;
			return children, err;
		}
	} else {
		err = sr.Error;
		return children, err;
	}
}

// 写入角色的所有子角色名
func (z *ZrStorage) WriteChildren (id string, children []string) (err error) {
	// 如果启用了缓存，则启用全局的读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 是否为本地
	mode, conn := z.findConn(id);
	if mode == CONN_IS_LOCAL {
		// 如果在本地
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteChildren: %v", err);
			return err;
		}
		// 锁住这个角色
		rolec.lock.Lock();
		defer rolec.lock.Unlock();
		rolec.role.SetChildren(children);
		return nil;
	} else {
		// 如果在slave
		err = z.writeChildren(id, children, conn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteChildren: %v", err);
		}
		return err;
	}
}

// 在slave上设置children
func (z *ZrStorage) writeChildren (id string, children []string, conns []*slaveIn) (err error) {
	// 构造要传输的信息
	children_b , err := nst.StructGobBytes(children);
	if err != nil {
		return err;
	}
	// 遍历所有的连接
	var errstring string;
	for _, onec := range conns {
		err = z.writeChildren_one (id, children_b, onec);
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring);
	} else {
		return nil;
	}
}

// 向某一个slavev发送设置children的内容
//
//	分配连接
//	--> OPREATE_SET_CHILDREN (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id
//	<-- DATA_PLEASE (slave回执)
//	--> children's []byte
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) writeChildren_one (id string, children_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	// OPREATE_SET_CHILDREN 前导
	slave_receipt, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_SET_CHILDREN);
	if err != nil {
		return err;
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return slave_receipt.Error;
	}
	// 发送角色id
	slave_receipt_b, err := cprocess.SendAndReturn([]byte(id));
	if err != nil {
		return err;
	}
	slave_receipt, err = z.decodeSlaveReceipt(slave_receipt_b);
	if err != nil {
		return err;
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return slave_receipt.Error;
	}
	// 发送children
	slave_receipt_b, err = cprocess.SendAndReturn(children_b);
	if err != nil {
		return err;
	}
	slave_receipt, err = z.decodeSlaveReceipt(slave_receipt_b);
	if err != nil {
		return err;
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return slave_receipt.Error;
	}
	return nil;
}

// 重置角色的子角色关系，只是调用WriteCildren
func (z *ZrStorage) ResetChildren (id string) (err error) {
	children := make([]string,0);
	return z.WriteChildren(id, children);
}

// 写入一个子角色关系
func (z *ZrStorage) WriteChild (id, child string) (err error) {
	// 如果启用了缓存，启用全局读锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 检查是否为本地
	mode, conn := z.findConn(id);
	if mode == CONN_IS_LOCAL {
		// 如果是本地，则读出或查看缓存什么的
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteChild: %v", err);
			return err;
		}
		// 给这个role加锁
		rolec.lock.Lock();
		defer rolec.lock.Unlock();
		// 设置关系
		rolec.role.AddChild(child);
		return nil;
	} else {
		// 如果是slave的
		err = z.writeChild(id, child, conn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]WriteChild: %v", err);
		}
		return err;
	}
}

// 向slave发送添加child关系的命令
func (z *ZrStorage) writeChild (id , child string, conns []*slaveIn) (err error) {
	// 构建要发送的信息
	role_child := Net_RoleAndChild{
		Id : id,
		Child : child,
	};
	role_child_b, err := nst.StructGobBytes(role_child);
	if err != nil {
		return err;
	}
	//遍历slave连接
	var errstr string;
	for _, onec := range conns {
		err = z.writeChild_one(role_child_b, onec);
		if err != nil {
			errstr += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstr) != 0 {
		return fmt.Errorf(errstr);
	}else{
		return nil;
	}
}

// 向其中一个slave发送添加child的命令
//
//	--> OPERATE_ADD_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) writeChild_one (role_child_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	// 发送前导
	slave_reply, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_ADD_CHILD);
	if err != nil {
		return err;
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return slave_reply.Error;
	}
	slave_reply, err = z.sendAndDecodeSlaveReceipt(cprocess, role_child_b);
	if err != nil {
		return err;
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return slave_reply.Error;
	}
	return nil;
}

// 从永久存储里删除一个子角色关系
func (z *ZrStorage) DeleteChild (id, child string) (err error) {
	// 如果启动了缓存，则启动全局锁
	if z.cacheMax >= 0 {
		z.lock.RLock();
		defer z.lock.RUnlock();
	}
	// 查询是否本地
	mode, conn := z.findConn(id);
	if mode == CONN_IS_LOCAL {
		// 如果是本地
		rolec, err := z.readRole_small(id);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]DeleteChild: %v", err);
			return err;
		}
		rolec.lock.Lock();
		defer rolec.lock.Unlock();
		rolec.role.DeleteChild(child);
		return nil;
	} else {
		// 如果是slave
		err = z.deleteChild(id, child, conn);
		if err != nil {
			err = fmt.Errorf("drcm[ZrStorage]DeleteChild: %v", err);
		}
		return err;
	}
}

// 让slave删除一个child关系
func (z *ZrStorage) deleteChild (id, child string, conns []*slaveIn) (err error) {
	// 构造要发送的信息
	role_child := Net_RoleAndChild{
		Id:			id,
		Child:		child,
	};
	role_child_b, err := nst.StructGobBytes(role_child);
	if err != nil {
		return err;
	}
	// 遍历slave连接
	var errstr string;
	for _, onec := range conns {
		err = z.deleteChild_one(role_child_b, onec);
		if err != nil {
			errstr += fmt.Sprint(onec.name, ": ", err, " | ");
		}
	}
	if len(errstr) != 0 {
		return fmt.Errorf(errstr);
	}
	return nil;
}

// 对某一个slave发送删除child关系的请求
//
//	--> OPERATE_DEL_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) deleteChild_one (role_child_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress();
	defer cprocess.Close();
	// 发送前导
	slave_reply, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_DEL_CHILD);
	if err != nil {
		return err;
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return slave_reply.Error;
	}
	// 发送数据
	slave_reply, err = z.sendAndDecodeSlaveReceipt(cprocess, role_child_b);
	if err != nil {
		return err;
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return slave_reply.Error;
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
func (z *ZrStorage) decodeSlaveReceipt (b []byte) (receipt Net_SlaveReceipt, err error) {
	receipt = Net_SlaveReceipt{};
	err = nst.BytesGobStruct(b, &receipt);
	return;
}

// 发送数据并解码返回的SlaveReceipt
func (z *ZrStorage) sendAndDecodeSlaveReceipt (cprocess *nst.ProgressData, data []byte) (receipt Net_SlaveReceipt, err error) {
	s_r_b, err := cprocess.SendAndReturn(data);
	if err != nil {
		return;
	}
	receipt, err = z.decodeSlaveReceipt(s_r_b);
	return;
}

// 向slave发送前导状态，也就是身份验证码和要操作的状态，并获取slave是否可以继续传输的要求
func (z *ZrStorage) sendPrefixStat (process *nst.ProgressData, code string, operate int) (receipt Net_SlaveReceipt, err error) {
	thestat := Net_PrefixStat{
		Operate : operate,
		Code : code,
	};
	statbyte, err := nst.StructGobBytes(thestat);
	if err != nil {
		return;
	}
	rdata, err := process.SendAndReturn(statbyte);
	if err != nil {
		return;
	}
	receipt = Net_SlaveReceipt{};
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
