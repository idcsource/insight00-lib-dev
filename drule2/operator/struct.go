// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// drule2的远程操作者
package operator

import (
	"time"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

type tranService struct {
	unid   string            // 事务ID
	askfor TransactionAskFor // 请求操作
}

type operatorService struct {
	tran_signal chan tranService // 事务信号
}

// 这叫做“操作机”，是用来远程连接DRule的。
type Operator struct {
	selfname    string                   // 自己的名字
	drule       *druleInfo               // 服务器端
	login       bool                     // 是否在登陆状态，如果不是则是false
	service     *operatorService         // 操作者服务
	transaction map[string]*OTransaction // 事务列表，time是事务的活跃时间
	logs        *ilogs.Logs              // 日志
}

// 操作机事务
type OTransaction struct {
	selfname       string           // 自己的名字
	transaction_id string           // 事务id
	drule          *druleInfo       // 服务器端
	service        *operatorService // 操作者服务
	logs           *ilogs.Logs      // 日志
	bedelete       bool             // 如果为true则被删除
	activetime     time.Time        // 活跃日期
}

// 一台服务器的信息
type druleInfo struct {
	name        string         // 机器名称
	username    string         // 用户名
	password    string         // 密码
	unid        string         // 登录唯一码
	active_time time.Time      // 活跃日期
	tcpconn     *nst.TcpClient // tcp连接
}

/* 下面是网络传输所需要的结构 */

// Operator的发送
type O_OperatorSend struct {
	OperateZone   OperateZone  // 操作分区
	Operate       OperatorType // 操作类型，从OPERATE_*
	OperatorName  string       // 客户端名称
	InTransaction bool         // 在事务中
	TransactionId string       // 事务ID
	RoleId        string       // 涉及到的角色id
	AreaId        string       // 涉及到的区域ID
	User          string       // 登陆的用户名
	Unid          string       // 登录的Unid
	Data          []byte       // 数据体
}

// DRule回执带数据体
type O_DRuleReceipt struct {
	DataStat DRuleReturnStatus // 数据状态，来自DATA_*
	Error    string            // 返回的错误
	Data     []byte            // 数据体
}

// 对事务的数据
type O_Transaction struct {
	TransactionId string   // 事务的ID
	InTransaction bool     // 是否在事务中
	Area          string   // 区域
	PrepareIDs    []string // 准备的角色ID
}

// 角色的接收与发送格式
type O_RoleSendAndReceive struct {
	Area     string               // 区域
	RoleID   string               // 角色的ID
	IfHave   bool                 // 是否存在
	RoleBody roles.RoleMiddleData // 角色的身体
}

// 角色的father修改的数据格式
type O_RoleFatherChange struct {
	Area   string
	Id     string
	Father string
}

// 角色的所有子角色
type O_RoleAndChildren struct {
	Area     string
	Id       string
	Children []string
}

// 角色的单个子角色关系的网络数据格式
type O_RoleAndChild struct {
	Area  string
	Id    string
	Child string
	Exist bool
}

// 角色的所有朋友
type O_RoleAndFriends struct {
	Area    string
	Id      string
	Friends map[string]roles.Status
}

// 角色的单个朋友角色关系的网络数据格式
type O_RoleAndFriend struct {
	Area   string
	Id     string
	Friend string
	Bind   int64
	Status roles.Status
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single roles.StatusValueType
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
type O_RoleAndContext struct {
	Area string
	Id   string
	// 上下文的名字
	Context string
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown roles.ContextUpDown
	// 要操作的绑定角色的ID
	BindRole string
}

// 角色的全部上下文
type O_RoleAndContexts struct {
	Area     string
	Id       string
	Contexts map[string]roles.Context
}

// 角色的单个上下文关系数据的网络数据格式
type O_RoleAndContext_Data struct {
	Area string
	Id   string
	// 上下文的名字
	Context string
	// 要求的上下文是否存在
	Exist bool
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown roles.ContextUpDown
	// 要操作的绑定角色的ID
	BindRole string
	// 一个的状态位结构
	Status roles.Status
	// 上下文的结构
	ContextBody roles.Context
	// 名字等的集合
	Gather []string
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single roles.StatusValueType
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
type O_RoleData_Data struct {
	Area string
	Id   string
	// 数据点的名字
	Name string
	// 数据类型
	Type string
	// 数据的字节流
	Data []byte
}

// 区域
type O_Area struct {
	AreaName string
	Rename   string
	Exist    bool
}

// 来往网络的用户信息
type O_DRuleUser struct {
	UserName  string        // 用户名
	Password  string        // 密码
	Email     string        // 邮箱
	Authority UserAuthority // 权限，USER_AUTHORITY_*
	Unid      string        // 唯一码
}

// 用户和区域的权限
type O_Area_User struct {
	UserName string // 用户名
	Area     string // 区域名
	WRable   bool   // true为读权限，false为写权限
	Add      bool   // true为增加，false为减少

}

// 远端操作者的记录
type O_DRuleOperator struct {
	Name     string // 名称
	Address  string // 地址与端口
	ConnNum  int    // 连接数
	TLS      bool   // 是否加密
	Username string // 用户名
	Password string // 密码
}

// 蔓延到其他drule上的区域
type O_AreasRouter struct {
	AreaName string              // 区域名称
	Mirror   bool                // 是否为镜像，ture为镜像，则所有的文件都发给下面所有的drule
	Mirrors  []string            // string为drule的名字
	Chars    map[string][]string // 如果mirror为false，则看这个根据不同的字母进行路由，第一个stirng为首字母，第二个string为operator的名称
}
