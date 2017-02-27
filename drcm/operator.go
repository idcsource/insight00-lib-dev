// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 新建一个操作机，addr和code是默认的ZrStorage的地址（含端口号）和身份码，conn_num为连接池的个数
func NewOperator(addr, code string, conn_num int, logs *ilogs.Logs) (operator *Operator, err error) {
	operator = &Operator{
		slaves: make([]*slaveIn, 0),
		logs:   logs,
		lock:   new(sync.RWMutex),
	}
	slave, err := nst.NewTcpClient(addr, conn_num, logs)
	if err != nil {
		return nil, err
	}
	oneSlaveIn := &slaveIn{
		name:    addr,
		code:    code,
		tcpconn: slave,
	}
	operator.slaves = append(operator.slaves, oneSlaveIn)
	return operator, nil
}

// 增加一个服务器到控制器
func (o *Operator) AddServer(addr, code string, conn_num int) (err error) {
	slave, err := nst.NewTcpClient(addr, conn_num, o.logs)
	if err != nil {
		return err
	}
	oneSlaveIn := &slaveIn{
		name:    addr,
		code:    code,
		tcpconn: slave,
	}
	o.slaves = append(o.slaves, oneSlaveIn)
	return nil
}

// 新角色
func (o *Operator) NewRole(id string, new roles.Roleer) roles.Roleer {
	new.New(id)
	return new
}

// 运行时保存
func (o *Operator) ToStore() (err error) {
	for _, onec := range o.slaves {
		// 分配连接
		cprocess := onec.tcpconn.OpenProgress()
		defer cprocess.Close()
		// 发送前导
		slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_TOSTORE)
		if err != nil {
			o.logerr(err)
			//return err
		}
		if slave_receipt.DataStat != DATA_ALL_OK {
			o.logerr(slave_receipt.Error)
			//return slave_receipt.Error
		}
	}
	return nil
}

// 读取角色
func (o *Operator) ReadRole(id string) (role roles.Roleer, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	role, err = o.readRole(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadRole: %v", err)
		return nil, err
	}
	return role, nil
}

func (o *Operator) readRole(id string, slave *slaveIn) (role roles.Roleer, err error) {
	cprocess := slave.tcpconn.OpenProgress()
	defer cprocess.Close()
	slavereceipt, err := SendPrefixStat(cprocess, slave.code, OPERATE_READ_ROLE)
	if err != nil {
		return nil, err
	}
	// 如果获取到的DATA_PLEASE则说明认证已经通过
	if slavereceipt.DataStat != DATA_PLEASE {
		return nil, fmt.Errorf(slavereceipt.Error)
	}
	// 发送想要的id，并接收slave的返回
	slave_receipt_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
	if err != nil {
		return nil, err
	}
	if slave_receipt_data.DataStat != DATA_ALL_OK {
		return nil, fmt.Errorf(slave_receipt_data.Error)
	}
	// 解码Net_RoleSendAndReceive。
	rolegetstruct := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(slave_receipt_data.Data, &rolegetstruct)
	if err != nil {
		return nil, err
	}
	// 合成出role来
	role, err = hardstore.DecodeRole(rolegetstruct.RoleBody, rolegetstruct.RoleRela, rolegetstruct.RoleVer)
	return role, err
}

// 存储角色
func (o *Operator) StoreRole(role roles.Roleer) (err error) {
	// 角色编码
	roleb, relab, verb, err := hardstore.EncodeRole(role)
	if err != nil {
		return err
	}
	roleS := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	roleS_b, err := nst.StructGobBytes(roleS)
	if err != nil {
		return err
	}
	// 遍历slave的连接，如果slave出现错误就输出，继续下一个结点
	var errstring string
	for _, onec := range o.slaves {
		err = o.storeRole(roleS_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring)
	}
	return nil
}

func (o *Operator) storeRole(roleS_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	//发送前导
	slavereceipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_WRITE_ROLE)
	fmt.Println("1")
	if err != nil {
		return err
	}
	// 如果slave请求发送数据
	if slavereceipt.DataStat == DATA_PLEASE {
		fmt.Println("2")
		srb, err := cprocess.SendAndReturn(roleS_b)
		if err != nil {
			return err
		}
		fmt.Println("3")
		sr, err := DecodeSlaveReceipt(srb)
		if err != nil {
			return err
		}
		if sr.DataStat != DATA_ALL_OK {
			return fmt.Errorf(sr.Error)
		}
		return nil
	} else {
		return fmt.Errorf(slavereceipt.Error)
	}
}

