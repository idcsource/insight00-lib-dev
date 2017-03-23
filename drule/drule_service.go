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

	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// ExecTCP nst的ConnExecer接口
func (d *DRule) ExecTCP(conn_exec *nst.ConnExec) (err error) {
	// 接收前导
	prefix_stat_b, err := conn_exec.GetData()
	if err != nil {
		d.logerr(fmt.Errorf("Get the Prefix Stat err : %v", err))
		return err
	}
	// 解码前导
	prefix_stat := Net_PrefixStat{}
	err = nst.BytesGobStruct(prefix_stat_b, &prefix_stat)
	if err != nil {
		d.logerr(fmt.Errorf("Get the Prefix Stat err : %v", err))
		return err
	}
	// 检查身份验证码
	if prefix_stat.Code != d.code {
		// 发送错误
		d.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, fmt.Errorf("The service code is wrong : %v .", prefix_stat.Code))
		conn_exec.SendClose()
		return fmt.Errorf("The service code is wrong : %v .", prefix_stat.Code)
	}
	// 开始遍历操作
	var operate_string string
	switch prefix_stat.Operate {
	case OPERATE_TRAN_BEGIN:
		// 开启事务，如果出错则自己负责回滚掉开启的slave的事务
		err = d.beginTransaction(conn_exec)
		operate_string = "OPERATE_TRAN_BEGIN"
	case OPERATE_TRAN_PREPARE:
		// 准备事务，唯一不同就是做好角色的写锁准备
		err = d.prepareTransaction(conn_exec)
		operate_string = "OPERATE_TRAN_PREPARE"
	case OPERATE_TRAN_ROLLBACK:
		// 回滚事务
		if prefix_stat.InTransaction == false || len(prefix_stat.TransactionId) == 0 {
			err = fmt.Errorf("This is not in a transaction.")
		} else {
			err = d.rollbackTransaction(prefix_stat.TransactionId, conn_exec)
		}
		operate_string = "OPERATE_TRAN_ROLLBACK"
	case OPERATE_TRAN_COMMIT:
		// 执行事务
		if prefix_stat.InTransaction == false || len(prefix_stat.TransactionId) == 0 {
			err = fmt.Errorf("This is not in a transaction.")
		} else {
			err = d.commitTransaction(prefix_stat.TransactionId, conn_exec)
		}
	case OPERATE_READ_ROLE:
		// 读取角色
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_READ_ROLE"
	case OPERATE_WRITE_ROLE:
		// 保存角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_WRITE_ROLE"
	case OPERATE_DEL_ROLE:
		// 删除角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_DEL_ROLE"
	case OPERATE_SET_FATHER:
		// 设置父角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_FATHER"
	case OPERATE_GET_FATHER:
		// 获取父角色
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_FATHER"
	case OPERATE_GET_CHILDREN:
		// 获取所有子角色
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_CHILDREN"
	case OPERATE_SET_CHILDREN:
		// 设置所有子角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_CHILDREN"
	case OPERATE_ADD_CHILD:
		// 设置一个子角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_ADD_CHILD"
	case OPERATE_DEL_CHILD:
		// 删除一个子角色
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_DEL_CHILD"
	case OPERATE_EXIST_CHILD:
		// 含有某个子角色
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_EXIST_CHILD"
	case OPERATE_GET_FRIENDS:
		// 读取所有的朋友
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_FRIENDS"
	case OPERATE_SET_FRIENDS:
		// 设置所有朋友
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_FRIENDS"
	case OPERATE_DEL_FRIEND:
		// 删除一个朋友
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_DEL_FRIEND"
	case OPERATE_ADD_CONTEXT:
		// 创建空上下文
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_ADD_CONTEXT"
	case OPERATE_DROP_CONTEXT:
		// 删除一个上下文
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_DROP_CONTEXT"
	case OPERATE_READ_CONTEXT:
		// 读出某个上下文的全部
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_READ_CONTEXT"
	case OPERATE_DEL_CONTEXT_BIND:
		// 删除一个上下文中的绑定
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_DEL_CONTEXT_BIND"
	case OPERATE_SAME_BIND_CONTEXT:
		// 返回某个上下文中同样的绑定值的所有
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SAME_BIND_CONTEXT"
	case OPERATE_GET_CONTEXTS_NAME:
		// 返回所有上下文组的名称
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_CONTEXTS_NAME"
	case OPERATE_SET_FRIEND_STATUS:
		// 设置朋友的状态
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_FRIEND_STATUS"
	case OPERATE_GET_FRIEND_STATUS:
		// 获取朋友的状态
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_FRIEND_STATUS"
	case OPERATE_SET_CONTEXT_STATUS:
		// 设置某个上下文的属性
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_CONTEXT_STATUS"
	case OPERATE_GET_CONTEXT_STATUS:
		// 读取某个上下文的属性
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_CONTEXT_STATUS"
	case OPERATE_SET_DATA:
		// 设定一个值
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_DATA"
	case OPERATE_GET_DATA:
		// 读取一个值
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_DATA"
	case OPERATE_SET_CONTEXTS:
		// 设置全部的上下文
		err = d.writeSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_SET_CONTEXTS"
	case OPERATE_GET_CONTEXTS:
		// 读取全部的上下文
		err = d.readSomeThing(prefix_stat, conn_exec)
		operate_string = "OPERATE_GET_CONTEXTS"
	default:
		err = d.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, fmt.Errorf("The oprerate can not found."))
		d.logerr(fmt.Errorf("drule[DRule]Runtime Error: The client requested a nonexistent operation."))
		conn_exec.SendClose()
		return nil
	}
	if err != nil {
		d.logerr(fmt.Errorf("drule[DRule]Runtime Error: The client %v's operation %v returned an error: ", prefix_stat.ClientName, operate_string, err))
		d.serverDataReceipt(conn_exec, DATA_RETURN_ERROR, nil, err)
		return nil
	}
	return
}

