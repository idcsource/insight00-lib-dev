// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

// 内部保护区
const (
	INSIDE_DMZ = "_DRULE_INSIDE_DMZ" // 内部隔离区
)

// 分布式模式
const (
	// 在分布式里做slave
	DMODE_SLAVE = iota
	// 在分布式里做master
	DMODE_MASTER
)

// 角色保存位置，或叫连接模式
const (
	// 连接为本地存储
	CONN_IS_LOCAL = iota
	// 连接为网络
	CONN_IS_NET
)
