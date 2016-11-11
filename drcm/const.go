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
	OPREATE_TOSTORE
	
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
	// 重置children
	OPERATE_RESET_CHILDREN
	
	// 添加一个child
	OPERATE_ADD_CHILD
	// 删除一个child
	OPERATE_DEL_CHILD
	
	// 设置friends
	OPERATE_SET_FRIENDS
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
	OPREATE_DEL_CONTEXT
	// 获取所有上下文的名称
	OPERATE_GET_CONTEXTS_NAME
	
	// 添加一个上文
	OPERATE_ADD_CONTEXT_UP
	// 删除上文
	OPREATE_DEL_CONTEXT_UP
	// 修改上文
	OPREATE_CHANGE_CONTEXT_UP
	// 返回同样绑定的上文
	OPREATE_SAME_BIND_CONTEXT_UP
	
	// 添加一个下文
	OPERATE_ADD_CONTEXT_DOWN
	// 删除下文
	OPREATE_DEL_CONTEXT_DOWN
	// 修改下文
	OPREATE_CHANGE_CONTEXT_DOWN
	// 返回同样绑定的下文
	OPREATE_SAME_BIND_CONTEXT_DOWN
	
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
