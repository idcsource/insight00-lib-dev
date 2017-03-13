// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 这是DRule——分布式统治者
type DRule struct {
	// 配置信息
	config *cpool.Block
	// 事务统治者
	trule *TRule

	// 连接服务
	connect *druleConnectService

	// 日志
	logs *ilogs.Logs
}

// drule的连接服务
type druleConnectService struct {
	// 分布式服务模式，DMODE_*
	dmode uint8
	// 自身的身份码，slave和master时需要
	code string
	// 请求slave执行或返回数据的连接，string为slave对应的管理第一个值的首字母，而那个切片则是做镜像的
	slaves map[string][]*slaveIn
	// 监听的实例，slave下或master下
	listen *nst.TcpServer
	// slave的连接池，从这里分配给slaveIn
	slavepool map[string]*nst.TcpClient
	// slave的slaveIn连接池
	slavecpool map[string]*slaveIn
}

// drule的事务模式
type druleTransaction struct {
	// 事务的id
	unid string
	// 事务
	transaction *Transaction
	// 连接服务
	connect *druleConnectService
}

// 一台从机的信息
type slaveIn struct {
	name    string
	code    string
	tcpconn *nst.TcpClient
}
