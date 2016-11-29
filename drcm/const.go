// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

// 操作方式列表
const (
	// 什么操作都没有
	OPERATE_NOTHING					= iota
	// 强制保存
	OPERATE_TOSTORE

	// 获取一个角色
	OPERATE_READ_ROLE
	// 写入一个角色
	OPERATE_WRITE_ROLE

	// 创建一个新角色
	OPERATE_NEW_ROLE
	// 删除一个角色
	OPERATE_DEL_ROLE

	// 获取角色的一个值
	OPERATE_GET_DATA
	// 设置角色的一个值
	OPERATE_SET_DATA

	// 设置father
	OPERATE_SET_FATHER
	// 获取father
	OPERATE_GET_FATHER
	// 重置father
	OPERATE_RESET_FATHER

	// 设置children
	OPERATE_SET_CHILDREN
	// 获取children
	OPERATE_GET_CHILDREN
	// 重置children
	OPERATE_RESET_CHILDREN

	// 添加一个child
	OPERATE_ADD_CHILD
	// 删除一个child
	OPERATE_DEL_CHILD
	// 存在某个child
	OPERATE_EXIST_CHILD

	// 设置friends
	OPERATE_SET_FRIENDS
	// 获取friends
	OPERATE_GET_FRIENDS
	// 重置friends
	OPERATE_RESET_FRIENDS

	// 添加一个friend
	OPERATE_ADD_FRIEND
	// 删除一个friend
	OPERATE_DEL_FRIEND
	// 修改一个friend
	OPERATE_CHANGE_FRIEND
	// 获取同样绑定值的friend
	OPERATE_SAME_BIND_FRIEND

	// 添加一个空的上下文组
	OPERATE_ADD_CONTEXT
	// 删除一个上下文组
	OPERATE_DROP_CONTEXT
	// 获取所有上下文的名称
	OPERATE_GET_CONTEXTS_NAME

	// 添加一个上文
	OPERATE_ADD_CONTEXT_UP
	// 删除上文
	OPERATE_DEL_CONTEXT_UP
	// 修改上文
	OPERATE_CHANGE_CONTEXT_UP
	// 返回同样绑定的上文
	OPERATE_SAME_BIND_CONTEXT_UP

	// 添加一个下文
	OPERATE_ADD_CONTEXT_DOWN
	// 删除下文
	OPERATE_DEL_CONTEXT_DOWN
	// 修改下文
	OPERATE_CHANGE_CONTEXT_DOWN
	// 返回同样绑定的下文
	OPERATE_SAME_BIND_CONTEXT_DOWN

	// 设置朋友的状态
	OPERATE_SET_FRIEND_STATUS
	// 获取朋友的状态
	OPERATE_GET_FRIEND_STATUS

	// 设置上下文的状态
	OPERATE_SET_CONTEXT_STATUS
	// 获取上下文的状态
	OPERATE_GET_CONTEXT_STATUS
)

// 分布式模式
const (
	// 没有分布式，只有自己
	DMODE_OWN					= iota
	// 在分布式里做master
	DMODE_MASTER
	// 在分布式里做slave
	DMODE_SLAVE
)

// 连接模式
const (
	// 连接为本地存储
	CONN_IS_LOCAL				= iota
	// 连接为slave
	CONN_IS_SLAVE
)

// 数据标记状态
const (
	// 数据没有任何的状态
	DATA_NOTHING				= iota
	// 数据并不是期望的
	DATA_NOT_EXPECT
	// 数据一切正常
	DATA_ALL_OK
	// 数据终止
	DATA_END
	// 请发送数据
	DATA_PLEASE
	// 数据将发送
	DATA_WILL_SEND
	// 数据返回有错误
	DATA_RETURN_ERROR
	// 数据返回为True
	DATA_RETURN_IS_TRUE
	// 束缚返回为flase
	DATA_RETURN_IS_FALSE
)
