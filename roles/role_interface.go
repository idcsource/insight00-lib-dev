// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 角色（Role）概念封装的数据存储与数据关系
//
// Roleer
//
// Roller是一个“角色”。
//
// RoleMiddleData
//
// “角色”的中间存储格式
package roles

// 基础的角色接口，提供基本的角色间关系存储
type Roleer interface {
	// 新建自己的方法
	New(id string)
	// 获取自己的版本
	Version() int
	// 设置自己的版本
	SetVersion(version int)

	// 返回角色自身的ID
	ReturnId() string
	// 获取父角色
	GetFather() string
	// 获取所有子角色
	GetChildren() []string
	// 获取所有朋友角色
	GetFriends() map[string]Status

	// 重置父关系，也就是把父亲弄没
	ResetFather()
	// 重置子关系，也就是把孩子弄没
	ResetChilren()
	// 重置朋友关系，也就是把朋友弄没
	ResetFriends()

	// 设定父关系
	SetFather(id string)
	// 设置全部子关系
	SetChildren([]string)
	// 设置全部朋友关系
	SetFriends(map[string]Status)

	// 是否存在某个子角色
	ExistChild(id string) (have bool)
	// 是否存在某个朋友，并且返回这个朋友的疏离关系
	ExistFriend(id string) (have bool, bind int64)

	// 将一个子角色添加进去
	AddChild(id string) error
	// 删除一个子角色
	DeleteChild(id string) error

	// 添加一个朋友关系
	AddFriend(id string, bind int64) error
	// 删除一个朋友关系
	DeleteFriend(id string) error
	// 修改一个朋友关系，只是改关系远近
	ChangeFriend(id string, bind int64) error
	// 获取相同远近关系下的所有朋友的ID
	GetSameBindFriendsId(bind int64) []string

	// 获取删除状态
	ReturnDelete() bool
	// 设置删除状态
	SetDelete(del bool)

	// 返回某个部分是否被改变
	ReturnChanged(status uint8) bool
	// 设置数据体被改变
	SetDataChanged()

	// 创建一个空的上下文
	NewContext(contextname string) (err error)
	// 是否含有这个上下文
	ExistContext(contextname string) (have bool)
	// 设定一个上下文的上游
	AddContextUp(contextname, upname string, bind int64)
	// 设定一个上下文的下游
	AddContextDown(contextname, downname string, bind int64)
	// 删除一个上下文的上游
	DelContextUp(contextname, upname string)
	// 删除一个上下文的下游
	DelContextDown(contextname, downname string)
	// 清除一个上下文
	DelContext(contextname string)
	// 找到一个上下文的上文，返回绑定值
	GetContextUp(contextname, upname string) (bind int64, have bool)
	// 找到一个上下文的下文，返回绑定值
	GetContextDown(contextname, downname string) (bind int64, have bool)
	// 返回某个上下文的全部信息
	GetContext(contextname string) (context Context, have bool)
	// 返回某个上下文中的上游同样绑定值的所有
	GetContextUpSameBind(contextname string, bind int64) (rolesid []string, have bool)
	// 返回某个上下文中的下游同样绑定值的所有
	GetContextDownSameBind(contextname string, bind int64) (rolesid []string, have bool)
	// 返回所有上下文组的名称
	GetContextsName() (names []string)

	// 设置朋友的状态属性
	SetFriendStatus(id string, bit int, value interface{}) (err error)
	// 获取朋友的状态属性
	GetFriendStatus(id string, bit int, value interface{}) (err error)
	// 设置上下文的状态属性
	SetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (err error)
	// 获取上下文的状态属性
	GetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (err error)

	// 设定上下文
	SetContexts(context map[string]Context)
	// 获取上下文
	GetContexts() map[string]Context
}
