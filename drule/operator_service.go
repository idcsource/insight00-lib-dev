// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"reflect"

	"github.com/idcsource/insight00-lib/nst"
	"github.com/idcsource/insight00-lib/roles"
)

// 是否存在一个角色
func (o *Operator) RoleExist(id string) (have bool, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_sr := Net_RoleSendAndReceive{}
	role_sr.RoleID = id

	role_sr_b, err := nst.StructGobBytes(role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]RoleExist: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_EXIST_ROLE, role_sr_b, &role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]RoleExist: %v", err)
		return
	}
	have = role_sr.IfHave
	return
}

// 读取一个角色
func (o *Operator) ReadRole(id string, role roles.Roleer) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 去执行
	role_sr := Net_RoleSendAndReceive{}
	role_sr.RoleID = id
	// 转码
	role_sr_b, err := nst.StructGobBytes(role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadRole: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_READ_ROLE, role_sr_b, &role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadRole: %v", err)
		return
	}
	// 还原角色
	mid := roles.RoleMiddleData{}
	err = nst.BytesGobStruct(role_sr.RoleBody, &mid)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadRole: %v", err)
		return
	}
	err = roles.DecodeMiddleToRole(mid, role)
	return
}

// 存储一个角色
func (o *Operator) StoreRole(role roles.Roleer) (err error) {
	// 获取角色ID
	roleid := role.ReturnId()

	err = o.checkDMZ(roleid)
	if err != nil {
		return
	}

	// 编码角色
	role_sr := Net_RoleSendAndReceive{}
	mid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	role_sr.RoleBody, err = nst.StructGobBytes(mid)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	role_sr.RoleID = roleid
	// 编码传输的代码
	role_sr_b, err := nst.StructGobBytes(role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	err = o.sendWriteToServer(roleid, OPERATE_WRITE_ROLE, role_sr_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	return
}

// 删除一个角色
func (o *Operator) DeleteRole(id string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 编码角色id
	roleid_b := []byte(id)
	err = o.sendWriteToServer(id, OPERATE_DEL_ROLE, roleid_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DeleteRole: %v", err)
	}
	return
}

// 写入一个父角色
func (o *Operator) WriteFather(id, father string) (err error) {
	err = o.checkDMZ(id, father)
	if err != nil {
		return
	}

	// 生成发送数据
	role_father := Net_RoleFatherChange{
		Id:     id,
		Father: father,
	}
	//编码发送的数据
	role_father_b, err := nst.StructGobBytes(role_father)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFather: %v", err)
		return
	}
	// 送出
	err = o.sendWriteToServer(id, OPERATE_SET_FATHER, role_father_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFather: %v", err)
	}
	return
}

// 重置父角色
func (o *Operator) ResetFather(id string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 生成发送数据
	role_father := Net_RoleFatherChange{
		Id:     id,
		Father: "",
	}
	//编码发送的数据
	role_father_b, err := nst.StructGobBytes(role_father)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetFather: %v", err)
		return
	}
	// 送出
	err = o.sendWriteToServer(id, OPERATE_SET_FATHER, role_father_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetFather: %v", err)
	}
	return
}

// 读取父角色
func (o *Operator) ReadFather(id string) (father string, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 生成发送数据
	role_father := Net_RoleFatherChange{
		Id: id,
	}
	//编码发送的数据
	role_father_b, err := nst.StructGobBytes(role_father)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadFather: %v", err)
		return
	}
	// 发送并接收数据
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_FATHER, role_father_b, &role_father)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadFather: %v", err)
		return
	}
	father = role_father.Father
	return
}

// 读取全部子角色
func (o *Operator) ReadChildren(id string) (children []string, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 编码角色id
	roleid_b := []byte(id)
	// 发送并接收
	children = make([]string, 0)
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_CHILDREN, roleid_b, &children)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadChildren: %v", err)
	}
	return
}