// 删除一个角色
func (o *Operator) DeleteRole(id string) (err error) {
	var errstring string
	for _, onec := range o.slaves {
		err = o.deleteRole(id, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring)
	}
	return nil
}

// 删除的一个slave链接
//
//	--> OPERATE_DEL_ROLE (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> 角色ID
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) deleteRole(id string, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	//发送前导,OPERATE_DEL_ROLE
	slavereceipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_DEL_ROLE)
	if err != nil {
		return err
	}
	// 如果slave请求发送数据
	if slavereceipt.DataStat == DATA_PLEASE {
		// 将id编码后发出去
		slavereceipt, err = SendAndDecodeSlaveReceipt(cprocess, []byte(id))
		if err != nil {
			return err
		}
		if slavereceipt.DataStat != DATA_ALL_OK {
			return fmt.Errorf(slavereceipt.Error)
		}
		return nil
	} else {
		return fmt.Errorf(slavereceipt.Error)
	}
}

// 设置父角色
func (o *Operator) WriteFather(id, father string) (err error) {
	// 构造要发送的信息
	sd := Net_RoleFatherChange{Id: id, Father: father}
	sdb, err := nst.StructGobBytes(sd)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeFather(sdb, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteFather: %v", errstring)
	}
	return nil
}

// 发送slave设置角色的父角色——一个slave的
//	--> OPERATE_SET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleFatherChange (结构)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeFather(sdb []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	//发送前导，OPERATE_SET_FATHER
	slavereceipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_FATHER)
	if err != nil {
		return err
	}
	if slavereceipt.DataStat == DATA_PLEASE {

		sr, err := SendAndDecodeSlaveReceipt(cprocess, sdb)
		if err != nil {
			return err
		}
		if sr.DataStat != DATA_ALL_OK {
			return fmt.Errorf(sr.Error)
		}
		return nil
	} else {
		return fmt.Errorf(slavereceipt.Error)
	}
}

// 获取父角色的ID
func (o *Operator) ReadFather(id string) (father string, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	father, err = o.readFather(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadFather: %v", err)
		return "", err
	}
	return
}

// 一个slave的返回父亲
//
//	分配连接进程
//	--> OPERATE_GET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色id的byte)
//	<-- father's id (父角色id的byte，Net_SlaveReceipt_Data封装)
func (o *Operator) readFather(id string, conn *slaveIn) (father string, err error) {
	// 分配连接进程
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导词，OPERATE_GET_FATHER
	slavereceipt, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_FATHER)
	if err != nil {
		return "", err
	}
	if slavereceipt.DataStat == DATA_PLEASE {
		// 将自己的id发送出去
		slave_receipt_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
		if err != nil {
			return "", err
		}
		if slave_receipt_data.DataStat != DATA_ALL_OK {
			return "", fmt.Errorf(slave_receipt_data.Error)
		}
		father = string(slave_receipt_data.Data)
		return father, nil
	} else {
		return "", fmt.Errorf(slavereceipt.Error)
	}
}

// 重置父角色，这里只是调用WriteFather
func (o *Operator) ResetFather(id string) error {
	return o.WriteFather(id, "")
}

// 读取角色的所有子角色名
func (o *Operator) ReadChildren(id string) (children []string, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	children, err = o.readChildren(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadChildren: %v", err)
		return nil, err
	}
	return
}

// 从slave读出一个角色的children
//
//	分配连接进程
//	--> OPERATE_GET_CHILDREN (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色的id)
//	<-- children's id ([]string，Net_SlaveReceipt_Data封装)
func (o *Operator) readChildren(id string, conn *slaveIn) (children []string, err error) {
	// 分配连接进程
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导 OPERATE_GET_CHILDREN
	sr, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_CHILDREN)
	if err != nil {
		return
	}
	if sr.DataStat == DATA_PLEASE {
		// 发送要查询的id
		sr_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
		if err != nil {
			return children, err
		}
		if sr_data.DataStat != DATA_ALL_OK {
			return children, fmt.Errorf(sr_data.Error)
		}
		children = make([]string, 0)
		err = nst.BytesGobStruct(sr_data.Data, &children)
		return children, err
	} else {
		err = fmt.Errorf(sr.Error)
		return children, err
	}
}

