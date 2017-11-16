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
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

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

func (o *O_DRuleReceipt) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// datastat 8
	buf.Write(iendecode.UintToBytes(uint(o.DataStat)))

	// error
	error_b := []byte(o.Error)
	error_b_len := len(error_b)
	buf.Write(iendecode.IntToBytes(error_b_len))
	buf.Write(error_b)

	data = buf.Bytes()
	return
}

func (o *O_DRuleReceipt) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	o.DataStat = DRuleReturnStatus(iendecode.BytesToUint(buf.Next(8)))

	error_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Error = string(buf.Next(error_b_len))

	return
}

// 对事务的数据
type O_Transaction struct {
	TransactionId string   // 事务的ID
	InTransaction bool     // 是否在事务中
	Area          string   // 区域
	PrepareIDs    []string // 准备的角色ID
}

func (o O_Transaction) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// InTransaction 1
	InTransaction_b := iendecode.BoolToBytes(o.InTransaction)
	buf.Write(InTransaction_b)

	// TransactionId (40 or 0)
	if o.InTransaction == true {
		buf.Write([]byte(o.TransactionId))
	}

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// PrepareIDs
	prepare_ids_b := iendecode.SliceStringToBytes(o.PrepareIDs)
	prepare_ids_b_len := len(prepare_ids_b)
	buf.Write(iendecode.IntToBytes(prepare_ids_b_len))
	buf.Write(prepare_ids_b)

	data = buf.Bytes()
	return
}

func (o *O_Transaction) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// InTransaction 1
	o.InTransaction = iendecode.BytesToBool(buf.Next(1))

	// TransactionId (40 or 0)
	if o.InTransaction == true {
		o.TransactionId = string(buf.Next(40))
	}

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// PrepareIDs
	prepare_ids_b_len := iendecode.BytesToInt(buf.Next(8))
	o.PrepareIDs = iendecode.BytesToSliceString(buf.Next(prepare_ids_b_len))

	return
}

// Spot的接收与发送格式（后面跟真正的Spot数据）
type O_SpotSendAndReceive struct {
	IfHave bool   // 是否存在
	Area   string // 区域
	SpotId string // Spot的ID
}

func (o O_SpotSendAndReceive) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// IfHave 1
	buf.Write(iendecode.BoolToBytes(o.IfHave))

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	data = buf.Bytes()
	return
}

func (o *O_SpotSendAndReceive) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// IfHave 1
	o.IfHave = iendecode.BytesToBool(buf.Next(1))

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	return
}

// Spot的father修改的数据格式
type O_SpotFatherChange struct {
	Area   string
	SpotId string
	Father string
}

func (o O_SpotFatherChange) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Father
	father_b := []byte(o.Father)
	father_b_len := len(father_b)
	buf.Write(iendecode.IntToBytes(father_b_len))
	buf.Write(father_b)

	data = buf.Bytes()
	return
}

func (o *O_SpotFatherChange) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Father
	father_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Father = string(buf.Next(father_b_len))

	return
}

// Spot的所有子Spot
type O_SpotAndChildren struct {
	Area     string
	SpotId   string
	Children []string
}

func (o O_SpotAndChildren) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Children
	children_b := iendecode.SliceStringToBytes(o.Children)
	children_b_len := len(children_b)
	buf.Write(iendecode.IntToBytes(children_b_len))
	buf.Write(children_b)

	data = buf.Bytes()
	return
}

func (o *O_SpotAndChildren) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Children
	children_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Children = iendecode.BytesToSliceString(buf.Next(children_b_len))

	return
}

// Spot的单个子Spot关系的网络数据格式
type O_SpotAndChild struct {
	Area   string
	SpotId string
	Child  string
	Exist  bool
}

func (o O_SpotAndChild) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Child
	child_b := []byte(o.Child)
	child_b_len := len(child_b)
	buf.Write(iendecode.IntToBytes(child_b_len))
	buf.Write(child_b)

	// Exist 1
	buf.Write(iendecode.BoolToBytes(o.Exist))

	return buf.Bytes(), err
}

func (o *O_SpotAndChild) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Child
	child_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Child = string(buf.Next(child_b_len))

	// Exist 1
	o.Exist = iendecode.BytesToBool(buf.Next(1))

	return
}

