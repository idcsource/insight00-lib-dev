// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Status Monitor & Configure Spread [ 状态监控与配置蔓延 ]。
// 中心与节点之间的状态与配置的相互通讯。
package smcs

import(
	"github.com/idcsource/Insight-0-0-lib/cpool"
)

const (
	NODE_TYPE_SPIDER	= "spider"										// 节点类型，蜘蛛爬虫
)

const (
	NODE_STATUS_NO_CONFIG		=		iota							// 节点状态，没有配置文件
	NODE_STATUS_OK														// 一切OK
	NODE_STATUS_BUSY													// 忙碌
	NODE_STATUS_IDLE													// 闲置
	NODE_STATUS_STORE_FULL												// 存储满
)

const (
	WORK_SET_NO					=		iota							// 没有这个节点
	WORK_SET_GOON														// 节点的工作设置，继续之前
	WORK_SET_START														// 开始工作
	WORK_SET_STOP														// 停止工作
)

const (
	SLEEP_TIME					= 60									// 每隔多长间隔发送一次，单位为秒
)

// 节点发送给中心的数据结构
type NodeSend struct {
	WorkSet		uint8													// 当前工作状态
	Type		string													// 节点的类型
	Name		string													// 节点的名称
	Status		uint8													// 状态
	RunLog		[]string												// 要发送出去的日志
	ErrLog		[]string												// 要发送出去的日志
}

// 中心发送给节点的数据结构
type CenterSend struct {
	NextWorkSet			uint8											// 下一个工作状态设置
	SetStartTime		int64											// 下一个工作状态的开始时间
	NewConfig			bool											// 是否有新配置文件
	Config				cpool.BlockEncode								// 配置文件
}