// 写入角色的所有子角色名
func (o *Operator) WriteChildren(id string, children []string) (err error) {
	// 构造要发送的信息
	role_children := Net_RoleAndChildren{
		Id:       id,
		Children: children,
	}
	children_b, err := nst.StructGobBytes(role_children)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeChildren(id, children_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteChildren: %v", errstring)
	}
	return nil
}

// 向某一个slavev发送设置children的内容
//
//	分配连接
//	--> OPERATE_SET_CHILDREN (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChildren
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeChildren(id string, children_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// OPREATE_SET_CHILDREN 前导
	slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_CHILDREN)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送children
	slave_receipt, err = SendAndDecodeSlaveReceipt(cprocess, children_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 重置角色的子角色关系，只是调用WriteCildren
func (o *Operator) ResetChildren(id string) (err error) {
	children := make([]string, 0)
	return o.WriteChildren(id, children)
}

// 写入一个子角色关系
func (o *Operator) WriteChild(id, child string) (err error) {
	// 构造要发送的信息
	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeChild(role_child_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteChild: %v", errstring)
	}
	return nil
}

// 向其中一个slave发送添加child的命令
//
//	--> OPERATE_ADD_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeChild(role_child_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_ADD_CHILD)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_child_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 从永久存储里删除一个子角色关系
func (o *Operator) DeleteChild(id, child string) (err error) {
	// 构造要发送的信息
	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.deleteChild(role_child_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]DeleteChild: %v", errstring)
	}
	return nil
}

// 对某一个slave发送删除child关系的请求
//
//	--> OPERATE_DEL_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) deleteChild(role_child_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_DEL_CHILD)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	// 发送数据
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_child_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 查询是否有这个子角色关系，如果有则返回true
func (o *Operator) ExistChild(id, child string) (have bool, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	have, err = o.existChild(id, child, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ExistChild: %v", err)
	}
	return have, nil
}

// 从slave中查看是否有那么一个child角色
//
//	--> OPERATE_EXIST_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_RETURN_IS_TRUE 或 DATA_RETURN_IS_FALSE (slave回执)
func (o *Operator) existChild(id, child string, conn *slaveIn) (have bool, err error) {
	// 分配进程
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导OPERATE_EXIST_CHILD
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_EXIST_CHILD)
	if err != nil {
		return false, err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return false, fmt.Errorf(slave_reply.Error)
	}
	// 创建要发送的结构体
	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		return false, err
	}
	// 向slave发送查询的结构体
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_child_b)
	if err != nil {
		return false, err
	}
	if slave_reply.DataStat == DATA_RETURN_IS_TRUE {
		return true, nil
	} else if slave_reply.DataStat == DATA_RETURN_IS_FALSE {
		return false, nil
	} else {
		return false, fmt.Errorf(slave_reply.Error)
	}
}

// 读取id的所有朋友关系
func (o *Operator) ReadFriends(id string) (status map[string]roles.Status, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	status, err = o.readFriends(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadFriends: %v", err)
	}
	return status, nil
}

// 从slave中读取一个角色的friends关系
//
//	--> OPERATE_GET_FRIENDS (前导词)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色ID)
//	<-- friends's status (map[string]roles.Status，Net_SlaveReceipt_Data封装)
func (o *Operator) readFriends(id string, conn *slaveIn) (status map[string]roles.Status, err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导OPERATE_GET_FRIENDS
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_FRIENDS)
	if err != nil {
		return nil, err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return nil, fmt.Errorf(slave_reply.Error)
	}
	// 发送角色的ID
	slave_reply_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
	if err != nil {
		return nil, err
	}
	if slave_reply_data.DataStat != DATA_ALL_OK {
		return nil, fmt.Errorf(slave_reply_data.Error)
	}
	// 解码status
	status = make(map[string]roles.Status)
	err = nst.BytesGobStruct(slave_reply_data.Data, &status)
	return status, err
}

// 写入角色的所有朋友关系
func (o *Operator) WriteFriends(id string, friends map[string]roles.Status) (err error) {
	// 构造要发送的信息
	role_friends := Net_RoleAndFriends{
		Id:      id,
		Friends: friends,
	}
	friends_b, err := nst.StructGobBytes(role_friends)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeFriends(id, friends_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteFriends: %v", errstring)
	}
	return nil
}