// Spot的所有朋友
type O_SpotAndFriends struct {
	Area    string
	SpotId  string
	Friends map[string]spots.Status
}

func (o O_SpotAndFriends) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Friends
	thecount := len(o.Friends)
	buf.Write(iendecode.IntToBytes(thecount))
	for key, _ := range o.Friends {

		key_b := []byte(key)
		key_b_len := len(key_b)
		buf.Write(iendecode.IntToBytes(key_b_len))
		buf.Write(key_b)

		var value_b []byte
		value_b, err = o.Friends[key].MarshalBinary()
		if err != nil {
			return
		}
		value_b_len := len(value_b)
		buf.Write(iendecode.IntToBytes(value_b_len))
		buf.Write(value_b)
	}

	return buf.Bytes(), err
}

func (o *O_SpotAndFriends) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	o.Friends = make(map[string]spots.Status)
	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Friends
	thecount := iendecode.BytesToInt(buf.Next(8))
	for i := 0; i < thecount; i++ {
		key_b_len := iendecode.BytesToInt(buf.Next(8))
		key := string(buf.Next(key_b_len))

		value_b_len := iendecode.BytesToInt(buf.Next(8))
		value := spots.NewStatus()
		err = value.UnmarshalBinary(buf.Next(value_b_len))
		if err != nil {
			return
		}
		o.Friends[key] = value
	}

	return
}

// Spot的单个朋友关系的网络数据格式
type O_SpotAndFriend struct {
	Area   string
	SpotId string
	Friend string
	Exist  bool // 要求的是否存在
}

func (o O_SpotAndFriend) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Friend
	friend_b := []byte(o.Friend)
	friend_b_len := len(friend_b)
	buf.Write(iendecode.IntToBytes(friend_b_len))
	buf.Write(friend_b)

	// Exist 1
	buf.Write(iendecode.BoolToBytes(o.Exist))

	return buf.Bytes(), err
}

func (o *O_SpotAndFriend) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Friend
	friend_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Friend = string(buf.Next(friend_b_len))

	// Exist 1
	o.Exist = iendecode.BytesToBool(buf.Next(1))

	return
}

type O_SpotAndFriend_Data struct {
	Single  spots.StatusValueType // 单一的绑定属性修改，1为int，2为float，3为complex
	Status  spots.Status
	Bit     int        // 单一的绑定修改所对应的位置，也就是0到9
	Int     int64      // 单一修改的Int
	Float   float64    // 单一修改的Float
	Complex complex128 // 单一修改的Complex
	String  string     // 单一修改的string
}

func (o O_SpotAndFriend_Data) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Single 1
	single_b := iendecode.Uint8ToBytes(uint8(o.Single))
	buf.Write(single_b)

	if o.Single == spots.STATUS_VALUE_TYPE_NULL {
		// Status
		var status_b []byte
		status_b, err = o.Status.MarshalBinary()
		if err != nil {
			return
		}
		status_b_len := len(status_b)
		buf.Write(iendecode.IntToBytes(status_b_len))
		buf.Write(status_b)
	} else {
		// Bit 8
		buf.Write(iendecode.IntToBytes(o.Bit))
		if o.Single == spots.STATUS_VALUE_TYPE_INT {
			// Int 8
			buf.Write(iendecode.Int64ToBytes(o.Int))
		} else if o.Single == spots.STATUS_VALUE_TYPE_FLOAT {
			// Float 8
			var float_b []byte
			float_b, err = iendecode.ToBinary(o.Float)
			if err != nil {
				return
			}
			buf.Write(float_b)
		} else if o.Single == spots.STATUS_VALUE_TYPE_COMPLEX {
			// Complex 16
			var complex_b []byte
			complex_b, err = iendecode.ToBinary(o.Complex)
			if err != nil {
				return
			}
			buf.Write(complex_b)
		} else {
			// String
			string_b := []byte(o.String)
			string_b_len := len(string_b)
			buf.Write(iendecode.IntToBytes(string_b_len))
			buf.Write(string_b)
		}
	}

	return buf.Bytes(), err
}

