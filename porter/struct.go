// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 搬运工——将角色从一个地方原封不动的搬到另一个地方
//
// 本package使用符合rolesio.RolesInOutManager接口的存储器对角色进行保存，并使用nst所提供的网络支持。
//
//本package使用hardstore对角色进行编码和解码。也就是说我们需要gob.Register()对roles.Roleer接口的类型进行注册。
//
// Receiver
//
// 这是货物的接收者，它监听一个网络端口并对收到的角色进行处理。
//
// 它需要两个配置选项，具体如下：
// 	# 监听端口
//	listen = 9998
//	# 身份验证码
//	code = ReceiverCode
// 所有配置选项都提供方法在运行时更改，包括换用新的监听接口。
//
// 收到角色后的处理方法不依靠配置文件指定，需要自己注册。
// 如果要直接进行保存则使用SetStorage()方法注册一个符合rolesio.RolesInOutManager接口的角色存储装置。
// 如果需要交由某个外部函数处理，则使用SetOperater()方法注册一个符合ReceiverOperater接口的外部实例。
//
// 一定要注意，上面的注册方法，永远都只会生效一个，且是最后注册的那个。
//
// Sender
//
// 这是货物的发送者，将角色发送给接收者。
//
// 它需要一个*cpool.Block类型的配置选项，具体示例如下：
//	# main为必不可少的节点
//	[main]
//	# 发送者名字
//	name = IamAsender
//	# 接收者配置列表名，对应下面具体接收者的配置信息
//	receiver = R01,R02
//
//	[R01]
//	# 接收者访问地址
//	address = 192.168.1.101:11111
//	# 接收者的身份验证码
//	code = ReceiverCode
//	# 接收者的连接数
//	conn_num = 5
//
//	[R02]
//	# 接收者访问地址
//	address = 192.168.1.102:11111
//	# 接收者的身份验证码
//	code = ReceiverCode
//	# 接收者的连接数
//	conn_num = 5
// 发送者提供最简单的将角色发送出去的方法：SendRole、SendRoleToReceiver。
// 第一个会将角色发给所有连接的接收者，而第二个则只会发送给某一个接收者。
package porter

import (
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 接收者接收后的处理方法，保存还是交由处理函数处理
const (
	OPERATE_NOT_SET = iota
	OPERATE_TO_STORE
	OPERATE_TO_FUNCTION
)

// 来往的数据状态
const (
	DATA_NOTHING = iota
	// 数据并不是期望的
	DATA_NOT_EXPECT
	// 请发送数据
	DATA_PLEASE
	// 数据一切OK
	DATA_ALL_OK
)

// 货物，这是搬运工需要搬运的东西，也就是编码后的需要传递的角色信息，这里的编码使用hardstore所提供的方法
type Cargo struct {
	// 角色Id
	Id string
	// 发送者名字
	SenderName string
	// 接收者code
	ReceiverCode string
	// 角色的身体
	RoleBody []byte
	// 角色的关系
	RoleRela []byte
	// 角色的版本
	RoleVer []byte
}

// 接收者
type Receiver struct {
	// 配置信息
	config *cpool.Section
	// 监听实例
	listen *nst.TcpServer
	// 身份验证码
	code string
	// 处理方法，OPERATE_TO_*
	optype uint8
	// 存储器实例
	store rolesio.RolesInOutManager
	// 注册的处理者
	function ReceiverOperater
	// 日志
	logs *ilogs.Logs
}

// 发送者
type Sender struct {
	// 配置信息
	config *cpool.Block
	// 发送者名字
	name string
	// 接收者
	receivers map[string]oneReceiver
	// 日志
	logs *ilogs.Logs
}

// 一个接收者的信息
type oneReceiver struct {
	// 名称
	name string
	// 身份验证码
	code string
	// 地址
	address string
	// 连接
	tcpconn *nst.TcpClient
}

// 接收者的注册处理方法的接口
type ReceiverOperater interface {
	Operate(sendername string, role roles.Roleer) (err error)
}

// slave回执，slave收到PrefixStat之后的第一步返回信息
type Net_ReceiverReceipt struct {
	// 数据状态，来自DATA_*
	DataStat uint8
	// 返回的错误
	Error string
}