// WriteFriends的向每一个slave发送
//
//	--> OPERATE_SET_FRIENDS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriends
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeFriends(id string, friends_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_FRIENDS)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送Net_RoleAndFriends的byte
	slave_receipt, err = SendAndDecodeSlaveReceipt(cprocess, friends_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 重置角色的所有朋友关系，也就是发送一个空的朋友关系给WriteFriends
func (o *Operator) ResetFriends(id string) (err error) {
	friends := make(map[string]roles.Status)
	return o.WriteFriends(id, friends)
}

// 加入一个朋友关系，并绑定，已经有的关系将之修改绑定值。
// 这是WriteFriendStatus绑定状态的特例，也就是绑定位为0,绑定值为int64类型。
func (o *Operator) WriteFriend(id, friend string, bind int64) (err error) {
	err = o.WriteFriendStatus(id, friend, 0, bind)
	return
}

// 删除一个朋友关系，如果没有则忽略
func (o *Operator) DeleteFriend(id, friend string) (err error) {
	// 构造要发送的信息
	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.deleteFriend(role_friend_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]DeleteFriend: %v", errstring)
	}
	return nil
}

// 一个slave的删除朋友关系
//
//	--> OPERATE_DEL_FRIEND (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) deleteFriend(role_friend_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_DEL_FRIEND)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	// 发送数据
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_friend_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 创建一个空的上下文，如果已经存在则忽略
func (o *Operator) CreateContext(id, contextname string) (err error) {
	// 构建要发送的信息
	role_context := Net_RoleAndContext{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.createContext(role_context_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]CreateContext: %v", errstring)
	}
	return nil
}

// 向某一个slave发送创建上下文的请求
//
//	--> OPERATE_ADD_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) createContext(role_context_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_ADD_CONTEXT)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_context_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	} else {
		return nil
	}
}

// 清除一个上下文，也就是删除
func (o *Operator) DropContext(id, contextname string) (err error) {
	// 构造要发送的信息
	role_context := Net_RoleAndContext{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.dropContext(role_context_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]DropContext: %v", errstring)
	}
	return nil
}

// 向某一个slave发送drop上下文的请求
//
//	--> OPERATE_DROP_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) dropContext(role_context_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_DROP_CONTEXT)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_context_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	} else {
		return nil
	}
}

// 返回某个上下文的全部信息，如果没有这个上下文则have返回false
func (o *Operator) ReadContext(id, contextname string) (context roles.Context, have bool, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	context, have, err = o.readContext(id, contextname, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadRole: %v", err)
	}
	return
}

// slave上的readContext
//
//	--> OPERATE_READ_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- context (roles.Context，Net_SlaveReceipt_Data封装)
func (o *Operator) readContext(id, contextname string, conn *slaveIn) (context roles.Context, have bool, err error) {
	have = false
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_READ_CONTEXT)
	if err != nil {
		return
	}
	if slave_reply.DataStat != DATA_PLEASE {
		err = fmt.Errorf(slave_reply.Error)
		return
	}
	// 构造要发送的结构体
	role_context := Net_RoleAndContext{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return
	}
	slave_reply_data, err := SendAndDecodeSlaveReceiptData(cprocess, role_context_b)
	if err != nil {
		return
	}
	// 看看slave是没有找到还是其他错误
	if slave_reply_data.DataStat != DATA_ALL_OK {
		if slave_reply_data.DataStat == DATA_RETURN_IS_FALSE {
			return context, false, nil
		} else {
			return context, false, fmt.Errorf(slave_reply_data.Error)
		}
	}
	err = nst.BytesGobStruct(slave_reply_data.Data, &context)
	if err != nil {
		return
	}
	return context, true, nil
}

// 清除一个上下文的绑定，upordown为roles包中的CONTEXT_UP或CONTEXT_DOWN，binderole是绑定的角色id
func (o *Operator) DeleteContextBind(id, contextname string, upordown uint8, bindrole string) (err error) {
	// 构造传输的信息
	role_context := Net_RoleAndContext{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindrole,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.deleteContextBind(role_context_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]DeleteContextBind: %v", errstring)
	}
	return nil
}

// 一个的slave清除一个上下文的绑定
//
//	--> OPERATE_DEL_CONTEXT_BIND (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext ([]byte)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) deleteContextBind(role_context_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_DEL_CONTEXT_BIND)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	// 发送数据
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_context_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 返回某个上下文中的同样绑定值的所有，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN，如果给定的contextname不存在，则have返回false。
func (o *Operator) ReadContextSameBind(id, contextname string, upordown uint8, bind int64) (rolesid []string, have bool, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	rolesid, have, err = o.readContextSameBind(id, contextname, upordown, bind, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadContextSameBind: %v", err)
	}
	return
}

