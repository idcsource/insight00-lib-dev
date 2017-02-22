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
	"github.com/idcsource/Insight-0-0-lib/roles"
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
		// 写children
		err = z.serverToWriteChildren(conn_exec)
	case OPERATE_GET_CHILDREN:
		// 读取children
		err = z.serverToReadChildren(conn_exec)
	case OPERATE_RESET_CHILDREN:

	case OPERATE_ADD_CHILD:
		// 加一个child
		err = z.serverToWriteChild(conn_exec)
	case OPERATE_DEL_CHILD:
		// 删除一个child
		err = z.serverToDeleteChild(conn_exec)
	case OPERATE_EXIST_CHILD:
		// 是否存在child
		err = z.serverToExistChild(conn_exec)
	case OPERATE_SET_FRIENDS:
		// 写入全部朋友
		err = z.serverToWriteFriends(conn_exec)
	case OPERATE_GET_FRIENDS:
		// 读全部朋友
		err = z.serverToReadFriends(conn_exec)
	case OPERATE_RESET_FRIENDS:

	case OPERATE_ADD_FRIEND:

	case OPERATE_DEL_FRIEND:
		// 删除一个朋友
		err = z.serverToDeleteFriend(conn_exec)
	case OPERATE_CHANGE_FRIEND:

	case OPERATE_SAME_BIND_FRIEND:

	case OPERATE_ADD_CONTEXT:
		//创建一个空的上下文
		err = z.serverToCreateContext(conn_exec)
	case OPERATE_DROP_CONTEXT:
		// 删除一个上下文
		err = z.serverToDropContext(conn_exec)
	case OPERATE_GET_CONTEXTS_NAME:
		// 返回所有上下文组的名称
		err = z.serverToReadContextsName(conn_exec)
	case OPERATE_READ_CONTEXT:
		// 读出一个上下文
		err = z.serverToReadContext(conn_exec)
	case OPERATE_SAME_BIND_CONTEXT:
		//返回某个上下文中的同样绑定值的所有
		err = z.serverToReadContextSameBind(conn_exec)
	case OPERATE_ADD_CONTEXT_BIND:

	case OPERATE_DEL_CONTEXT_BIND:
		// 删除一个上下文的绑定
		err = z.serverToDeleteContextBind(conn_exec)
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
		// 设置朋友的状态属性
		err = z.serverToWriteFriendStatus(conn_exec)
	case OPERATE_GET_FRIEND_STATUS:
		// 读取朋友的某个状态
		err = z.serverToReadFriendStatus(conn_exec)
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
		//conn_exec.SendClose()
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

// 读取Children
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色的id)
//	<-- children's id ([]string，Net_SlaveReceipt_Data封装)
func (z *ZrStorage) serverToReadChildren(conn_exec *nst.ConnExec) (err error) {
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
	children, err := z.ReadChildren(role_id)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return err
	}
	// 构造要发送的内容
	children_b, err := nst.StructGobBytes(children)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return err
	}
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, children_b, nil)
	return err
}

// 写入Children
//	--> OPERATE_SET_CHILDREN (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChildren
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteChildren(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_children_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_children := Net_RoleAndChildren{}
	err = nst.BytesGobStruct(role_children_b, &role_children)
	if err != nil {
		return err
	}
	// 调用
	err = z.WriteChildren(role_children.Id, role_children.Children)
	if err != nil {
		return err
	}
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 写入一个child
//	--> OPERATE_ADD_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteChild(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_child_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(role_child_b, &role_child)
	if err != nil {
		return err
	}
	// 执行
	err = z.WriteChild(role_child.Id, role_child.Child)
	if err != nil {
		return err
	}
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 删除一个child
//	--> OPERATE_DEL_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToDeleteChild(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_child_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(role_child_b, &role_child)
	if err != nil {
		return err
	}
	// 执行
	err = z.DeleteChild(role_child.Id, role_child.Child)
	if err != nil {
		return err
	}
	// 成功回执
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 是否存在child
//	--> OPERATE_EXIST_CHILD (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndChild (结构体)
//	<-- DATA_RETURN_IS_TRUE 或 DATA_RETURN_IS_FALSE (slave回执)
func (z *ZrStorage) serverToExistChild(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_child_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_child := Net_RoleAndChild{}
	err = nst.BytesGobStruct(role_child_b, &role_child)
	if err != nil {
		return err
	}
	// 执行
	have, err := z.ExistChild(role_child.Id, role_child.Child)
	if err != nil {
		return err
	}
	// 回执
	if have == true {
		err = z.serverErrorReceipt(conn_exec, DATA_RETURN_IS_TRUE, nil)
	} else {
		err = z.serverErrorReceipt(conn_exec, DATA_RETURN_IS_FALSE, nil)
	}
	return err
}

// 读全部朋友
//	--> OPERATE_GET_FRIENDS (前导词)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id (角色ID)
//	<-- friends's status (map[string]roles.Status，Net_SlaveReceipt_Data封装)
func (z *ZrStorage) serverToReadFriends(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 读id
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	// 执行
	friends, err := z.ReadFriends(role_id)
	if err != nil {
		return err
	}
	// 封装
	friends_b, err := nst.StructGobBytes(friends)
	if err != nil {
		return err
	}
	// 发送
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, friends_b, nil)
	return err
}

