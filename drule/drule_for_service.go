// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

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

// 为readSometing的读取角色（事务版）
func (d *DRule) readRoleTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_sr := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(byte_slice_data, &role_sr)
	if err != nil {
		return
	}
	// 执行
	role, err := tran.ReadRole(role_sr.RoleID)
	if err != nil {
		return
	}
	// 编码角色
	role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer, err = hardstore.EncodeRole(role)
	if err != nil {
		return
	}
	// 编码发送
	return_data, err = nst.StructGobBytes(role_sr)
	return
}

// 为readSometing的读取角色（非事务版）
func (d *DRule) readRoleNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_sr := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(byte_slice_data, &role_sr)
	if err != nil {
		return
	}
	// 执行
	role, err := d.trule.ReadRole(role_sr.RoleID)
	if err != nil {
		return
	}
	// 编码角色
	role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer, err = hardstore.EncodeRole(role)
	if err != nil {
		return
	}
	// 编码发送
	return_data, err = nst.StructGobBytes(role_sr)
	return
}

// 为writeSometing的存一个角色（事务版）
func (d *DRule) storeRoleTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_sr := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(byte_slice_data, &role_sr)
	if err != nil {
		return
	}
	// 执行
	// 还原角色
	role, err := hardstore.DecodeRole(role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer)
	if err != nil {
		return err
	}
	err = tran.StoreRole(role)
	return
}

// 为writeSometing的存一个角色（非事务版）
func (d *DRule) storeRoleNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_sr := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(byte_slice_data, &role_sr)
	if err != nil {
		return
	}
	// 执行
	// 还原角色
	role, err := hardstore.DecodeRole(role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer)
	if err != nil {
		return err
	}
	err = d.trule.StoreRole(role)
	return
}

// 为writeSometing的删除一个角色（事务版）
func (d *DRule) deleteRoleTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_id := string(byte_slice_data)
	// 执行
	err = tran.DeleteRole(role_id)
	return
}

// 为writeSometing的删除一个角色（非事务版）
func (d *DRule) deleteRoleNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_id := string(byte_slice_data)
	// 执行
	err = d.trule.DeleteRole(role_id)
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

// 为readSometing的读取某个上下文（非事务版）
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

// 为writeSometing的设定一个值（事务版）
func (d *DRule) writeDataTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_data := Net_RoleData_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_data)
	if err != nil {
		return
	}
	// 执行
	err = tran.writeDataFromByte(role_data.Id, role_data.Name, role_data.Data)
	return
}

// 为writeSometing的设定一个值（非事务版）
func (d *DRule) writeDataNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_data := Net_RoleData_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_data)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.writeDataFromByte(role_data.Id, role_data.Name, role_data.Data)
	return
}

// 为readSometing的读一个值（事务版）
func (d *DRule) readDataTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_data := Net_RoleData_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_data)
	if err != nil {
		return
	}
	// 执行
	err = tran.readDataToByte(&role_data)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_data)
	return
}

// 为readSometing的读一个值（非事务版）
func (d *DRule) readDataNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_data := Net_RoleData_Data{}
	err = nst.BytesGobStruct(byte_slice_data, &role_data)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.readDataToByte(&role_data)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_data)
	return
}

// 为writeSometing的设置全部上下文（事务版）
func (d *DRule) writeContextsTran(tran *Transaction, byte_slice_data []byte) (err error) {
	// 解码
	role_contexts := Net_RoleAndContexts{}
	err = nst.BytesGobStruct(byte_slice_data, &role_contexts)
	if err != nil {
		return
	}
	// 执行
	err = tran.WriteContexts(role_contexts.Id, role_contexts.Contexts)
	return
}

// 为writeSometing的设置全部上下文（非事务版）
func (d *DRule) writeContextsNoTran(byte_slice_data []byte) (err error) {
	// 解码
	role_contexts := Net_RoleAndContexts{}
	err = nst.BytesGobStruct(byte_slice_data, &role_contexts)
	if err != nil {
		return
	}
	// 执行
	err = d.trule.WriteContexts(role_contexts.Id, role_contexts.Contexts)
	return
}

// 为readSometing的返回所有上下文（事务版）
func (d *DRule) readContextsTran(tran *Transaction, byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_contexts := Net_RoleAndContexts{}
	err = nst.BytesGobStruct(byte_slice_data, &role_contexts)
	if err != nil {
		return
	}
	// 执行
	role_contexts.Contexts, err = tran.ReadContexts(role_contexts.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_contexts)
	return
}

// 为readSometing的返回所有上下文（非事务版）
func (d *DRule) readContextsNoTran(byte_slice_data []byte) (return_data []byte, err error) {
	// 解码
	role_contexts := Net_RoleAndContexts{}
	err = nst.BytesGobStruct(byte_slice_data, &role_contexts)
	if err != nil {
		return
	}
	// 执行
	role_contexts.Contexts, err = d.trule.ReadContexts(role_contexts.Id)
	if err != nil {
		return
	}
	// 编码
	return_data, err = nst.StructGobBytes(role_contexts)
	return
}
