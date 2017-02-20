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
		// 进行运行时存储的操作
		err = z.severToToStore(conn_exec)
	case OPERATE_READ_ROLE:
		// 读取角色，对应ReadRole
		err = z.serverToReadRole(conn_exec)
	case OPERATE_WRITE_ROLE:
		// 写角色，对应StoreRole
		err = z.serverToStoreRole(conn_exec)
	case OPERATE_NEW_ROLE:

	case OPERATE_DEL_ROLE:
		// 删除一个角色，对应DeleteRole
		err = z.serverToDeleteRole(conn_exec)
	case OPERATE_GET_DATA:

	case OPERATE_SET_DATA:

	case OPERATE_SET_FATHER:
		// 设置父角色
		err = z.serverToWriteFather(conn_exec)
	case OPERATE_GET_FATHER:
		// 读取父角色
		err = z.serverToReadFather(conn_exec)
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

// 进行运行时存储的操作
// 发送DATA_ALL_OK
// 然后调用本地的运行时保存方法
func (z *ZrStorage) severToToStore(conn_exec *nst.ConnExec) (err error) {
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	if err != nil {
		return err
	}
	err = z.toCacheStore()
	return err
}

// 读取角色
// <-- 发送DATA_PLEASE (slave回执)
// --> 接收role'id
// <-- 发送Net_RoleSendAndReceive (结构体，用Net_SlaveReceipt_Data封装)
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
		z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return nil
	}
	roleb, relab, verb, err := z.local_store.EncodeRole(role)
	if err != nil {
		z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return nil
	}
	// 构建要发送的数据
	role_send := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	role_send_b, err := nst.StructGobBytes(role_send)
	if err != nil {
		z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return nil
	}
	// 发送
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, role_send_b, nil)
	return err
}

// 存储角色
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleSendAndReceive (结构体)
//	<-- DATA_ALL_OK (salve回执)
func (z *ZrStorage) serverToStoreRole(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收数据Net_RoleSendAndReceive
	role_send_and_receive_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_send_and_receive := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(role_send_and_receive_b, &role_send_and_receive)
	if err != nil {
		return err
	}
	role, err := z.local_store.DecodeRole(role_send_and_receive.RoleBody, role_send_and_receive.RoleRela, role_send_and_receive.RoleVer)
	if err != nil {
		return err
	}
	err = z.StoreRole(role)
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
	} else {
		err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	}
	return err
}

// 删除角色
//
//	<-- DATA_PLEASE (slave回执)
//	--> 角色ID
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToDeleteRole(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收信息
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	err = z.DeleteRole(role_id)
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
		return err
	}
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 设置父角色
//	--> OPERATE_SET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleFatherChange (结构)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteFather(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 获取Net_RoleFatherChange结构
	net_role_father_change_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	net_role_father_change := Net_RoleFatherChange{}
	err = nst.BytesGobStruct(net_role_father_change_b, &net_role_father_change_b)
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
		return err
	}
	err = z.WriteFather(net_role_father_change.Id, net_role_father_change.Father)
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
		return err
	}
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 读取父角色
//	--> OPERATE_GET_FATHER (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色id的byte)
//	<-- father's id (父角色id的byte，Net_SlaveReceipt_Data封装)
func (z *ZrStorage) serverToReadFather(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收id
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	// 找到父
	father_id, err := z.ReadFather(role_id)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return nil
	}
	// 发送Send回执
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, []byte(father_id), nil)
	return err
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

// 向请求方返回带有数据体的回执信息
func (z *ZrStorage) serverDataReceipt(conn_exec *nst.ConnExec, stat uint8, data []byte, err error) (err2 error) {
	// 构建回执
	slave_receipt := Net_SlaveReceipt_Data{
		DataStat: stat,
		Error:    err,
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