// 从一个slave读取某个上下文中的同样绑定值的所有
//
//	--> OPERATE_SAME_BIND_CONTEXT (前导)
//	<-- DATA_PLEASE (slave 回执)
//	--> Net_RoleAndContext_Data (结构体)
//	<-- rolesid []string ([]byte数据，Net_SlaveReceipt_Data封装)
func (o *Operator) readContextSameBind(id, contextname string, upordown uint8, bind int64, conn *slaveIn) (rolesid []string, have bool, err error) {
	// 构造发出的信息
	contextsamebind := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		Single:   roles.STATUS_VALUE_TYPE_INT,
		Bit:      0,
		Int:      bind,
	}
	contextsamebind_b, err := nst.StructGobBytes(contextsamebind)
	if err != nil {
		return nil, false, err
	}
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, OPERATE_SAME_BIND_CONTEXT)
	if err != nil {
		return nil, false, err
	}
	// 查看回执
	if slave_receipt.DataStat != DATA_PLEASE {
		return nil, false, fmt.Errorf(slave_receipt.Error)
	}
	// 发送结构
	slave_receipt_data, err := SendAndDecodeSlaveReceiptData(cprocess, contextsamebind_b)
	// 查看回执
	if slave_receipt_data.DataStat == DATA_RETURN_IS_FALSE {
		// 这是如果没有找到的解决方法
		return nil, false, fmt.Errorf(slave_receipt_data.Error)
	}
	if slave_receipt_data.DataStat != DATA_ALL_OK {
		// 这是不期望的发送
		return nil, false, fmt.Errorf(slave_receipt_data.Error)
	}
	rolesid = make([]string, 0)
	err = nst.BytesGobStruct(slave_receipt_data.Data, &rolesid)
	if err != nil {
		return nil, false, err
	}
	return rolesid, true, nil
}

// 返回所有上下文组的名称
func (o *Operator) ReadContextsName(id string) (names []string, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	names, err = o.readContextsName(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadContextsName: %v", err)
	}
	return
}

// slave上的返回所有上下文组的名称
//
//	--> OPERATE_GET_CONTEXTS_NAME (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id
//	<-- names (slave回执带数据体)
func (o *Operator) readContextsName(id string, conn *slaveIn) (names []string, err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_CONTEXTS_NAME)
	if err != nil {
		return
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return nil, fmt.Errorf(slave_reply.Error)
	}
	// 发送id，并接收带数据体的slave回执
	slave_reply_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
	if err != nil {
		return
	}
	if slave_reply_data.DataStat != DATA_ALL_OK {
		return nil, fmt.Errorf(slave_reply_data.Error)
	}
	names = make([]string, 0)
	err = nst.BytesGobStruct(slave_reply_data.Data, &names)
	return
}

