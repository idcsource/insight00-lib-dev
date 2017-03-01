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
package porter

import (
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 货物，这是搬运工需要搬运的东西，也就是编码后的需要传递的角色信息，这里的编码使用hardstore所提供的方法
type Cargo struct {
	// 角色Id
	Id string
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
	config *cpool.Block
	// 监听实例
	listen *nst.TcpServer
	// 身份验证码
	code string
	// 存储器实例
	store rolesio.RolesInOutManager
}

// 发送者
type Sender struct {
	// 配置信息
	config *cpool.Block
	// 接收者
	receivers map[string]oneReceiver
}

// 一个接收者的信息
type oneReceiver struct {
	// 名称
	name string
	// 身份验证码
	code string
	// 连接
	tcpconn *nst.TcpClient
}
