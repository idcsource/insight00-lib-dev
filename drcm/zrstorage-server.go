// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/nst"
)

// ExecTCP nst的ConnExecer接口
func (z *ZrStorage) ExecTCP(conn_exec *nst.ConnExec) (err error) {
	// 接收前导
	prefix_stat_b, err := conn_exec.GetData()
	if err != nil {
		z.logerr(err)
		return err
	}
	// 解码前导
	prefix_stat := Net_PrefixStat{}
	err = nst.BytesGobStruct(prefix_stat_b, &prefix_stat)
	if err != nil {
		z.logerr(err)
		// 发送关闭
		conn_exec.SendClose()
		return err
	}
	// 查看身份验证码
	if prefix_stat.Code != z.code {
		// 如果身份码不相符，发送关闭
		conn_exec.SendClose()
		return nil
	}
	// 遍历所有操作,转到相应方法
	switch prefix_stat.Operate {

	case OPERATE_TOSTORE:

	case OPERATE_READ_ROLE:
		// 读取角色，对应ReadRole
		err = z.serverToReadRole(conn_exec)
	case OPERATE_WRITE_ROLE:

	case OPERATE_NEW_ROLE:

	case OPERATE_DEL_ROLE:

	case OPERATE_GET_DATA:

	case OPERATE_SET_DATA:

	case OPERATE_SET_FATHER:

	case OPERATE_GET_FATHER:

	case OPERATE_RESET_FATHER:

	case OPERATE_SET_CHILDREN:

	case OPERATE_GET_CHILDREN:

	case OPERATE_RESET_CHILDREN:

	case OPERATE_ADD_CHILD:

	case OPERATE_DEL_CHILD:

	case OPERATE_EXIST_CHILD:

	case OPERATE_SET_FRIENDS:

	case OPERATE_GET_FRIENDS:

	case OPERATE_RESET_FRIENDS:

	case OPERATE_ADD_FRIEND:

	case OPERATE_DEL_FRIEND:

	case OPERATE_CHANGE_FRIEND:

	case OPERATE_SAME_BIND_FRIEND:

	case OPERATE_ADD_CONTEXT:

	case OPERATE_DROP_CONTEXT:

	case OPERATE_GET_CONTEXTS_NAME:

	case OPERATE_READ_CONTEXT:

	case OPERATE_SAME_BIND_CONTEXT:

	case OPERATE_ADD_CONTEXT_BIND:

	case OPERATE_DEL_CONTEXT_BIND:

	case OPERATE_CHANGE_CONTEXT_BIND:

	case OPERATE_CONTEXT_SAME_BIND:

	case OPERATE_ADD_CONTEXT_UP:

	case OPERATE_DEL_CONTEXT_UP:

	case OPERATE_CHANGE_CONTEXT_UP:

	case OPERATE_SAME_BIND_CONTEXT_UP:

	case OPERATE_ADD_CONTEXT_DOWN:

	case OPERATE_DEL_CONTEXT_DOWN:

	case OPERATE_CHANGE_CONTEXT_DOWN:

	case OPERATE_SAME_BIND_CONTEXT_DOWN:

	case OPERATE_SET_FRIEND_STATUS:

	case OPERATE_GET_FRIEND_STATUS:

	case OPERATE_SET_CONTEXT_STATUS:

	case OPERATE_GET_CONTEXT_STATUS:

	case OPERATE_SET_CONTEXTS:

	case OPERATE_GET_CONTEXTS:

	case OPERATE_RESET_CONTEXTS:

	default:
		// 构建关闭
		z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, fmt.Errorf("The oprerate can not found."))
		conn_exec.SendClose()
		return nil
	}
	if err != nil {
		z.serverErrorReceipt(conn_exec, DATA_RETURN_ERROR, err)
		return fmt.Errorf("drcm[ZrStorage]: %v", err)
	}
	return nil
}

// 读取角色
// <-- 发送DATA_PLEASE (slave回执)
// --> 接收role'id
// <-- 发送DATA_WILL_SEND
// --> 接收DATA_PLEASE
func (z *ZrStorage) serverToReadRole(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收role'id
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	// 找有没有这个role
	role, err := z.ReadRole(role_id)
	if err != nil {
		// 找不到就构建找不到的回执并发出去
		z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
		return nil
	}
	roleb, relab, verb, err := z.local_store.EncodeRole(role)
	if err != nil {
		return err
	}
	// 构建要发送的数据
	role_send := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	role_send_b, err := nst.StructGobBytes(role_send)
	if err != nil {
		return err
	}
	// 发送准备接收的回执
	err = z.serverErrorReceipt(conn_exec, DATA_WILL_SEND, nil)
	if err != nil {
		return err
	}
	statb, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	stat := nst.BytesToUint8(statb)
	if stat != DATA_PLEASE {
		return fmt.Errorf("The stat not expect.")
	}
	// 发送数据体
	conn_exec.SendData(role_send_b)
	return nil
}

// 向请求方返回错误回执
func (z *ZrStorage) serverErrorReceipt(conn_exec *nst.ConnExec, stat uint8, err error) (err2 error) {
	// 构建回执
	slave_receipt := Net_SlaveReceipt{
		DataStat: stat,
		Error:    err,
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