// 设置朋友的状态属性
func (o *Operator) WriteFriendStatus(id, friend string, bindbit int, value interface{}) (err error) {
	// 构建要发送的信息
	statustype := o.statusValueType(value)
	if statustype == 0 {
		return fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
		Single: statustype,
		Bit:    bindbit,
	}
	switch statustype {
	case roles.STATUS_VALUE_TYPE_INT:
		role_friend.Int = value.(int64)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		role_friend.Float = value.(float64)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		role_friend.Complex = value.(complex128)
	default:
		role_friend.Single = roles.STATUS_VALUE_TYPE_NULL
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		return nil
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeFriendStatus(role_friend_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteFriendStatus: %v", errstring)
	}
	return nil
}

// 一个slave的设置朋友的状态属性
// 	--> OPERATE_SET_FRIEND_STATUS (前导词)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeFriendStatus(role_friend_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_FRIEND_STATUS)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_friend_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 获取朋友的状态属性
func (o *Operator) ReadFriendStatus(id, friend string, bindbit int, value interface{}) (err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	err = o.readFriendStatus(o.slaves[connrandom], id, friend, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadFriendStatus: %v", err)
	}
	return
}

// slave获取朋友的状态属性
//
//	--> OPERATE_GET_FRIEND_STATUS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- Net_RoleAndFriend带上value (slave回执带数据体)
func (o *Operator) readFriendStatus(conn *slaveIn, id, friend string, bindbit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	// 看看要什么类型的值
	valuetype := o.statusValueType(value)
	if valuetype == roles.STATUS_VALUE_TYPE_NULL {
		return fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	// 构造查询结构
	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
		Single: valuetype,
		Bit:    bindbit,
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		return err
	}
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_FRIEND_STATUS)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	// 发送要查询的结构，并接收带数据体的slave回执
	slave_reply_data, err := SendAndDecodeSlaveReceiptData(cprocess, role_friend_b)
	if err != nil {
		return err
	}
	if slave_reply_data.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply_data.Error)
	}
	err = nst.BytesGobStruct(slave_reply_data.Data, &role_friend)
	if err != nil {
		return err
	}
	value_reflect := reflect.Indirect(reflect.ValueOf(value))
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		value_reflect.SetInt(role_friend.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		value_reflect.SetFloat(role_friend.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		value_reflect.SetComplex(role_friend.Complex)
	default:
		return fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return nil
}

// 设定上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
func (o *Operator) WriteContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	// 构建要发送的信息
	statustype := o.statusValueType(value)
	if statustype == 0 {
		return fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	role_context := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindroleid,
		Single:   statustype,
		Bit:      bindbit,
	}
	switch statustype {
	case roles.STATUS_VALUE_TYPE_INT:
		role_context.Int = value.(int64)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		role_context.Float = value.(float64)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		role_context.Complex = value.(complex128)
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return nil
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeContextStatus(role_context_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteContextStatus: %v", errstring)
	}
	return nil
}

// 一个slave的设定上下文的状态属性
//
//	--> OPERATE_SET_CONTEXT_STATUS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext_Data (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeContextStatus(role_context_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_CONTEXT_STATUS)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	slave_reply, err = SendAndDecodeSlaveReceipt(cprocess, role_context_b)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply.Error)
	}
	return nil
}

// 获取上下文的状态属性，upordown为roles.CONTEXT_UP或roles.CONTEXT_DOWN
func (o *Operator) ReadContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	err = o.readContextStatus(o.slaves[connrandom], id, contextname, upordown, bindroleid, bindbit, value)
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadContextStatus: %v", err)
	}
	return
}

