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
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 这是DRule——分布式统治者
type DRule struct {
	// 配置信息
	config *cpool.Block
	// 事务统治者
	trule *TRule
	// 自己的名字
	selfname string
	// 已经关闭
	closed bool

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

	// 日志
	logs *ilogs.Logs
}

// 一台从机的信息
type slaveIn struct {
	name    string
	code    string
	tcpconn *nst.TcpClient
}

/* 下面是网络传输所需要的结构 */

// 对事务的数据
type Net_Transaction struct {
	// 事务的ID
	TransactionId string
	// 是否在事务中
	InTransaction bool
	// 准备的角色ID
	PrepareIDs []string
	// 请求什么操作，见DRULE_TRAN_*
	AskFor uint8
}

// 前缀状态，每次向slave发信息都要先把这个状态发出去
type Net_PrefixStat struct {
	// 操作类型，从OPERATE_*
	Operate int
	// 客户端名称
	ClientName string
	// 身份验证码
	Code string
	// 在事务中
	InTransaction bool
	// 事务ID
	TransactionId string
	// 涉及到的角色id
	RoleId string
}

// slave回执带数据体
type Net_SlaveReceipt_Data struct {
	// 数据状态，来自DATA_*
	DataStat uint8
	// 返回的错误
	Error string
	// 数据体
	Data []byte
}

// 角色的接收与发送格式
type Net_RoleSendAndReceive struct {
	// 角色的ID
	RoleID string
	// 是否存在
	IfHave bool
	// 角色的身体
	RoleBody []byte
	// 角色的关系
	RoleRela []byte
	// 角色的版本
	RoleVer []byte
}

// 角色的father修改的数据格式
type Net_RoleFatherChange struct {
	Id     string
	Father string
}

// 角色的所有子角色
type Net_RoleAndChildren struct {
	Id       string
	Children []string
}

// 角色的单个子角色关系的网络数据格式
type Net_RoleAndChild struct {
	Id    string
	Child string
	Exist bool
}

// 角色的所有朋友
type Net_RoleAndFriends struct {
	Id      string
	Friends map[string]roles.Status
}

// 角色的单个朋友角色关系的网络数据格式
type Net_RoleAndFriend struct {
	Id     string
	Friend string
	Bind   int64
	Status roles.Status
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single uint8
	// 单一的绑定修改所对应的位置，也就是0到9
	Bit int
	// 单一修改的Int
	Int int64
	// 单一修改的Float
	Float float64
	// 单一修改的Complex
	Complex complex128
}

// 角色的单个上下文关系的网络数据格式
type Net_RoleAndContext struct {
	Id string
	// 上下文的名字
	Context string
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown uint8
	// 要操作的绑定角色的ID
	BindRole string
}

// 角色的全部上下文
type Net_RoleAndContexts struct {
	Id       string
	Contexts map[string]roles.Context
}

// 角色的单个上下文关系数据的网络数据格式
type Net_RoleAndContext_Data struct {
	Id string
	// 上下文的名字
	Context string
	// 要求的上下文是否存在
	Exist bool
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown uint8
	// 要操作的绑定角色的ID
	BindRole string
	// 一个的状态位结构
	Status roles.Status
	// 上下文的结构
	ContextBody roles.Context
	// 名字等的集合
	Gather []string
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single uint8
	// 单一的绑定修改所对应的位置，也就是0到9
	Bit int
	// 单一修改的Int
	Int int64
	// 单一修改的Float
	Float float64
	// 单一修改的Complex
	Complex complex128
}

// 角色的单个数据的数据体的网络格式
type Net_RoleData_Data struct {
	Id string
	// 数据点的名字
	Name string
	// 数据的字节流
	Data []byte
}