func (o *O_SpotAndFriend_Data) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Single 1
	o.Single = spots.StatusValueType(iendecode.BytesToUint8(buf.Next(1)))

	if o.Single == spots.STATUS_VALUE_TYPE_NULL {
		// Status
		status_b_len := iendecode.BytesToInt(buf.Next(8))
		status := spots.NewStatus()
		err = status.UnmarshalBinary(buf.Next(status_b_len))
		if err != nil {
			return
		}
		o.Status = status
	} else {
		// Bit 8
		o.Bit = iendecode.BytesToInt(buf.Next(8))
		if o.Single == spots.STATUS_VALUE_TYPE_INT {
			// Int 8
			o.Int = iendecode.BytesToInt64(buf.Next(8))
		} else if o.Single == spots.STATUS_VALUE_TYPE_FLOAT {
			// Float 8
			err = iendecode.FromBinary(buf.Next(8), &o.Float)
		} else if o.Single == spots.STATUS_VALUE_TYPE_COMPLEX {
			// Complex 16
			err = iendecode.FromBinary(buf.Next(16), &o.Complex)
		} else {
			// String
			str_b_len := iendecode.BytesToInt(buf.Next(8))
			o.String = string(buf.Next(str_b_len))
		}
	}

	return
}

// Spot的单个上下文关系的网络数据格式
type O_SpotAndContext struct {
	Area     string
	SpotId   string
	Context  string              // 上下文的名字
	UpOrDown spots.ContextUpDown // 这是spots包中的CONTEXT_UP或CONTEXT_DOWN
	BindSpot string              // 要操作的绑定Spot的ID
	Exist    bool                // 存在否
}

func (o O_SpotAndContext) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Context
	context_b := []byte(o.Context)
	context_b_len := len(context_b)
	buf.Write(iendecode.IntToBytes(context_b_len))
	buf.Write(context_b)

	// UpOrDown uint8 1
	buf.Write(iendecode.Uint8ToBytes(uint8(o.UpOrDown)))

	// BindSpot
	bindspot_b := []byte(o.BindSpot)
	bindspot_b_len := len(bindspot_b)
	buf.Write(iendecode.IntToBytes(bindspot_b_len))
	buf.Write(bindspot_b)

	// Exist bool 1
	buf.Write(iendecode.BoolToBytes(o.Exist))

	return buf.Bytes(), err
}

func (o *O_SpotAndContext) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Context
	context_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Context = string(buf.Next(context_b_len))

	// UpOrDown uint8 1
	o.UpOrDown = spots.ContextUpDown(iendecode.BytesToUint8(buf.Next(1)))

	// BindSpot
	bindspot_b_len := iendecode.BytesToInt(buf.Next(8))
	o.BindSpot = string(buf.Next(bindspot_b_len))

	// Exist 1
	o.Exist = iendecode.BytesToBool(buf.Next(1))

	return
}

// Spot的单个上下文关系数据的网络数据格式
type O_SpotAndContext_Data struct {
	Type        uint8                 // 1为单一绑定，2为单一Status，3为一个Context，4为Gather
	Single      spots.StatusValueType // 单一的绑定属性修改
	Status      spots.Status          // 一个的状态位结构
	ContextBody spots.Context         // 上下文的结构
	Gather      []string              // 名字等的集合
	Bit         int                   // 单一的绑定修改所对应的位置，也就是0到9
	Int         int64                 // 单一修改的Int
	Float       float64               // 单一修改的Float
	Complex     complex128            // 单一修改的Complex
	String      string                // 单一修改的string
}