// 写入全部子角色
func (o *Operator) WriteChildren(id string, children []string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 生成要发送的数据
	role_children := Net_RoleAndChildren{
		Id:       id,
		Children: children,
	}
	// 编码
	role_children_b, err := nst.StructGobBytes(role_children)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChildren: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_CHILDREN, role_children_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChildren: %v", err)
	}
	return
}

// 重置所有子角色
func (o *Operator) ResetChildren(id string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 生成要发送的数据
	role_children := Net_RoleAndChildren{
		Id:       id,
		Children: make([]string, 0),
	}
	// 编码
	role_children_b, err := nst.StructGobBytes(role_children)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetChildren: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_CHILDREN, role_children_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetChildren: %v", err)
	}
	return
}

// 设置子角色
func (o *Operator) WriteChild(id, child string) (err error) {
	err = o.checkDMZ(id, child)
	if err != nil {
		return
	}

	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChild: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_ADD_CHILD, role_child_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChild: %v", err)
	}
	return
}

// 删除一个子角色
func (o *Operator) DeleteChild(id, child string) (err error) {
	err = o.checkDMZ(id, child)
	if err != nil {
		return
	}

	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChild: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_DEL_CHILD, role_child_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteChild: %v", err)
	}
	return
}

// 是否存在一个子角色
func (o *Operator) ExistChild(id, child string) (have bool, err error) {
	err = o.checkDMZ(id, child)
	if err != nil {
		return
	}

	role_child := Net_RoleAndChild{
		Id:    id,
		Child: child,
	}
	role_child_b, err := nst.StructGobBytes(role_child)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ExistChild: %v", err)
		return
	}
	// 发送并接收返回
	_, err = o.sendReadAndDecodeData(id, OPERATE_EXIST_CHILD, role_child_b, &role_child)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ExistChild: %v", err)
		return
	}
	have = role_child.Exist
	return
}

// 读出所有的朋友角色
func (o *Operator) ReadFriends(id string) (friends map[string]roles.Status, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	// 编码角色id
	roleid_b := []byte(id)
	// 发送并接收返回
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_FRIENDS, roleid_b, &friends)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadFriends: %v", err)
	}
	return
}

// 写入全部朋友关系
func (o *Operator) WriteFriends(id string, friends map[string]roles.Status) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_friends := Net_RoleAndFriends{
		Id:      id,
		Friends: friends,
	}
	role_friends_b, err := nst.StructGobBytes(role_friends)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFriends: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_FRIENDS, role_friends_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFriends: %v", err)
	}
	return
}

// 重置全部朋友关系
func (o *Operator) ResetFriends(id string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_friends := Net_RoleAndFriends{
		Id:      id,
		Friends: make(map[string]roles.Status),
	}
	role_friends_b, err := nst.StructGobBytes(role_friends)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetFriends: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_FRIENDS, role_friends_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetFriends: %v", err)
	}
	return
}

// 创建空的上下文
func (o *Operator) CreateContext(id, contextname string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]CreateContext: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_ADD_CONTEXT, role_context_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]CreateContext: %v", err)
	}
	return
}

// 删除一个上下文
func (o *Operator) DropContext(id, contextname string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DropContext: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_DROP_CONTEXT, role_context_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DropContext: %v", err)
	}
	return
}

// 读取一个上下文
func (o *Operator) ReadContext(id, contextname string) (context roles.Context, have bool, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext_Data{
		Id:      id,
		Context: contextname,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContext: %v", err)
		return
	}
	// 发送
	_, err = o.sendReadAndDecodeData(id, OPERATE_READ_CONTEXT, role_context_b, &role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContext: %v", err)
		return
	}
	context = role_context.ContextBody
	have = role_context.Exist
	return
}

// 删除一个上下文中的绑定
func (o *Operator) DeleteContextBind(id, contextname string, upordown roles.ContextUpDown, bindrole string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindrole,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DeleteContextBind: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_DEL_CONTEXT_BIND, role_context_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DeleteContextBind: %v", err)
	}
	return
}