// 向请求方返回带有数据体的回执信息
func (d *DRule) serverDataReceipt(conn_exec *nst.ConnExec, stat uint8, data []byte, err error) (err2 error) {
	// 构建回执
	slave_receipt := Net_SlaveReceipt_Data{
		DataStat: stat,
		Error:    fmt.Sprint(err),
		Data:     data,
	}
	slave_receipt_b, err2 := nst.StructGobBytes(slave_receipt)
	if err2 != nil {
		return err2
	}
	err2 = conn_exec.SendData(slave_receipt_b)
	if err2 != nil {
		return err2
	}
	return nil
}

// ExecTCP的要写点什么
//	判断又无Role ID指定
//	--> DATA_PLEASE
//	<-- 获取[]byte编码的结构体数据
//	分析Net_PrefixStat：角色位置，事务状态
//	DATA_ALL_OK
func (d *DRule) writeSomeThing(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 查看有无role id

	roleid := prefix_stat.RoleId
	if len(roleid) == 0 {
		err = fmt.Errorf("The Role id not be set.")
		return err
	}
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 获取[]byte编码的结构体数据
	byte_slice_data, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 分析Net_PrefixStat，查看角色位置
	connmode, conn := d.findConn(roleid)
	if connmode == CONN_IS_LOCAL {
		// 本地的
		// 查看事务状态
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 事务中，找到事务
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			// 遍历请求
			switch prefix_stat.Operate {
			case OPERATE_WRITE_ROLE:
				// 存一个角色
				err = d.storeRoleTran(tran, byte_slice_data)
			case OPERATE_DEL_ROLE:
				// 删除角色
				err = d.deleteRoleTran(tran, byte_slice_data)
			case OPERATE_SET_FATHER:
				// 设置父角色
				err = d.writeFatherTran(tran, byte_slice_data)
			case OPERATE_SET_CHILDREN:
				// 设置所有子角色
				err = d.writeChildrenTran(tran, byte_slice_data)
			case OPERATE_ADD_CHILD:
				// 设置一个子角色
				err = d.writeChildTran(tran, byte_slice_data)
			case OPERATE_DEL_CHILD:
				// 删掉一个子角色
				err = d.deleteChildTran(tran, byte_slice_data)
			case OPERATE_SET_FRIENDS:
				// 设置所有朋友
				err = d.writeFriendsTran(tran, byte_slice_data)
			case OPERATE_DEL_FRIEND:
				// 删除一个朋友
				err = d.deleteFriendTran(tran, byte_slice_data)
			case OPERATE_ADD_CONTEXT:
				// 创建空上下文
				err = d.createContextTran(tran, byte_slice_data)
			case OPERATE_DROP_CONTEXT:
				// 删除一个上下文
				err = d.dropContextTran(tran, byte_slice_data)
			case OPERATE_DEL_CONTEXT_BIND:
				// 删除一个上下文中的绑定
				err = d.deleteContextBindTran(tran, byte_slice_data)
			case OPERATE_SET_FRIEND_STATUS:
				// 设置朋友的状态
				err = d.writeFriendStatusTran(tran, byte_slice_data)
			case OPERATE_SET_CONTEXT_STATUS:
				// 设置某个上下文的属性
				err = d.writeContextStatusTran(tran, byte_slice_data)
			case OPERATE_SET_DATA:
				// 设定一个值
				err = d.writeDataTran(tran, byte_slice_data)
			case OPERATE_SET_CONTEXTS:
				// 设置全部的上下文
				err = d.writeContextsTran(tran, byte_slice_data)
			default:
				err = fmt.Errorf("The oprerate can not found.")
			}
		} else {
			// 不在事务中，遍历请求
			switch prefix_stat.Operate {
			case OPERATE_WRITE_ROLE:
				// 存一个角色
				err = d.storeRoleNoTran(byte_slice_data)
			case OPERATE_DEL_ROLE:
				// 删除角色
				err = d.deleteRoleNoTran(byte_slice_data)
			case OPERATE_SET_FATHER:
				// 设置父角色
				err = d.writeFatherNoTran(byte_slice_data)
			case OPERATE_SET_CHILDREN:
				// 设置所有子角色
				err = d.writeChildrenNoTran(byte_slice_data)
			case OPERATE_ADD_CHILD:
				// 设置一个子角色
				err = d.writeChildNoTran(byte_slice_data)
			case OPERATE_DEL_CHILD:
				// 删掉一个子角色
				err = d.deleteChildNoTran(byte_slice_data)
			case OPERATE_SET_FRIENDS:
				// 设置所有朋友
				err = d.writeFriendsNoTran(byte_slice_data)
			case OPERATE_DEL_FRIEND:
				// 删除一个朋友
				err = d.deleteFriendNoTran(byte_slice_data)
			case OPERATE_ADD_CONTEXT:
				// 创建空上下文
				err = d.createContextNoTran(byte_slice_data)
			case OPERATE_DROP_CONTEXT:
				// 删除一个上下文
				err = d.dropContextNoTran(byte_slice_data)
			case OPERATE_DEL_CONTEXT_BIND:
				// 删除一个上下文中的绑定
				err = d.deleteContextBindNoTran(byte_slice_data)
			case OPERATE_SET_FRIEND_STATUS:
				// 设置朋友的状态
				err = d.writeFriendStatusNoTran(byte_slice_data)
			case OPERATE_SET_CONTEXT_STATUS:
				// 设置某个上下文的属性
				err = d.writeContextStatusNoTran(byte_slice_data)
			case OPERATE_SET_DATA:
				// 设定一个值
				err = d.writeDataNoTran(byte_slice_data)
			case OPERATE_SET_CONTEXTS:
				// 设置全部的上下文
				err = d.writeContextsNoTran(byte_slice_data)
			default:
				err = fmt.Errorf("The oprerate can not found.")
			}
		}
	} else {
		// 在slave的
		if len(d.selfname) != 0 {
			prefix_stat.ClientName = prefix_stat.ClientName + "|" + d.selfname
		}
		err = d.writeSomeThingFromSlaves(prefix_stat, byte_slice_data, conn)
	}
	if err == nil {
		d.serverDataReceipt(conn_exec, DATA_ALL_OK, nil, nil)
		return nil
	}
	return err
}

