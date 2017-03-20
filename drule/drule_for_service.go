// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// ExecTCP的读取角色，需要判断是否在事务中
//	<-- 发送DATA_PLEASE
//	--> 接收角色id
//	<-- Net_RoleSendAndReceive
func (d *DRule) readRole(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 角色id
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	// 查看位置
	connmode, conn := d.findConn(role_id)
	if connmode == CONN_IS_LOCAL {
		//本地处理
		var role roles.Roleer
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 在事务中
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			role, err = tran.ReadRole(role_id)
			if err != nil {
				return err
			}
		} else {
			// 不在事务中
			role, err = d.trule.ReadRole(role_id)
			if err != nil {
				return err
			}
		}
		err = d.sendRoleForExecTCP(role, conn_exec)
		if err != nil {
			return err
		}
	} else {
		// 在slave上
		role_send_b, err := d.getRoleFromSlaves(prefix_stat, role_id, conn)
		if err != nil {
			return err
		}
		// 发送数据
		err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, role_send_b, nil)
		return err
	}
	return
}

// 向请求方发送角色的本体
func (d *DRule) sendRoleForExecTCP(role roles.Roleer, conn_exec *nst.ConnExec) (err error) {
	// 构造发送结构
	roleb, relab, verb, err := hardstore.EncodeRole(role)
	if err != nil {
		return err
	}
	role_send := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	role_send_b, err := nst.StructGobBytes(role_send)
	if err != nil {
		return err
	}
	// 发送数据
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, role_send_b, nil)
	return err
}

// 向slave发送请求角色
//	--> OPERATE_READ_ROLE（前导）
//	<-- 角色id
func (d *DRule) getRoleFromSlaves(prefix_stat Net_PrefixStat, roleid string, conns []*slaveIn) (role_send_b []byte, err error) {
	// 找一个随机的连接
	conncount := len(conns)
	connrandom := random.GetRandNum(conncount - 1)
	cprocess := conns[connrandom].tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_receipt, err := SendPrefixStat(cprocess, conns[connrandom].code, prefix_stat.TransactionId, prefix_stat.InTransaction, roleid, OPERATE_READ_ROLE)
	if err != nil {
		return
	}
	// 角色ID
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, []byte(roleid))
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	role_send_b = slave_receipt.Data
	return
}

// 为ExecTCP的保存角色
//	--> DATA_PLEASE
//	<-- Net_RoleSendAndReceive
//	--> 发送DATA_ALL_OK
func (d *DRule) storeRole(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 看看有roleid吗
	roleid := prefix_stat.RoleId
	if len(roleid) == 0 {
		err = fmt.Errorf("The Role id no be set.")
		return err
	}
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}

	// 接收Net_RoleSendAndReceive
	role_body_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 查看角色应该在哪里保存
	connmode, conn := d.findConn(roleid)
	if connmode == CONN_IS_LOCAL {
		// 本地处理
		// 解码角色
		role_body := Net_RoleSendAndReceive{}
		err = nst.BytesGobStruct(role_body_b, &role_body)
		if err != nil {
			return err
		}
		role, err := hardstore.DecodeRole(role_body.RoleBody, role_body.RoleRela, role_body.RoleVer)
		if err != nil {
			return err
		}
		// 查看是不是在事务中
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 在事务中
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			err = tran.StoreRole(role)
			if err != nil {
				return err
			}
		} else {
			// 不在事务中
			err = d.trule.StoreRole(role)
			if err != nil {
				return err
			}
		}
	} else {
		// 在slave上的话
		err = d.storeRoleForSlaves(prefix_stat, role_body_b, conn)
		if err != nil {
			return err
		}
	}
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, nil, nil)
	return
}

// 为所有slave的存储角色
func (d *DRule) storeRoleForSlaves(prefix_stat Net_PrefixStat, role_body_b []byte, conns []*slaveIn) (err error) {
	errarray := make([]string, 0)
	for _, conn := range conns {
		errone := d.storeRoleForOneSlave(prefix_stat, role_body_b, conn)
		if errone != nil {
			errarray = append(errarray, fmt.Sprint(errone))
		}
	}
	if len(errarray) != 0 {
		errstr := strings.Join(errarray, " | ")
		return fmt.Errorf(errstr)
	}
	return
}