// 返回某个上下文的同样绑定值的所有
func (o *Operator) ReadContextSameBind(id, contextname string, upordown roles.ContextUpDown, bind int64) (rolesid []string, have bool, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		UpOrDown: upordown,
		Int:      bind,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextSameBind: %v", err)
		return
	}
	// 发送
	_, err = o.sendReadAndDecodeData(id, OPERATE_SAME_BIND_CONTEXT, role_context_b, &role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextSameBind: %v", err)
		return
	}
	rolesid = role_context.Gather
	have = role_context.Exist
	return
}

// 返回所有上下文的名称
func (o *Operator) ReadContextsName(id string) (names []string, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_context := Net_RoleAndContext_Data{
		Id: id,
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextsName: %v", err)
		return
	}
	// 发送
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_CONTEXTS_NAME, role_context_b, &role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextsName: %v", err)
		return
	}
	names = role_context.Gather
	return
}

// 设置所有上下文
func (o *Operator) WriteContexts(id string, contexts map[string]roles.Context) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_contexts := Net_RoleAndContexts{
		Id:       id,
		Contexts: contexts,
	}
	role_contexts_b, err := nst.StructGobBytes(role_contexts)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteContexts: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_CONTEXTS, role_contexts_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteContexts: %v", err)
	}
	return
}

// 重置上下文
func (o *Operator) ResetContexts(id string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_contexts := Net_RoleAndContexts{
		Id:       id,
		Contexts: make(map[string]roles.Context),
	}
	role_contexts_b, err := nst.StructGobBytes(role_contexts)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetContexts: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_CONTEXTS, role_contexts_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ResetContexts: %v", err)
	}
	return
}

// 读取全部上下文
func (o *Operator) ReadContexts(id string) (contexts map[string]roles.Context, err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_contexts := Net_RoleAndContexts{
		Id: id,
	}
	role_contexts_b, err := nst.StructGobBytes(role_contexts)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContexts: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_CONTEXTS, role_contexts_b, &role_contexts)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContexts: %v", err)
		return
	}
	contexts = role_contexts.Contexts
	return
}

// 设置朋友的属性
func (o *Operator) WriteFriendStatus(id, friend string, bindbit int, value interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]WriteFriendStatus: %v", e)
			return
		}
	}()

	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
		Bit:    bindbit,
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int":
		role_friend.Int = valuer.Int()
		role_friend.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		role_friend.Int = valuer.Int()
		role_friend.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		role_friend.Float = valuer.Float()
		role_friend.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		role_friend.Float = valuer.Float()
		role_friend.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		role_friend.Complex = valuer.Complex()
		role_friend.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		role_friend.Complex = valuer.Complex()
		role_friend.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		err = fmt.Errorf("drule[Operator]WriteFriendStatus: The value's type must int64, float64 or complex128.")
		return
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFriendStatus: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_FRIEND_STATUS, role_friend_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteFriendStatus: %v", err)
		return
	}
	return
}

// 写一个朋友
func (o *Operator) WriteFriend(id, friend string, bind int64) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	err = o.WriteFriendStatus(id, friend, 0, bind)
	return
}

// 删除一个朋友
func (o *Operator) DeleteFriend(id, friend string) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DeleteFriend: %v", err)
		return
	}
	err = o.sendWriteToServer(id, OPERATE_DEL_FRIEND, role_friend_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]DeleteFriend: %v", err)
	}
	return
}

// 读取朋友的状态
func (o *Operator) ReadFriendStatus(id, friend string, bindbit int, value interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]ReadFriendStatus: %v", e)
			return
		}
	}()

	role_friend := Net_RoleAndFriend{
		Id:     id,
		Friend: friend,
		Bit:    bindbit,
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		role_friend.Single = roles.STATUS_VALUE_TYPE_INT
	case "float64":
		role_friend.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex128":
		role_friend.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		err = fmt.Errorf("drule[Operator]ReadFriendStatus: The value's type must int64, float64 or complex128.")
		return
	}
	role_friend_b, err := nst.StructGobBytes(role_friend)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadFriendStatus: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_FRIEND_STATUS, role_friend_b, &role_friend)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadFriendStatus: %v", err)
		return
	}
	switch role_friend.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		valuer.SetInt(role_friend.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		valuer.SetFloat(role_friend.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		valuer.SetComplex(role_friend.Complex)
	default:
		err = fmt.Errorf("drule[Operator]ReadFriendStatus: The value's type not int64, float64 or complex128.")
	}
	return
}