// 写全部朋友
//	--> OPERATE_SET_FRIENDS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriends
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteFriends(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 获取信息
	role_friends_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_friends := Net_RoleAndFriends{}
	err = nst.BytesGobStruct(role_friends_b, &role_friends)
	if err != nil {
		return err
	}
	// 执行
	err = z.WriteFriends(role_friends.Id, role_friends.Friends)
	if err != nil {
		return err
	}
	// 发送成功
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 删除一个朋友
//	--> OPERATE_DEL_FRIEND (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToDeleteFriend(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_friend_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(role_friend_b, &role_friend)
	if err != nil {
		return err
	}
	// 执行
	err = z.DeleteFriend(role_friend.Id, role_friend.Friend)
	if err != nil {
		return err
	}
	// 回执成功
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 创建一个空的上下文，如果已经存在则忽略
//	--> OPERATE_ADD_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToCreateContext(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(role_context_b, &role_context)
	if err != nil {
		return err
	}
	// 执行
	err = z.CreateContext(role_context.Id, role_context.Context)
	if err != nil {
		return err
	}
	// 发送成功
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// drop上下文的请求
//
//	--> OPERATE_DROP_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToDropContext(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(role_context_b, &role_context)
	if err != nil {
		return err
	}
	// 执行
	err = z.DropContext(role_context.Id, role_context.Context)
	if err != nil {
		return err
	}
	// 发送成功
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// readContext
//
//	--> OPERATE_READ_CONTEXT (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext (结构体)
//	<-- context (roles.Context，Net_SlaveReceipt_Data封装)
func (z *ZrStorage) serverToReadContext(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(role_context_b, &role_context)
	if err != nil {
		return err
	}
	// 执行
	context, have, err := z.ReadContext(role_context.Id, role_context.Context)
	if err != nil {
		return err
	}
	// 如果没找到的发送
	if have == false {
		err = z.serverDataReceipt(conn_exec, DATA_RETURN_IS_FALSE, nil, nil)
		if err != nil {
			return err
		}
	}
	// 构建发送数据
	context_b, err := nst.StructGobBytes(context)
	if err != nil {
		return err
	}
	// 发送数据
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, context_b, nil)
	return err
}

// 清除一个上下文的绑定
//	--> OPERATE_DEL_CONTEXT_BIND (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext ([]byte)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToDeleteContextBind(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_context := Net_RoleAndContext{}
	err = nst.BytesGobStruct(role_context_b, &role_context)
	if err != nil {
		return err
	}
	// 执行
	err = z.DeleteContextBind(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole)
	if err != nil {
		return err
	}
	// 构建成功
	err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	return err
}