// 为某一个slave存储角色
//	--> OPERATE_WRITE_ROLE(前导)
//	<-- DATA_PLEASE
//	--> Net_RoleSendAndReceive(byte)
//	<-- DATA_ALL_OK
func (d *DRule) storeRoleForOneSlave(prefix_stat Net_PrefixStat, role_body_b []byte, conn *slaveIn) (err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, prefix_stat.TransactionId, prefix_stat.InTransaction, prefix_stat.RoleId, OPERATE_WRITE_ROLE)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送Net_RoleSendAndReceive
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, role_body_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 为ExecTCP的删除角色
//	<-- 交由ExecTCP发送DATA_ALL_OK
func (d *DRule) deleteRole(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 查看有roleid吗
	roleid := prefix_stat.RoleId
	if len(roleid) == 0 {
		err = fmt.Errorf("The Role id not be set.")
		return err
	}
	// 查看角色的位置
	connmode, conn := d.findConn(roleid)
	if connmode == CONN_IS_LOCAL {
		// 本地
		// 查看事务情况
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// In transaction
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			err = tran.DeleteRole(roleid)
			if err != nil {
				return err
			}
		} else {
			// Not in transaction
			err = d.trule.DeleteRole(roleid)
			if err != nil {
				return err
			}
		}
	} else {
		// In slaves
		err = d.deleteRoleFromSlaves(prefix_stat, conn)
		if err != nil {
			return err
		}
	}
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, nil, nil)
	return nil
}

// 为所有slave的删除角色
func (d *DRule) deleteRoleFromSlaves(prefix_stat Net_PrefixStat, conns []*slaveIn) (err error) {
	errarray := make([]string, 0)
	for _, conn := range conns {
		errone := d.deleteRoleFromOneSlave(prefix_stat, conn)
		if errone != nil {
			errarray = append(errarray, fmt.Sprint(errone))
		}
	}
	if len(errarray) != 0 {
		errstr := strings.Join(errarray, " | ")
		return fmt.Errorf(errstr)
	}
	return nil
}

// 为某一个slave删除角色
//	--> OPERATE_DEL_ROLE
//	<-- DATA_ALL_OK
func (d *DRule) deleteRoleFromOneSlave(prefix_stat Net_PrefixStat, conn *slaveIn) (err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, prefix_stat.TransactionId, prefix_stat.InTransaction, prefix_stat.RoleId, OPERATE_DEL_ROLE)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// 为WriteSometing的设置父角色（事务版）
func (d *DRule) writeFatherTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	net_role_father_change := Net_RoleFatherChange{}
	err = nst.BytesGobStruct(byte_slice_data, &net_role_father_change)
	if err != nil {
		return
	}
	// 执行
	err = tran.WriteFather(net_role_father_change.Id, net_role_father_change.Father)
	if err != nil {
		return
	}
	return
}

// 为WriteSometing的设置父角色（非事务版）
func (d *DRule) writeFatherNoTran(byte_slice_data []byte) (err error) {
	// 解码
	net_role_father_change := Net_RoleFatherChange{}
	err = nst.BytesGobStruct(byte_slice_data, &net_role_father_change)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.WriteFather(net_role_father_change.Id, net_role_father_change.Father)
	if err != nil {
		return
	}
	return
}

// 为readSometing的读取父角色（事务版）
func (d *DRule) readFatherTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	net_role_father_change := Net_RoleFatherChange{}
	err = nst.BytesGobStruct(byte_slice_data, &net_role_father_change)
	if err != nil {
		return
	}
	// 执行
	net_role_father_change.Father, err = tran.ReadFather(net_role_father_change.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(net_role_father_change)
	return
}

// 为readSometing的读取父角色（非事务版）
func (d *DRule) readFatherNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	net_role_father_change := Net_RoleFatherChange{}
	err = nst.BytesGobStruct(byte_slice_data, &net_role_father_change)
	if err != nil {
		return
	}
	// 执行
	net_role_father_change.Father, err = d.trule.ReadFather(net_role_father_change.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(net_role_father_change)
	return
}

// 为readSometing的读取所有子角色（事务版）
func (d *DRule) readChildrenTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	roleid := string(byte_slice_data)
	// 执行
	children, err := tran.ReadChildren(roleid)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(children)
	return
}