func (o O_SpotAndContext_Data) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Type 1
	type_b := iendecode.Uint8ToBytes(o.Type)
	buf.Write(type_b)

	if o.Type == 1 {
		// Bit 8
		buf.Write(iendecode.IntToBytes(o.Bit))
		if o.Single == spots.STATUS_VALUE_TYPE_INT {
			// Int 8
			buf.Write(iendecode.Int64ToBytes(o.Int))
		} else if o.Single == spots.STATUS_VALUE_TYPE_FLOAT {
			// Float 8
			var float_b []byte
			float_b, err = iendecode.ToBinary(o.Float)
			if err != nil {
				return
			}
			buf.Write(float_b)
		} else if o.Single == spots.STATUS_VALUE_TYPE_COMPLEX {
			// Complex 16
			var complex_b []byte
			complex_b, err = iendecode.ToBinary(o.Complex)
			if err != nil {
				return
			}
			buf.Write(complex_b)
		} else {
			// String
			string_b := []byte(o.String)
			string_b_len := len(string_b)
			buf.Write(iendecode.IntToBytes(string_b_len))
			buf.Write(string_b)
		}
	} else if o.Type == 2 {
		// Status
		var status_b []byte
		status_b, err = o.Status.MarshalBinary()
		if err != nil {
			return
		}
		status_b_len := len(status_b)
		buf.Write(iendecode.IntToBytes(status_b_len))
		buf.Write(status_b)
	} else if o.Type == 3 {
		// ContextBody
		var context_b []byte
		context_b, err = o.ContextBody.MarshalBinary()
		if err != nil {
			return
		}
		context_b_len := len(context_b)
		buf.Write(iendecode.IntToBytes(context_b_len))
		buf.Write(context_b)
	} else if o.Type == 4 {
		// Gather
		gather_b := iendecode.SliceStringToBytes(o.Gather)
		gather_b_len := len(gather_b)
		buf.Write(iendecode.IntToBytes(gather_b_len))
		buf.Write(gather_b)
	}

	return buf.Bytes(), err
}

func (o *O_SpotAndContext_Data) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Single 1
	o.Type = iendecode.BytesToUint8(buf.Next(1))

	if o.Type == 1 {
		// Bit 8
		o.Bit = iendecode.BytesToInt(buf.Next(8))
		if o.Single == spots.STATUS_VALUE_TYPE_INT {
			// Int 8
			o.Int = iendecode.BytesToInt64(buf.Next(8))
		} else if o.Single == spots.STATUS_VALUE_TYPE_FLOAT {
			// Float 8
			err = iendecode.FromBinary(buf.Next(8), &o.Float)
		} else if o.Single == spots.STATUS_VALUE_TYPE_COMPLEX {
			// Complex 16
			err = iendecode.FromBinary(buf.Next(16), &o.Complex)
		} else {
			// String
			str_b_len := iendecode.BytesToInt(buf.Next(8))
			o.String = string(buf.Next(str_b_len))
		}
	} else if o.Type == 2 {
		// Status
		status_b_len := iendecode.BytesToInt(buf.Next(8))
		o.Status = spots.NewStatus()
		err = o.Status.UnmarshalBinary(buf.Next(status_b_len))
		if err != nil {
			return
		}
	} else if o.Type == 3 {
		// ContextBody
		context_b_len := iendecode.BytesToInt(buf.Next(8))
		o.ContextBody = spots.NewContext()
		err = o.ContextBody.UnmarshalBinary(buf.Next(context_b_len))
		if err != nil {
			return
		}
	} else if o.Type == 4 {
		// Gather
		gather_b_len := iendecode.BytesToInt(buf.Next(8))
		o.Gather = iendecode.BytesToSliceString(buf.Next(gather_b_len))
	}

	return
}

// Spot的全部上下文
type O_SpotAndContexts struct {
	Area     string
	SpotId   string
	Contexts map[string]spots.Context
}

func (o O_SpotAndContexts) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Contexts
	contexts_count := len(o.Contexts)
	buf.Write(iendecode.IntToBytes(contexts_count))
	for key, _ := range o.Contexts {
		key_b := []byte(key)
		key_b_len := len(key_b)
		buf.Write(iendecode.IntToBytes(key_b_len))
		buf.Write(key_b)

		var value_b []byte
		value_b, err = o.Contexts[key].MarshalBinary()
		if err != nil {
			return
		}
		value_b_len := len(value_b)
		buf.Write(iendecode.IntToBytes(value_b_len))
		buf.Write(value_b)
	}

	return buf.Bytes(), err
}

func (o *O_SpotAndContexts) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Contexts
	o.Contexts = make(map[string]spots.Context)
	thecount := iendecode.BytesToInt(buf.Next(8))
	for i := 0; i < thecount; i++ {
		key_b_len := iendecode.BytesToInt(buf.Next(8))
		key := string(buf.Next(key_b_len))
		thecontext := spots.NewContext()
		value_b_len := iendecode.BytesToInt(buf.Next(8))
		err = thecontext.UnmarshalBinary(buf.Next(value_b_len))
		if err != nil {
			return
		}
		o.Contexts[key] = thecontext
	}

	return
}