// 设置一个上下文的属性
func (o *Operator) WriteContextStatus(id, contextname string, upordown roles.ContextUpDown, bindroleid string, bindbit int, value interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]WriteContextStatus: %v", e)
			return
		}
	}()

	role_context := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		Bit:      bindbit,
		UpOrDown: upordown,
		BindRole: bindroleid,
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int":
		role_context.Int = valuer.Int()
		role_context.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		role_context.Int = valuer.Int()
		role_context.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		role_context.Float = valuer.Float()
		role_context.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		role_context.Float = valuer.Float()
		role_context.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		role_context.Complex = valuer.Complex()
		role_context.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		role_context.Complex = valuer.Complex()
		role_context.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		err = fmt.Errorf("drule[Operator]WriteContextStatus: The value's type must int64, float64 or complex128.")
		return
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteContextStatus: %v", err)
		return
	}
	// 发送
	err = o.sendWriteToServer(id, OPERATE_SET_CONTEXT_STATUS, role_context_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteContextStatus: %v", err)
		return
	}
	return
}

// 读取一个上下文的状态
func (o *Operator) ReadContextStatus(id, contextname string, upordown roles.ContextUpDown, bindroleid string, bindbit int, value interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]ReadContextStatus: %v", e)
			return
		}
	}()

	role_context := Net_RoleAndContext_Data{
		Id:       id,
		Context:  contextname,
		Bit:      bindbit,
		UpOrDown: upordown,
		BindRole: bindroleid,
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		role_context.Single = roles.STATUS_VALUE_TYPE_INT
	case "float64":
		role_context.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex128":
		role_context.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		err = fmt.Errorf("drule[Operator]ReadContextStatus: The value's type must int64, float64 or complex128.")
		return
	}
	role_context_b, err := nst.StructGobBytes(role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextStatus: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_CONTEXT_STATUS, role_context_b, &role_context)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadContextStatus: %v", err)
		return
	}
	switch role_context.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		valuer.SetInt(role_context.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		valuer.SetFloat(role_context.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		valuer.SetComplex(role_context.Complex)
	default:
		err = fmt.Errorf("drule[Operator]ReadContextStatus: The value's type not int64, float64 or complex128.")
	}
	return
}

// 写一个数据
func (o *Operator) WriteData(id, name string, data interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]ReadData: %v", err)
		}
	}()

	data_reflect := reflect.Indirect(reflect.ValueOf(data))
	typename := data_reflect.Type().String()

	data_b, err := nst.StructGobBytes(data)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteData: %v", err)
		return
	}
	role_data := Net_RoleData_Data{
		Id:   id,
		Name: name,
		Type: typename,
		Data: data_b,
	}
	role_data_b, err := nst.StructGobBytes(role_data)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteData: %v", err)
		return
	}
	err = o.sendWriteToServer(id, OPERATE_SET_DATA, role_data_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]WriteData: %v", err)
	}
	return
}

// 读取一个数据
func (o *Operator) ReadData(id, name string, data interface{}) (err error) {
	err = o.checkDMZ(id)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("drule[Operator]ReadData: %v", err)
		}
	}()

	data_reflect := reflect.Indirect(reflect.ValueOf(data))
	typename := data_reflect.Type().String()

	role_data := Net_RoleData_Data{
		Id:   id,
		Name: name,
		Type: typename,
	}

	role_data_b, err := nst.StructGobBytes(role_data)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadData: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_GET_DATA, role_data_b, &role_data)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadData: %v", err)
		return
	}

	err = nst.BytesGobReflect(role_data.Data, data_reflect)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadData: %v", err)
	}
	return
}