// 为readSometing的读取所有子角色（事务版）
func (d *DRule) readChildrenNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	roleid := string(byte_slice_data)
	// 执行
	children, err := d.trule.ReadChildren(roleid)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(children)
	return
}

// 为writeSometing的设置所有子角色（事务版）
func (d *DRule) writeChildrenTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_children := Net_RoleAndChildren{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_children)
	if err != nil {
		return
	}
	// 执行
	err = tran.WriteChildren(role_and_children.Id, role_and_children.Children)
	return
}

// 为writeSometing的设置所有子角色（非事务版）
func (d *DRule) writeChildrenNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_children := Net_RoleAndChildren{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_children)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.WriteChildren(role_and_children.Id, role_and_children.Children)
	return
}

// 为writeSometing的设置一个子角色（事务版）
func (d *DRule) writeChildTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	err = tran.WriteChild(role_and_child.Id, role_and_child.Child)
	return
}

// 为writeSometing的设置一个子角色（非事务版）
func (d *DRule) writeChildNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.WriteChild(role_and_child.Id, role_and_child.Child)
	return
}

// 为writeSometing的删除一个子角色（事务版）
func (d *DRule) deleteChildTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	err = tran.DeleteChild(role_and_child.Id, role_and_child.Child)
	return
}

// 为writeSometing的删除一个子角色（非事务版）
func (d *DRule) deleteChildNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.DeleteChild(role_and_child.Id, role_and_child.Child)
	return
}

// 为readSometing的是否有子角色（事务版）
func (d *DRule) existChildTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	role_and_child.Exist, err = tran.ExistChild(role_and_child.Id, role_and_child.Child)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_child)
	return
}

// 为readSometing的是否有子角色（非事务版）
func (d *DRule) existChildNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_child)
	if err != nil {
		return
	}
	// 执行
	role_and_child.Exist, err = d.trule.ExistChild(role_and_child.Id, role_and_child.Child)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_child)
	return
}

// 为readSometing的读取所有朋友（事务版）
func (d *DRule) readFriendsTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	roleid := string(byte_slice_data)
	// 执行
	friends, err := tran.ReadFriends(roleid)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(friends)
	return
}

// 为readSometing的读取所有朋友（非事务版）
func (d *DRule) readFriendsNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	roleid := string(byte_slice_data)
	// 执行
	friends, err := d.trule.ReadFriends(roleid)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(friends)
	return
}

// 为writeSometing的设置所有朋友（事务版）
func (d *DRule) writeFriendsTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_friends := Net_RoleAndFriends{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_friends)
	if err != nil {
		return
	}
	// 执行
	err = tran.WriteFriends(role_and_friends.Id, role_and_friends.Friends)
	return
}

// 为writeSometing的设置所有朋友（非事务版）
func (d *DRule) writeFriendsNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_friends := Net_RoleAndFriends{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_friends)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.WriteFriends(role_and_friends.Id, role_and_friends.Friends)
	return
}

// 为writeSometing的删除一个朋友（事务版）
func (d *DRule) deleteFriendTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_friend)
	if err != nil {
		return
	}
	// 执行
	err = tran.DeleteFriend(role_and_friend.Id, role_and_friend.Friend)
	return
}

// 为writeSometing的删除一个朋友（非事务版）
func (d *DRule) deleteFriendNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_friend)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.DeleteFriend(role_and_friend.Id, role_and_friend.Friend)
	return
}

// 为writeSometing的创建一个空的上下文（事务版）
func (d *DRule) createContextTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = tran.CreateContext(role_and_context.Id, role_and_context.Context)
	return
}

// 为writeSometing的创建一个空的上下文（非事务版）
func (d *DRule) createContextNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.CreateContext(role_and_context.Id, role_and_context.Context)
	return
}

// 为writeSometing的删除一个上下文（事务版）
func (d *DRule) dropContextTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = tran.DropContext(role_and_context.Id, role_and_context.Context)
	return
}

// 为writeSometing的删除一个上下文（非事务版）
func (d *DRule) dropContextNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.DropContext(role_and_context.Id, role_and_context.Context)
	return
}

