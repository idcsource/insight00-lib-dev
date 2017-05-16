// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 网络套接字通讯功能的封装“Network Socket Transmission”
//
// 本包提供了各种类型与[]byte类型之间的转换函数。
// 并提供了一套tcp服务器和客户端的实现。
//
// TcpClient
//
// TcpClient的使用进程分配功能的流程为（方法的内部流程）：
//	--> 使用*TcpClient.OpenProgress()分配一个连接
//		--> 从连接池里顺序找下一个连接，如果连接正在被占用则往下找，如果连接断开则尝试重新连接
//		--> 为这个连接加锁（使用chan实现的锁）
//		--> 发送NORMAL_DATA状态(*TcpClient.checkOneConn2()的检查方法中)
//		--> 返回*ProgressData
//	--> 使用*ProgressData.SendAndReturn([]byte)发送数据并接收服务端返回值和错误值
//		--> 发送DATA_GOON状态
//		--> 发送具体数据
//		<-- 接收返回的DATA_GOON状态 --或者：接受返回的DATA_CLOSE状态
//		<-- 接收具体返回数据 --或者：抛出字符为“DATA_CLOSE”的错误
//	--> 可再次使用*ProgressData.SendAndReturn([]byte)发送数据并接收服务端返回值和错误值
//		...除非DATA_CLOSE已经关闭
//		...
//	--> 使用*ProgressData.Close()关闭这个连接进程
//		--> 发送DATA_CLOSE状态
//		--> 释放这个连接的锁
//
// TcpClient的直接发送和接受，不经过进程分配的流程为（方法的内部流程）：
//	--> 使用*TcpClient.SendAndReturn([]byte)发送数据并接收服务端返回值和错误值
//		--> 使用*TcpClient.OpenProgress()分配一个连接
//		--> 使用*ProgressData.SendAndReturn([]byte)发送数据并接收服务端返回值和错误值
//		--> 使用*ProgressData.Close()关闭这个连接进程
//
// [TODO]未来会实现*ProgressData除SendAndReturn以外的方法
//
// TcpClient的长连接心跳维持及中断检测。
// 每30秒轮询一遍连接池中的连接，只要没有正在被*TcpClient.OpenProgress()分配，则执行：
//	--> 连接加锁（使用chan实现的锁），无法加锁则认为正在被使用，直接跳过
//	--> 发送HEART_BEAT状态
//	--> 如果发送不成功，则进行重新连接
//	--> 释放连接的锁
//
// TcpServer
//
// TcpServer完全配合TcpClient的心跳机制，以及DATA_GOON、DATA_CLOSE状态的执行。
//
// TcpServer需要接收一个符合nst.ConnExecer接口的执行者负责Client请求的执行，也就是需要提供ExecTCP(ce *ConnExec)方法。
//
// ConnExec是对nst.Tcp的封装，提供了简单直接的发送接收数据以及关闭连接的功能。
package nst

const (
	HEART_BEAT  = iota // 心跳
	NORMAL_DATA        // 普通数据
	CONN_CLOSE         // 连接断开
	DATA_GOON          // 数据继续
	DATA_CLOSE         // 数据关闭
)

// TcpServer的转交方法所需要符合的接口
type ConnExecer interface {
	ExecTCP(ce *ConnExec) error
}