// 所有slave的写点什么
func (d *DRule) writeSomeThingFromSlaves(prefix_stat Net_PrefixStat, byte_slice_data []byte, conns []*slaveIn) (err error) {
	errarray := make([]string, 0)
	for _, conn := range conns {
		errone := d.writeSomeThingFromOneSlave(prefix_stat, byte_slice_data, conn)
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

// 某一个slave的写点什么
//	--> 前导
//	<-- DATA_PLEASE
//	--> 发送[]byte编码的结构体数据
//	<-- DATA_ALL_OK
func (d *DRule) writeSomeThingFromOneSlave(prefix_stat Net_PrefixStat, byte_slice_data []byte, conn *slaveIn) (err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, prefix_stat.ClientName, prefix_stat.TransactionId, prefix_stat.InTransaction, prefix_stat.RoleId, prefix_stat.Operate)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}

// ExecTCP的要读点什么
//	判断又无Role ID指定
//	--> DATA_PLEASE
//	<-- 获取[]byte编码的结构体数据
//	分析Net_PrefixStat：角色位置，事务状态
//	--> 带返回数据的DATA_ALL_OK
func (d *DRule) readSomeThing(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 查看又无role id
	roleid := prefix_stat.RoleId
	if len(roleid) == 0 {
		err = fmt.Errorf("The Role id not be set.")
		return err
	}
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 获取[]byte编码的结构体数据
	byte_slice_data, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 构建返回数据
	var return_data []byte
	// 分析Net_PrefixStat，查看角色位置
	connmode, conn := d.findConn(roleid)
	if connmode == CONN_IS_LOCAL {
		// 本地的
		// 查看事务状态
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 事务中，找到事务
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			// 遍历请求
			switch prefix_stat.Operate {
			case OPERATE_READ_ROLE:
				// 读取角色
				return_data, err = d.readRoleTran(tran, byte_slice_data)
			case OPERATE_GET_FATHER:
				// 获取父角色
				return_data, err = d.readFatherTran(tran, byte_slice_data)
			case OPERATE_GET_CHILDREN:
				// 获取所有子角色
				return_data, err = d.readChildrenTran(tran, byte_slice_data)
			case OPERATE_EXIST_CHILD:
				// 是否有一个子角色
				return_data, err = d.existChildTran(tran, byte_slice_data)
			case OPERATE_GET_FRIENDS:
				// 读取所有的朋友
				return_data, err = d.readFriendsTran(tran, byte_slice_data)
			case OPERATE_READ_CONTEXT:
				// 读出某个上下文的全部
				return_data, err = d.readContextTran(tran, byte_slice_data)
			case OPERATE_SAME_BIND_CONTEXT:
				// 返回某个上下文中同样的绑定值的所有
				return_data, err = d.readContextSameBindTran(tran, byte_slice_data)
			case OPERATE_GET_CONTEXTS_NAME:
				// 返回所有上下文组的名称
				return_data, err = d.readContextsNameTran(tran, byte_slice_data)
			case OPERATE_GET_FRIEND_STATUS:
				// 获取朋友的状态
				return_data, err = d.readFriendStatusTran(tran, byte_slice_data)
			case OPERATE_GET_CONTEXT_STATUS:
				// 读取某个上下文的属性
				return_data, err = d.readContextStatusTran(tran, byte_slice_data)
			case OPERATE_GET_DATA:
				// 读取一个值
				return_data, err = d.readDataTran(tran, byte_slice_data)
			case OPERATE_GET_CONTEXTS:
				// 读取全部的上下文
				return_data, err = d.readContextsTran(tran, byte_slice_data)
			default:
				err = fmt.Errorf("The oprerate can not found.")
			}
		} else {
			// 不在事务中，遍历请求
			switch prefix_stat.Operate {
			case OPERATE_READ_ROLE:
				// 读取角色
				return_data, err = d.readRoleNoTran(byte_slice_data)
			case OPERATE_GET_FATHER:
				// 获取父角色
				return_data, err = d.readFatherNoTran(byte_slice_data)
			case OPERATE_GET_CHILDREN:
				// 获取所有子角色
				return_data, err = d.readChildrenNoTran(byte_slice_data)
			case OPERATE_EXIST_CHILD:
				// 是否有一个子角色
				return_data, err = d.existChildNoTran(byte_slice_data)
			case OPERATE_GET_FRIENDS:
				// 读取所有的朋友
				return_data, err = d.readFriendsNoTran(byte_slice_data)
			case OPERATE_READ_CONTEXT:
				// 读出某个上下文的全部
				return_data, err = d.readContextNoTran(byte_slice_data)
			case OPERATE_SAME_BIND_CONTEXT:
				// 返回某个上下文中同样的绑定值的所有
				return_data, err = d.readContextSameBindNoTran(byte_slice_data)
			case OPERATE_GET_CONTEXTS_NAME:
				// 返回所有上下文组的名称
				return_data, err = d.readContextsNameNoTran(byte_slice_data)
			case OPERATE_GET_FRIEND_STATUS:
				// 获取朋友的状态
				return_data, err = d.readFriendStatusNoTran(byte_slice_data)
			case OPERATE_GET_CONTEXT_STATUS:
				// 读取某个上下文的属性
				return_data, err = d.readContextStatusNoTran(byte_slice_data)
			case OPERATE_GET_DATA:
				// 读取一个值
				return_data, err = d.readDataNoTran(byte_slice_data)
			case OPERATE_GET_CONTEXTS:
				// 读取全部的上下文
				return_data, err = d.readContextsNoTran(byte_slice_data)
			default:
				err = fmt.Errorf("The oprerate can not found.")
			}
		}
	} else {
		// 在slave的
		conncount := len(conn)
		connrandom := random.GetRandNum(conncount - 1)
		if len(d.selfname) != 0 {
			prefix_stat.ClientName = prefix_stat.ClientName + "|" + d.selfname
		}
		return_data, err = d.readSomeThingFromSlave(prefix_stat, byte_slice_data, conn[connrandom])
	}
	if err == nil {
		err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, return_data, nil)
		return nil
	}
	return err
}

// 某一个slave的读点什么
//	--> 前导
//	<-- DATA_PLEASE
//	--> 发送[]byte编码的结构体数据
//	<-- 带数据的DATA_ALL_OK
func (d *DRule) readSomeThingFromSlave(prefix_stat Net_PrefixStat, byte_slice_data []byte, conn *slaveIn) (return_data []byte, err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, prefix_stat.ClientName, prefix_stat.TransactionId, prefix_stat.InTransaction, prefix_stat.RoleId, prefix_stat.Operate)
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	return_data = slave_receipt.Data
	return
}