// 为readSometing的读取某个上下文（事务版）
func (d *DRule) readContextTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.ContextBody, role_and_context.Exist, err = tran.ReadContext(role_and_context.Id, role_and_context.Context)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为readSometing的读取某个上下文（事务版）
func (d *DRule) readContextNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.ContextBody, role_and_context.Exist, err = d.trule.ReadContext(role_and_context.Id, role_and_context.Context)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为writeSometing的删除一个上下文中的绑定（事务版）
func (d *DRule) deleteContextBindTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = tran.DeleteContextBind(role_and_context.Id, role_and_context.Context, role_and_context.UpOrDown, role_and_context.BindRole)
	return
}

// 为writeSometing的删除一个上下文中的绑定（非事务版）
func (d *DRule) deleteContextBindNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_and_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.DeleteContextBind(role_and_context.Id, role_and_context.Context, role_and_context.UpOrDown, role_and_context.BindRole)
	return
}

// 为readSometing的返回某个上下文的同样绑定值的所有（事务版）
func (d *DRule) readContextSameBindTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.Gather, role_and_context.Exist, err = tran.ReadContextSameBind(role_and_context.Id, role_and_context.Context, role_and_context.UpOrDown, role_and_context.Int)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为readSometing的返回某个上下文的同样绑定值的所有（非事务版）
func (d *DRule) readContextSameBindNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.Gather, role_and_context.Exist, err = d.trule.ReadContextSameBind(role_and_context.Id, role_and_context.Context, role_and_context.UpOrDown, role_and_context.Int)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为readSometing的返回所有上下文组的名称（事务版）
func (d *DRule) readContextsNameTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.Gather, err = tran.ReadContextsName(role_and_context.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为readSometing的返回所有上下文组的名称（非事务版）
func (d *DRule) readContextsNameNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_and_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_and_context)
	if err != nil {
		return
	}
	// 执行
	role_and_context.Gather, err = d.trule.ReadContextsName(role_and_context.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_and_context)
	return
}

// 为writeSometing的设置朋友的状态（事务版）
func (d *DRule) writeFriendStatusTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_friend)
	if err != nil {
		return
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = tran.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = tran.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = tran.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return
}

// 为writeSometing的设置朋友的状态（非事务版）
func (d *DRule) writeFriendStatusNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_friend)
	if err != nil {
		return
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = d.trule.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = d.trule.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = d.trule.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return
}

// 为readSometing的获取朋友的状态（事务版）
func (d *DRule) readFriendStatusTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_friend)
	if err != nil {
		return
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		var value int64
		err = tran.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Int = value
	case roles.STATUS_VALUE_TYPE_FLOAT:
		var value float64
		err = tran.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Float = value
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		var value complex128
		err = tran.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Complex = value
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_friend)
	return
}

// 为readSometing的获取朋友的状态（非事务版）
func (d *DRule) readFriendStatusNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(byte_slice_data, &role_friend)
	if err != nil {
		return
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		var value int64
		err = d.trule.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Int = value
	case roles.STATUS_VALUE_TYPE_FLOAT:
		var value float64
		err = d.trule.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Float = value
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		var value complex128
		err = d.trule.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		role_friend.Complex = value
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_friend)
	return
}

// 为writeSometing的设置某个上下文的属性（事务版）
func (d *DRule) writeContextStatusTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_context)
	if err != nil {
		return
	}
	// 执行
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = tran.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = tran.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = tran.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return
}

// 为writeSometing的设置某个上下文的属性（非事务版）
func (d *DRule) writeContextStatusNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_context)
	if err != nil {
		return
	}
	// 执行
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = d.trule.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = d.trule.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = d.trule.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	return
}

// 为readSometing的获取某个上下文的状态（事务版）
func (d *DRule) readContextStatusTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_context)
	if err != nil {
		return
	}
	// 执行
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		var value int64
		err = tran.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Int = value
	case roles.STATUS_VALUE_TYPE_FLOAT:
		var value float64
		err = tran.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Float = value
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		var value complex128
		err = tran.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Complex = value
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_context)
	return
}

// 为readSometing的获取某个上下文的状态（非事务版）
func (d *DRule) readContextStatusNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_context)
	if err != nil {
		return
	}
	// 执行
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		var value int64
		err = d.trule.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Int = value
	case roles.STATUS_VALUE_TYPE_FLOAT:
		var value float64
		err = d.trule.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Float = value
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		var value complex128
		err = d.trule.ReadContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, &value)
		role_context.Complex = value
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_context)
	return
}