// Spot的单个数据的数据体的网络格式
type O_SpotData struct {
	Area   string
	SpotId string
	Name   string // 数据点的名字
	Data   []byte // 数据的字节流
}

func (o O_SpotData) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Area
	area_b := []byte(o.Area)
	area_b_len := len(area_b)
	buf.Write(iendecode.IntToBytes(area_b_len))
	buf.Write(area_b)

	// SpotId
	spotid_b := []byte(o.SpotId)
	spotid_b_len := len(spotid_b)
	buf.Write(iendecode.IntToBytes(spotid_b_len))
	buf.Write(spotid_b)

	// Name
	name_b := []byte(o.Name)
	name_b_len := len(name_b)
	buf.Write(iendecode.IntToBytes(name_b_len))
	buf.Write(name_b)

	// Data
	data_len := len(o.Data)
	buf.Write(iendecode.IntToBytes(data_len))
	buf.Write(o.Data)

	return buf.Bytes(), err
}

func (o *O_SpotData) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// SpotId
	spotid_b_len := iendecode.BytesToInt(buf.Next(8))
	o.SpotId = string(buf.Next(spotid_b_len))

	// Name
	name_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Name = string(buf.Next(name_b_len))

	// Data
	data_len := iendecode.BytesToInt(buf.Next(8))
	o.Data = buf.Next(data_len)

	return
}

// 区域
type O_Area struct {
	AreaName string
	Rename   string
	Exist    bool
}

func (o O_Area) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// AreaName
	area_name_b := []byte(o.AreaName)
	area_name_b_len := len(area_name_b)
	buf.Write(iendecode.IntToBytes(area_name_b_len))
	buf.Write(area_name_b)

	// Rename
	rename_b := []byte(o.Rename)
	rename_b_len := len(rename_b)
	buf.Write(iendecode.IntToBytes(rename_b_len))
	buf.Write(rename_b)

	// Exist bool 1
	buf.Write(iendecode.BoolToBytes(o.Exist))

	return buf.Bytes(), err
}

func (o *O_Area) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// AreaName
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.AreaName = string(buf.Next(area_b_len))

	// Rename
	rename_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Rename = string(buf.Next(rename_b_len))

	// Exist
	o.Exist = iendecode.BytesToBool(buf.Next(1))

	return
}

// 来往网络的用户信息
type O_DRuleUser struct {
	UserName  string        // 用户名
	Password  string        // 密码
	Email     string        // 邮箱
	Authority UserAuthority // 权限，USER_AUTHORITY_*
	Unid      string        // 唯一码
}

type O_DRuleUserKeep struct {
	UserName string // 用户名
	Unid     string // 唯一码(sha1 40)
}

func (o O_DRuleUserKeep) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// UserName
	username_b := []byte(o.UserName)
	username_b_len := len(username_b)
	buf.Write(iendecode.IntToBytes(username_b_len))
	buf.Write(username_b)

	// Unid sha1 40
	buf.Write([]byte(o.Unid))

	return buf.Bytes(), err
}

func (o *O_DRuleUserKeep) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// UserName
	username_b_len := iendecode.BytesToInt(buf.Next(8))
	o.UserName = string(buf.Next(username_b_len))

	// Unid sha1 40
	o.Unid = string(buf.Next(40))

	return
}

type O_DRuleUserLog struct {
	UserName string // 用户名
	Password string // 密码(sha1 40)
}

func (o O_DRuleUserLog) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// UserName
	username_b := []byte(o.UserName)
	username_b_len := len(username_b)
	buf.Write(iendecode.IntToBytes(username_b_len))
	buf.Write(username_b)

	// Password sha1 40
	buf.Write([]byte(o.Password))

	return buf.Bytes(), err
}

func (o *O_DRuleUserLog) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// UserName
	username_b_len := iendecode.BytesToInt(buf.Next(8))
	o.UserName = string(buf.Next(username_b_len))

	// Password sha1 40
	o.Password = string(buf.Next(40))

	return
}

type O_DRuleUserReg struct {
	UserName string // 用户名
	Password string // 密码(sha1 40)
	Email    string // 邮箱
}

