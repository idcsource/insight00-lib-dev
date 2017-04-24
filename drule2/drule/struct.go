// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 这是DRule——分布式统治者
type DRule struct {
	// 配置信息
	config *cpool.Block
	// 事务统治者
	trule *trule.TRule
	// 自己的名字
	selfname string

	// 已经关闭
	closed bool

	// 分布式服务模式，DMODE_*
	dmode uint8

	// 登录进来的用户
	loginuser map[string]*loginUser

	// 日志
	logs *ilogs.Logs
}

// 登录进来的用户
type loginUser struct {
	username   string
	unid       string
	authority  uint8
	activetime time.Time
}

// Drule和Operator的用户
type DRuleUser struct {
	roles.Role                 // 角色
	UserName   string          // 用户名
	Password   string          // 密码
	Email      string          // 邮箱
	Authority  uint8           // 权限，USER_AUTHORITY_*
	WRable     map[string]bool // 读写权限，string为区域的名称，bool为true则是写，为false则为读
}
