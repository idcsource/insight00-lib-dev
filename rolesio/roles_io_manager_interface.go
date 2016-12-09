// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 角色（Role）的永久存储接口与空实现
package rolesio

import(
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// RolesInOutManager 是一个角色永久存储的读写接口。
// 接口定义了一系列方法，为了将角色的信息与关系保存到永久的硬存储（如磁盘）上。
type RolesInOutManager interface {
	// 从永久存储读出一个角色
	ReadRole (id string) (roles.Roleer, error)
	// 往永久存储写入一个角色
	StoreRole (role roles.Roleer) error
	// 从永久存储删除一个角色
	DeleteRole (id string) error

	// 设置父角色
	WriteFather (id, father string) error
	// 获取父角色
	ReadFather (id string) (father string, err error)
	// 重置父角色
	ResetFather (id string) error

	// 读取角色的所有子角色的角色名
	ReadChildren (id string) ([]string, error)
	// 写入为name的角色的所有子角色名
	WriteChildren (id string, children []string) error
	// 删除name的角色的所有子角色（重置）
	ResetChildren (id string) error

	// 往永久存储里加入一个子角色的关系
	WriteChild (id, child string) error
	// 在永久存储里删除一个子角色的关系
	DeleteChild (id, child string) error
	// 查询是否有这个子角色
	ExistChild (id, child string) (have bool, err error)

	// 读取name角色的所有朋友关系
	ReadFriends (id string) (map[string]roles.Status, error)
	// 写入name角色的所有朋友关系
	WriteFriends (id string, friends map[string]roles.Status) error
	// 删除name角色的所有朋友关系（重置）
	ResetFriends (id string) error
	// 获取相同远近关系下的所有朋友的ID
	ReadSameBindFriendsId (id string, bind int64) ([]string, error)

	// 往永久存储里加入一个朋友关系，并绑定，已有关系将只是修改绑定值
	WriteFriend (id, friend string, bind int64) error
	// 在永久存储里删除一个朋友关系
	DeleteFriend (id, friend string) error
	// 是否存在朋友关系，并返回绑定值
	ExistFriend (id, friend string) (bind int64, have bool, err error)

	// 创建一个空的上下文，如果已经存在则忽略
	CreateContext (id, contextname string) error
	// 清除一个上下文，也就是删除它
	DropContext (id, contextname string) error
	// 是否含有某个上下文
	//ExistContext (id, contextname string) (have bool, err error)
	// 返回某个上下文的全部信息
	ReadContext (id, contextname string) (context roles.Context, have bool, err error)
	// 删除一个上下文绑定，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN，bindrole是对应绑定的角色id
	DeleteContextBind (id, contextname string, upordown uint8, bindrole string) (err error)
	// 返回某个上下文中的同样绑定值的所有，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
	ReadContextSameBind (id, contextname string, upordown uint8, bind int64) (rolesid []string, have bool, err error)
	// 返回所有上下文组的名称
	ReadContextsName () (names []string, err error)

	// 设置朋友的状态属性
	WriteFriendStatus (id, friends string, bindbit int, value interface{}) (err error)
	// 获取朋友的状态属性
	ReadFriendStatus (id, friends string, bindbit int, value interface{}) (err error)
	// 设置上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
	WriteContextStatus (id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error)
	// 获取上下文的状态属性，upordown为roles中的CONTEXT_UP或CONTEXT_DOWN
	ReadContextStatus (id, contextname string, upordown uint8, bindroleid string, bindbit int, value interface{}) (err error)

	// 设定上下文
	WriteContexts (id string, context map[string]roles.Context) error
	// 获取上下文
	ReadContexts (id string) (map[string]roles.Context, error)
	// 重置上下文
	ResetContexts (id string) error

	// 把data的数据装入role的name值下，如果找不到name，则返回错误。
	WriteData (id, name string, data interface{}) (err error)
	// 从角色中知道name的数据名并返回其数据。
	ReadData (id, name string) (data interface{}, err error)
}