func (o O_DRuleUserReg) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// UserName
	username_b := []byte(o.UserName)
	username_b_len := len(username_b)
	buf.Write(iendecode.IntToBytes(username_b_len))
	buf.Write(username_b)

	// Password sha1 40
	buf.Write([]byte(o.Password))

	// Email
	email_b := []byte(o.Email)
	email_b_len := len(email_b)
	buf.Write(iendecode.IntToBytes(email_b_len))
	buf.Write(email_b)

	return buf.Bytes(), err
}

func (o *O_DRuleUserReg) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// UserName
	username_b_len := iendecode.BytesToInt(buf.Next(8))
	o.UserName = string(buf.Next(username_b_len))

	// Password sha1 40
	o.Password = string(buf.Next(40))

	// Email
	email_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Email = string(buf.Next(email_b_len))

	return
}

// 用户和区域的权限
type O_Area_User struct {
	UserName string // 用户名
	Area     string // 区域名
	Add      bool   // true为增加，false为减少
	WRable   bool   // true为读权限，false为写权限
}

func (o O_Area_User) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// UserName
	username_b := []byte(o.UserName)
	username_b_len := len(username_b)
	buf.Write(iendecode.IntToBytes(username_b_len))
	buf.Write(username_b)

	// Area
	area_name_b := []byte(o.Area)
	area_name_b_len := len(area_name_b)
	buf.Write(iendecode.IntToBytes(area_name_b_len))
	buf.Write(area_name_b)

	// Add 1
	buf.Write(iendecode.BoolToBytes(o.Add))

	// WRable 1
	buf.Write(iendecode.BoolToBytes(o.WRable))

	return buf.Bytes(), err
}

func (o *O_Area_User) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Username
	username_b_len := iendecode.BytesToInt(buf.Next(8))
	o.UserName = string(buf.Next(username_b_len))

	// Area
	area_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Area = string(buf.Next(area_b_len))

	// Add 1
	o.Add = iendecode.BytesToBool(buf.Next(1))

	// WRable 1
	o.Add = iendecode.BytesToBool(buf.Next(1))

	return
}

// 远端操作者的记录
type O_DRuleOperator struct {
	Name     string // 名称
	Address  string // 地址与端口
	ConnNum  int    // 连接数
	TLS      bool   // 是否加密
	UserName string // 用户名
	Password string // 密码
}

func (o O_DRuleOperator) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Name
	name_b := []byte(o.Name)
	name_b_len := len(name_b)
	buf.Write(iendecode.IntToBytes(name_b_len))
	buf.Write(name_b)

	// Address
	address_b := []byte(o.Address)
	address_b_len := len(address_b)
	buf.Write(iendecode.IntToBytes(address_b_len))
	buf.Write(address_b)

	// ConnNum 8
	buf.Write(iendecode.IntToBytes(o.ConnNum))

	// TLS 1
	buf.Write(iendecode.BoolToBytes(o.TLS))

	// UserName
	username_b := []byte(o.UserName)
	username_b_len := len(username_b)
	buf.Write(iendecode.IntToBytes(username_b_len))
	buf.Write(username_b)

	// Password sha1 40
	password_b := []byte(o.Password)
	password_b_len := len(password_b)
	buf.Write(iendecode.IntToBytes(password_b_len))
	buf.Write(password_b)

	return buf.Bytes(), err
}

func (o *O_DRuleOperator) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Name
	name_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Name = string(buf.Next(name_b_len))

	// Address
	address_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Address = string(buf.Next(address_b_len))

	// ConnNum
	o.ConnNum = iendecode.BytesToInt(buf.Next(8))

	// TLS
	o.TLS = iendecode.BytesToBool(buf.Next(1))

	// Username
	username_b_len := iendecode.BytesToInt(buf.Next(8))
	o.UserName = string(buf.Next(username_b_len))

	// Password
	password_b_len := iendecode.BytesToInt(buf.Next(8))
	o.Password = string(buf.Next(password_b_len))

	return
}

// 蔓延到其他drule上的区域
type O_AreasRouter struct {
	AreaName string              // 区域名称
	Mirror   bool                // 是否为镜像，ture为镜像，则所有的文件都发给下面所有的drule
	Mirrors  []string            // string为drule的名字
	Chars    map[string][]string // 如果mirror为false，则看这个根据不同的字母进行路由，第一个stirng为首字母，第二个string为operator的名称
}
