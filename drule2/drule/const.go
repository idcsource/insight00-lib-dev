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

type RolePosition uint8

// 角色保存位置
const (
	ROLE_POSITION_IN_LOCAL  RolePosition = iota // 角色保存在本地
	ROLE_POSITION_IN_REMOTE                     // 角色保存在远程
)
