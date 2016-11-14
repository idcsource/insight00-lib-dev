// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this souDMODE_MASTERrce code is governed by GNU LGPL v3 license

package drcm

import (
	"sync"

	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 锆存储。
// 这是一个带缓存的存储体系，基本功能方面可以看作是hardstore与rcontrol的合并（虽然还有很大不同），而增强方面它支持分布式存储。
type ZrStorage struct {
	rolesplus.RolePlus
	
	/* 下面这部分是存储相关的 */
	
	// 配置信息
	config				*cpool.Block
	// 本地存储
	local_store			*hardstore.HardStore
	
	/* 下面这部分是缓存相关的 */
	
	// 角色缓存
	rolesCache				map[string]oneRoleCache
	// 最大缓存角色数
	cacheMax				int64
	// 缓存数量
	rolesCount				int64
	// 缓存满的触发
	cacheIsFull				chan bool
	// 删除缓存
	deleteCache				[]string
	// 检查缓存数量中
	checkCacheNumOn			bool
	
	/* 下面是分布式服务相关的 */
	
	// 分布式服务的模式，来自于常量DMODE_*
	dmode					uint8
	// 自身的身份码，做服务的时候使用
	code					string
	// 请求slave执行或返回数据的连接，string为slave对应的管理第一个值的首字母，而那个切片则是做镜像的
	slaves			map[string][]*slaveIn
	// 监听的实例
	listen			*nst.TcpServer
	// slave的连接池，从这里分配给slaveIn
	slavepool		map[string]*nst.TcpClient
	
	// 日志
	logs					*ilogs.Logs
	// 全局锁
	lock					*sync.RWMutex
}

// 一个角色的缓存，提供了锁
type oneRoleCache struct {
	lock					*sync.RWMutex
	role					roles.Roleer
}

// 一台从机的信息
type slaveIn struct {
	name string
	code string
	tcpconn *nst.TcpClient
}

// 前缀状态，每次向slave发信息都要先把这个状态发出去
type Net_PrefixStat struct {
	// 操作类型，从OPERATE_*
	Operate		int
	// 身份验证码
	Code		string
}

// slave回执，slave收到PrefixStat之后的第一步返回信息
type Net_SlaveReceipt struct {
	// 数据状态，来自DATA_*
	DataStat	uint8
	// 返回的错误
	Error		error
}

// 角色的接收与发送格式
type Net_RoleSendAndReceive struct {
	// 角色的身体
	RoleBody	[]byte
	// 角色的关系
	RoleRela	[]byte
}

// 角色的father修改的数据格式
type Net_RoleFatherChange struct {
	Id			string
	Father		string
}
