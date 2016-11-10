// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 网络套接字通讯功能的封装“Network Socket Transmission”。
// 本包提供了各种类型与[]byte类型之间的转换函数。
// 并提供了一套tcp服务器和客户端的实现。
package nst

const (
	HEART_BEAT				= iota
	NORMAL_DATA
)

// TcpServer的转交方法所需要符合的接口
type ConnExecer interface {
	ExecTCP (tcp *TCP)
}