// slave获取上下文的状态属性
//
//	--> OPERATE_GET_CONTEXT_STATUS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext_Data (结构体)
//	<-- Net_RoleAndContext_Data带上value (slave回执带数据体)
func (o *Operator) readContextStatus(conn *slaveIn, id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	// 看看要什么类型的值
	valuetype := o.statusValueType(value)
	if valuetype == roles.STATUS_VALUE_TYPE_NULL {
		return fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	// 构造查询结构
	role_context := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindroleid,
		Single:   valuetype,
		Bit:      bindbit,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		return err
	}
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_reply, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_CONTEXT_STATUS)
	if err != nil {
		return err
	}
	if slave_reply.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_reply.Error)
	}
	// 发送要查询的结构，并接收带数据体的slave回执
	slave_reply_data, err := SendAndDecodeSlaveReceiptData(cprocess, role_context_b)
	if err != nil {
		return err
	}
	if slave_reply_data.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_reply_data.Error)
	}
	// 解码收到的内容
	err = nst.BytesGobStruct(slave_reply_data.Data, &role_context)
	if err != nil {
		return err
	}
	value_reflect := reflect.Indirect(reflect.ValueOf(value))
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		value_reflect.SetInt(role_context.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		value_reflect.SetFloat(role_context.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		value_reflect.SetComplex(role_context.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return err
}

// 设定上下文
func (o *Operator) WriteContexts(id string, contexts map[string]roles.Context) (err error) {
	// 构建要传输的信息
	role_contexts := Net_RoleAndContexts{
		Id:       id,
		Contexts: contexts,
	}
	contexts_b, err := nst.StructGobBytes(role_contexts)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeContexts(id, contexts_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteContexts: %v", errstring)
	}
	return nil
}

// 在slave上设定上下文_一个
//
//	--> OPERATE_SET_CONTEXTS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContexts ([]byte)
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeContexts(id string, contexts_b []byte, onec *slaveIn) (err error) {
	// 分配连接进程
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_CONTEXTS)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送数据
	slave_receipt, err = SendAndDecodeSlaveReceipt(cprocess, contexts_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 获取上下文
func (o *Operator) ReadContexts(id string) (contexts map[string]roles.Context, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	contexts, err = o.readContexts(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadContexts: %v", err)
	}
	return
}

// Slave的获取上下文
//
//	分配连接进程
//	--> OPERATE_GET_CONTEXTS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色ID)
//	<-- contexts (byte，Net_SlaveReceipt_Data封装)
func (o *Operator) readContexts(id string, conn *slaveIn) (contexts map[string]roles.Context, err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导词
	sr, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_CONTEXTS)
	if err != nil {
		return
	}
	if sr.DataStat != DATA_PLEASE {
		return nil, fmt.Errorf(sr.Error)

	}
	// 发送角色ID
	sr_data, err := SendAndDecodeSlaveReceiptData(cprocess, []byte(id))
	if err != nil {
		return nil, err
	}
	if sr_data.DataStat != DATA_ALL_OK {
		return nil, fmt.Errorf(sr_data.Error)
	}
	// 解码
	contexts = make(map[string]roles.Context)
	err = nst.BytesGobStruct(sr_data.Data, &contexts)
	return contexts, err
}

// 重置上下文，实际也就是利用WriteContexts发一个空的过去
func (o *Operator) ResetContexts(id string) (err error) {
	contexts := make(map[string]roles.Context)
	return o.WriteContexts(id, contexts)
}

// 把data的数据装入role的name值下，如果找不到name，则返回错误
func (o *Operator) WriteData(id, name string, data interface{}) (err error) {
	// 构造要传输的信息
	data_b, err := nst.StructGobBytes(data)
	if err != nil {
		return err
	}
	trans := Net_RoleData_Data{
		Id:   id,
		Name: name,
		Data: data_b,
	}
	trans_b, err := nst.StructGobBytes(trans)
	if err != nil {
		return err
	}
	// 遍历
	var errstring string
	for _, onec := range o.slaves {
		err = o.writeData(id, trans_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf("drcm[Operate]WriteData: %v", errstring)
	}
	return nil
}

// slave把data的数据装入role的name值下，一个的
//
//	--> OPERATE_SET_DATA (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleData_Data
//	<-- DATA_ALL_OK (slave回执)
func (o *Operator) writeData(id string, trans_b []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_SET_DATA)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	//发送数据体
	slave_receipt, err = SendAndDecodeSlaveReceipt(cprocess, trans_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 从角色中知道name的数据名并返回其数据
func (o *Operator) ReadData(id, name string, data interface{}) (err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	err = o.readData(id, name, data, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadData: %v", err)
	}
	return
}

// slave的从角色中知道name的数据名并返回其数据
//
//	--> OPERATE_GET_DATA (前导词)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleData_Data
//	<-- Net_RoleData_Data (跟随DATA_ALL_OK)
func (o *Operator) readData(id, name string, data interface{}, conn *slaveIn) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 构建发送的数据
	trans := Net_RoleData_Data{
		Id:   id,
		Name: name,
	}
	trans_b, err := nst.StructGobBytes(trans)
	if err != nil {
		return err
	}
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, OPERATE_GET_DATA)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送数据体并接收
	slave_receipt_data, err := SendAndDecodeSlaveReceiptData(cprocess, trans_b)
	if err != nil {
		return err
	}
	if slave_receipt_data.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt_data.Error)
	}
	// 解码接收
	role_data := Net_RoleData_Data{}
	err = nst.BytesGobStruct(slave_receipt_data.Data, &role_data)
	if err != nil {
		return err
	}
	data_reflect := reflect.Indirect(reflect.ValueOf(data))
	err = nst.BytesGobReflect(role_data.Data, data_reflect)
	return
}

// 判断friend或context的状态的类型，types：1为int，2为float，3为complex
func (o *Operator) statusValueType(value interface{}) (types uint8) {
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		return roles.STATUS_VALUE_TYPE_INT
	case "float64":
		return roles.STATUS_VALUE_TYPE_FLOAT
	case "complex128":
		return roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		return roles.STATUS_VALUE_TYPE_NULL
	}
}

// 处理错误日志
func (o *Operator) logerr(err interface{}) {
	if err == nil {
		return
	}
	if o.logs != nil {
		o.logs.ErrLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (o *Operator) logrun(err interface{}) {
	if err == nil {
		return
	}
	if o.logs != nil {
		o.logs.RunLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 关闭
func (o *Operator) Close() (err error) {
	for _, onec := range o.slaves {
		onec.tcpconn.Close()
	}
	return nil
}