// 返回某个上下文中的同样绑定值的所有
//	--> OPERATE_SAME_BIND_CONTEXT (前导)
//	<-- DATA_PLEASE (slave 回执)
//	--> Net_RoleAndContext_Data (结构体)
//	<-- rolesid []string ([]byte数据，Net_SlaveReceipt_Data封装)
func (z *ZrStorage) serverToReadContextSameBind(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(role_context_b, &role_context)
	if err != nil {
		return err
	}
	// 执行
	rolesid, have, err := z.ReadContextSameBind(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.Int)
	if err != nil {
		return err
	}
	if have == false {
		err = z.serverDataReceipt(conn_exec, DATA_RETURN_IS_FALSE, nil, nil)
		return err
	}
	// 构建发送
	rolesid_b, err := nst.StructGobBytes(rolesid)
	if err != nil {
		return err
	}
	// 发送结果
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, rolesid_b, nil)
	return err
}

// 返回所有上下文组的名称
//	--> OPERATE_GET_CONTEXTS_NAME (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> role's id
//	<-- names (slave回执带数据体)
func (z *ZrStorage) serverToReadContextsName(conn_exec *nst.ConnExec) (err error) {
	// 回执
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
	// 执行
	contexts_name, err := z.ReadContextsName(role_id)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return err
	}
	// 构造要发送的数据
	contexts_name_b, err := nst.StructGobBytes(contexts_name)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return err
	}
	// 发送结果
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, contexts_name_b, nil)
	return err
}

// 设置朋友的状态属性
// 	--> OPERATE_SET_FRIEND_STATUS (前导词)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteFriendStatus(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_friend_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(role_friend_b, &role_friend)
	if err != nil {
		return err
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = z.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = z.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = z.WriteFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, role_friend.Complex)
	case roles.STATUS_VALUE_TYPE_NULL:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	// 回执
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
	} else {
		err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	}
	return err

}

// 获取朋友的状态属性
//	--> OPERATE_GET_FRIEND_STATUS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndFriend (结构体)
//	<-- Net_RoleAndFriend带上value (slave回执带数据体)
func (z *ZrStorage) serverToReadFriendStatus(conn_exec *nst.ConnExec) (err error) {
	// 回执
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收
	role_friend_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码
	role_friend := Net_RoleAndFriend{}
	err = nst.BytesGobStruct(role_friend_b, &role_friend)
	if err != nil {
		return err
	}
	// 执行
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		var value int64
		err = z.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		if err != nil {
			err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
			return err
		}
		role_friend.Int = value
	case roles.STATUS_VALUE_TYPE_FLOAT:
		var value float64
		err = z.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		if err != nil {
			err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
			return err
		}
		role_friend.Float = value
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		var value complex128
		err = z.ReadFriendStatus(role_friend.Id, role_friend.Friend, role_friend.Bit, &value)
		if err != nil {
			err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
			return err
		}
		role_friend.Complex = value
	case roles.STATUS_VALUE_TYPE_NULL:
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, fmt.Errorf("The value's type not int64, float64 or complex128."))
		return err
	}
	// 编码
	value_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		err = z.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return err
	}
	// 发送
	err = z.serverDataReceipt(conn_exec, DATA_ALL_OK, value_b, nil)
	return err
}

// 设定上下文的状态属性
//	--> OPERATE_SET_CONTEXT_STATUS (前导)
//	<-- DATA_PLEASE (slave回执)
//	--> Net_RoleAndContext_Data (结构体)
//	<-- DATA_ALL_OK (slave回执)
func (z *ZrStorage) serverToWriteContextStatus(conn_exec *nst.ConnExec) (err error) {
	// 回执please
	err = z.serverErrorReceipt(conn_exec, DATA_PLEASE, nil)
	if err != nil {
		return err
	}
	// 接收数据
	role_context_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 解码数据
	role_context := Net_RoleAndContext_Data{}
	err = nst.BytesGobStruct(role_context_b, role_context)
	if err != nil {
		return err
	}
	// 执行
	//WriteContextStatus(id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error)
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		err = z.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		err = z.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		err = z.WriteContextStatus(role_context.Id, role_context.Context, role_context.UpOrDown, role_context.BindRole, role_context.Bit, role_context.Complex)
	default:
		err = fmt.Errorf("The value's type not int64, float64 or complex128.")
	}
	// 回执
	if err != nil {
		err = z.serverErrorReceipt(conn_exec, DATA_NOT_EXPECT, err)
	} else {
		err = z.serverErrorReceipt(conn_exec, DATA_ALL_OK, nil)
	}
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
