// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"bytes"

	"github.com/idcsource/insight00-lib/iendecode"
	"github.com/idcsource/insight00-lib/spots"
)

/* 下面是网络传输所需要的结构 */

// Operator的发送（后面需要跟具体的数据）
type O_OperatorSend struct {
	OperateZone   OperateZone  // 操作分区
	Operate       OperatorType // 操作类型，从OPERATE_*
	OperatorName  string       // 客户端名称
	InTransaction bool         // 在事务中
	TransactionId string       // 事务ID
	SpotId        string       // 涉及到的Spot id
	AreaId        string       // 涉及到的区域ID
	User          string       // 登陆的用户名
	Unid          string       // 登录的Unid
}

func (o O_OperatorSend) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// 8
	operate_zone_b := iendecode.UintToBytes(uint(o.OperateZone))
	buf.Write(operate_zone_b)

	// 8
	operate_b := iendecode.UintToBytes(uint(o.Operate))
	buf.Write(operate_b)

	// OperatorName
	operator_name_b := []byte(o.OperatorName)
	operator_name_b_len := len(operator_name_b)
	buf.Write(iendecode.IntToBytes(operator_name_b_len))
	buf.Write(operator_name_b)

	// InTransaction
	in_transaction_b := iendecode.BoolToBytes(o.InTransaction)
	buf.Write(in_transaction_b)

	if o.InTransaction == true {
		// TransactionId(40 or 0)
		transaction_id_b := []byte(o.TransactionId)
		buf.Write(transaction_id_b)
	}

	// SpotId
	spot_id_b := []byte(o.SpotId)
	spot_id_b_len := len(spot_id_b)
	buf.Write(iendecode.IntToBytes(spot_id_b_len))
	buf.Write(spot_id_b)

	// AreaId
	area_id_b := []byte(o.AreaId)
	area_id_b_len := len(area_id_b)
	buf.Write(iendecode.IntToBytes(area_id_b_len))
	buf.Write(area_id_b)

	// User
	user_b := []byte(o.User)
	user_b_len := len(user_b)
	buf.Write(iendecode.IntToBytes(user_b_len))
	buf.Write(user_b)

	// Unid(40)
	unid_b := []byte(o.Unid)
	unid_b_len := len(unid_b)
	buf.Write(iendecode.IntToBytes(unid_b_len))
	buf.Write(unid_b)

	data = buf.Bytes()
	return
}

func (o *O_OperatorSend) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	// 8
	operate_zone_b := buf.Next(8)
	o.OperateZone = OperateZone(iendecode.BytesToUint(operate_zone_b))

	// 8
	operate_b := buf.Next(8)
	o.Operate = OperatorType(iendecode.BytesToUint(operate_b))

	// OperatorName
	operator_name_b_len := iendecode.BytesToInt(buf.Next(8))
	operator_name_b := buf.Next(operator_name_b_len)
	o.OperatorName = string(operator_name_b)

	// InTransaction
	o.InTransaction = iendecode.BytesToBool(buf.Next(1))

	if o.InTransaction == true {
		// TransactionId
		transaction_id_b := buf.Next(40)
		o.TransactionId = string(transaction_id_b)
	}

	// SpotId
	spot_id_b_len := iendecode.BytesToInt(buf.Next(8))
	spot_id_b := buf.Next(spot_id_b_len)
	o.SpotId = string(spot_id_b)

	// AreaId
	area_id_b_len := iendecode.BytesToInt(buf.Next(8))
	area_id_b := buf.Next(area_id_b_len)
	o.AreaId = string(area_id_b)

	// User
	user_b_len := iendecode.BytesToInt(buf.Next(8))
	user_b := buf.Next(user_b_len)
	o.User = string(user_b)

	// Unid(40)
	unid_b_len := iendecode.BytesToInt(buf.Next(8))
	unid_b := buf.Next(unid_b_len)
	o.Unid = string(unid_b)

	return
}

// DRule回执带数据体（后面要跟具体的数据）
type O_DRuleReceipt struct {
	DataStat DRuleReturnStatus // 数据状态，来自DATA_*
	Error    string            // 返回的错误
}

// 对事务的数据
type O_Transaction struct {
	TransactionId string   // 事务的ID
	InTransaction bool     // 是否在事务中
	Area          string   // 区域
	PrepareIDs    []string // 准备的角色ID
}

// 角色的接收与发送格式
type O_SpotSendAndReceive struct {
	Area   string       // 区域
	RoleID string       // 角色的ID
	IfHave bool         // 是否存在
	Spot   *spots.Spots // 角色的身体
}

// 角色的father修改的数据格式
type O_SpotFatherChange struct {
	Area   string
	Id     string
	Father string
}

// 角色的所有子角色
type O_SpotAndChildren struct {
	Area     string
	Id       string
	Children []string
}

// 角色的单个子角色关系的网络数据格式
type O_SpotAndChild struct {
	Area  string
	Id    string
	Child string
	Exist bool
}

// 角色的所有朋友
type O_SpotAndFriends struct {
	Area    string
	Id      string
	Friends map[string]spots.Status
}

// 角色的单个朋友角色关系的网络数据格式
type O_SpotAndFriend struct {
	Area   string
	Id     string
	Friend string
	Bind   int64
	// 要求的是否存在
	Exist  bool
	Status spots.Status
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single spots.StatusValueType
	// 单一的绑定修改所对应的位置，也就是0到9
	Bit int
	// 单一修改的Int
	Int int64
	// 单一修改的Float
	Float float64
	// 单一修改的Complex
	Complex complex128
	// 单一修改的string
	String string
}

// 角色的单个上下文关系的网络数据格式
type O_SpotAndContext struct {
	Area string
	Id   string
	// 上下文的名字
	Context string
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown spots.ContextUpDown
	// 要操作的绑定角色的ID
	BindRole string
	// 存在否
	Exist bool
}

// 角色的全部上下文
type O_SpotAndContexts struct {
	Area     string
	Id       string
	Contexts map[string]spots.Context
}

// 角色的单个上下文关系数据的网络数据格式
type O_SpotAndContext_Data struct {
	Area string
	Id   string
	// 上下文的名字
	Context string
	// 要求的上下文是否存在
	Exist bool
	// 这是roles包中的CONTEXT_UP或CONTEXT_DOWN
	UpOrDown spots.ContextUpDown
	// 要操作的绑定角色的ID
	BindRole string
	// 一个的状态位结构
	Status spots.Status
	// 上下文的结构
	ContextBody spots.Context
	// 名字等的集合
	Gather []string
	// 单一的绑定属性修改，1为int，2为float，3为complex
	Single spots.StatusValueType
	// 单一的绑定修改所对应的位置，也就是0到9
	Bit int
	// 单一修改的Int
	Int int64
	// 单一修改的Float
	Float float64
	// 单一修改的Complex
	Complex complex128
	// 单一修改的string
	String string
}

// 角色的单个数据的数据体的网络格式
type O_SpotData_Data struct {
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
