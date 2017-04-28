// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

// 内部数据
const (
	INSIDE_DMZ = "_DRule" // 内部隔离区，一个区域

	USER_PREFIX        = "_user_" // 用户的角色id前缀
	ROOT_USER          = "root"   // 根用户用户名
	ROOT_USER_PASSWORD = "123456" // 根用户默认密码

	OPERATOR_PREFIX = "_operator_"     // 远程控制者的角色id前缀
	OPERATOR_ROOT   = "_root_operator" // 远程控制者的存储根角色id

	AREA_DRULE_PREFIX = "_area_drule_"     // 路由蔓延规则的角色id前缀
	AREA_DRULE_ROOT   = "_root_area_drule" // 路由蔓延规则的根角色id
)

// 工作模式
type OperateMode uint8

const (
	OPERATE_MODE_SLAVE  OperateMode = iota // 从机模式
	OPERATE_MODE_MASTER                    // 主机模式
)

// 角色保存位置，或叫连接模式
const (
	// 连接为本地存储
	CONN_IS_LOCAL = iota
	// 连接为网络
	CONN_IS_NET
)
