// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package roles

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"

	"github.com/idcsource/insight00-lib/iendecode"
)

const (
	// Father be changed
	FATHER_CHANGED = iota
	// Children be changed
	CHILDREN_CHANGED
	// friends be changed
	FRIENDS_CHANGED
	// self be changed, all top three
	SELF_CHANGED
	// data be changed
	DATA_CHANGED
	// context be changed
	CONTEXT_CHANGED
)

type ContextUpDown uint8

const (
	// 上下文上游
	CONTEXT_UP ContextUpDown = iota
	// 上下文下游
	CONTEXT_DOWN
)

type StatusValueType uint8

const (
	// 状态位的值类型：null
	STATUS_VALUE_TYPE_NULL StatusValueType = iota
	// 状态位的值类型：int64
	STATUS_VALUE_TYPE_INT
	// 状态位的值类型：float64
	STATUS_VALUE_TYPE_FLOAT
	// 状态位的值类型：complex128
	STATUS_VALUE_TYPE_COMPLEX
	// 状态位的值类型：string
	STATUS_VALUE_TYPE_STRING
)

// Role为基本角色类型。
// 此类型实现了Roleer接口，
// 应被所有用户自定义角色类型所继承。
type Role struct {
	// 角色ID
	Id string
	// 角色的版本号
	_role_version uint32
	// 父角色（拓扑结构层面）
	_father string
	// 虚拟的子角色群，只保存键名
	_children []string
	// 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	_friends map[string]Status
	// 上下文关系列表
	_context map[string]Context
	// 父角色被更改
	_father_changed bool
	// 子角色关系被改变
	_children_changed bool
	// 朋友角色被改变
	_friends_changed bool
	// 上下文关系改变
	_context_changed bool
	// 数据被改变
	_data_changed bool
	// 是否被删除
	_be_delete bool
}

// 句柄上下文的结构
type Context struct {
	// 上游
	Up map[string]Status
	// 下游
	Down map[string]Status
}

func (c Context) EncodeBinary() (b []byte, lens int64, err error) {
	up_b, up_lens, err := c.mapByte(c.Up)
	if err != nil {
		return
	}
	down_b, down_lens, err := c.mapByte(c.Down)
	if err != nil {
		return
	}
	lens = up_lens + down_lens + 16
	b = make([]byte, lens)
	copy(b, iendecode.Uint64ToBytes(uint64(up_lens)))
	copy(b[8:], up_b)
	copy(b[up_lens+8:], iendecode.Uint64ToBytes(uint64(down_lens)))
	copy(b[up_lens+16:], down_b)
	return
}

func (c Context) mapByte(m map[string]Status) (b []byte, lens int64, err error) {
	b_buf := bytes.Buffer{}
	for key, _ := range m {
		key_len := iendecode.Uint64ToBytes(uint64(len(key)))
		b_buf.Write(key_len)
		b_buf.Write([]byte(key))
		var s_b []byte
		var s_lens int64
		s_b, s_lens, err = m[key].EncodeBinary()
		if err != nil {
			return
		}
		b_buf.Write(iendecode.Uint64ToBytes(uint64(s_lens)))
		b_buf.Write(s_b)
	}
	lens = int64(b_buf.Len())
	b = b_buf.Bytes()
	return
}

func (c *Context) DecodeBinary(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	c.Up = make(map[string]Status)
	c.Down = make(map[string]Status)

	up_len := iendecode.BytesToUint64(b[0:8])
	if err != nil {
		return err
	}
	c.Up, err = c.byteMap(b[8 : 8+up_len])
	if err != nil {
		return err
	}
	down_len := iendecode.BytesToUint64(b[8+up_len : 8+up_len+8])
	c.Down, err = c.byteMap(b[16+up_len : 16+up_len+down_len])
	if err != nil {
		return err
	}
	return
}

func (c Context) byteMap(b []byte) (m map[string]Status, err error) {
	m = make(map[string]Status)
	b_buf := bytes.NewBuffer(b)
	var i uint64 = 0
	b_len := uint64(len(b))
	for {
		if i >= b_len {
			break
		}
		key_len_b := b_buf.Next(8)
		key_len := iendecode.BytesToUint64(key_len_b)
		key := b_buf.Next(int(key_len))
		s_len_b := b_buf.Next(8)
		s_len := iendecode.BytesToUint64(s_len_b)
		s_b := b_buf.Next(int(s_len))
		s := Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
		err = s.DecodeBinary(s_b)
		if err != nil {
			return
		}
		m[string(key)] = s
		i += 16 + key_len + s_len
	}
	return
}

// 状态的数据结构
type Status struct {
	Int     []int64
	Float   []float64
	Complex []complex128
	String  []string
}

func (s Status) EncodeBinary() (b []byte, lens int64, err error) {
	var int_b []byte // 80
	int_b, err = iendecode.ToBinary(s.Int)
	if err != nil {
		return
	}
	var float_b []byte // 80
	float_b, err = iendecode.ToBinary(s.Float)
	if err != nil {
		return
	}
	var complex_b []byte // 160
	complex_b, err = iendecode.ToBinary(s.Complex)
	if err != nil {
		return
	}
	string_buf := bytes.Buffer{}
	var string_b_len int64 = 0
	for i := range s.String {
		sb := []byte(s.String[i])
		sb_l := int64(len(sb))
		if sb_l != 0 {
			string_b_len += 8 + sb_l
			string_buf.Write(iendecode.Uint64ToBytes(uint64(sb_l)))
			string_buf.Write(sb)
		} else {
			string_b_len += 8
			string_buf.Write(iendecode.Uint64ToBytes(uint64(sb_l)))
		}
	}
	string_b := string_buf.Bytes()
	lens = 320 + string_b_len
	b = make([]byte, lens)
	copy(b, int_b)
	copy(b[80:], float_b)
	copy(b[160:], complex_b)
	copy(b[320:], string_b)
	return
}

func (s *Status) DecodeBinary(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	s.Int = make([]int64, 10)
	s.Float = make([]float64, 10)
	s.Complex = make([]complex128, 10)
	s.String = make([]string, 10)

	err = iendecode.FromBinary(b[0:80], &s.Int)
	if err != nil {
		return
	}
	err = iendecode.FromBinary(b[80:160], &s.Float)
	if err != nil {
		return
	}
	err = iendecode.FromBinary(b[160:320], &s.Complex)
	if err != nil {
		return
	}
	j := 0
	var i uint64 = 320
	for {
		if i >= uint64(len(b)) {
			break
		}
		slen := iendecode.BytesToUint64(b[i : i+8])
		if slen != 0 {
			s.String[j] = string(b[i+8 : i+8+slen])
		}
		i += 8 + slen
		j++
	}
	return
}

// 新建自己
func (r *Role) New(id string) {
	r.Id = id
	r._role_version = 2
	r._children = make([]string, 0)
	r._friends = make(map[string]Status)
	r._context = make(map[string]Context)
	r._father_changed = false
	r._children_changed = false
	r._friends_changed = false
	r._context_changed = false
	r._data_changed = true
	r._be_delete = false
}

// 获取自己的版本
func (r *Role) Version() (version uint32) {
	return r._role_version
}

// 设置自己的版本（通常这个是在存储（如HardStore）来处理的时候才需要甬道）
func (r *Role) SetVersion(version uint32) {
	r._role_version = version
}

// 返回status代表的部分是否被改变，被改变则返回true
func (r *Role) ReturnChanged(status uint8) bool {
	switch status {
	case FATHER_CHANGED:
		return r._father_changed
	case CHILDREN_CHANGED:
		return r._children_changed
	case FRIENDS_CHANGED:
		return r._friends_changed
	case DATA_CHANGED:
		return r._data_changed
	case CONTEXT_CHANGED:
		return r._context_changed
	case SELF_CHANGED:
		if r._father_changed == true {
			return true
		} else if r._children_changed == true {
			return true
		} else if r._friends_changed == true {
			return true
		} else if r._context_changed == true {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

// 设置数据体被改变，这个方法应该由RolesControl来调用
func (r *Role) SetDataChanged() {
	r._data_changed = true
}

// 返回角色自身的ID
func (r *Role) ReturnId() string {
	return r.Id
}

// 返回自己的父亲是谁
func (r *Role) GetFather() string {
	return r._father
}

// 返回整个子角色关系
func (r *Role) GetChildren() []string {
	return r._children
}

// 返回整个朋友关系
func (r *Role) GetFriends() map[string]Status {
	return r._friends
}

// 重置父关系，也就是将父关系清空
func (r *Role) ResetFather() {
	r._father = ""
	r._father_changed = true
}

// 重置子关系，也就是将子关系清空
func (r *Role) ResetChilren() {
	r._children = make([]string, 0)
	r._children_changed = true
}

// 重置朋友关系，也就是将朋友关系清空
func (r *Role) ResetFriends() {
	r._friends = make(map[string]Status)
	r._friends_changed = true
}

// 设置父关系
func (r *Role) SetFather(id string) {
	r._father = id
	r._father_changed = true
}

// 设置整个子关系
func (r *Role) SetChildren(children []string) {
	r._children = children
	r._children_changed = true
}

// 设置整个朋友关系
func (r *Role) SetFriends(friends map[string]Status) {
	r._friends = friends
	r._friends_changed = true
}

// 看是否存在某个子角色，如果存在返回true
func (r *Role) ExistChild(name string) bool {
	for _, v := range r._children {
		if v == name {
			return true
			break
		}
	}
	return false
}

// 是否存在某个朋友，并且返回这个朋友的疏离关系
func (r *Role) ExistFriend(name string) (bool, int64) {
	v, FindV := r._friends[name]
	if FindV == true {
		return true, v.Int[0]
	}
	return false, 0
}

// 将一个子角色添加进去
func (r *Role) AddChild(cid string) error {
	exist := r.ExistChild(cid)
	if exist == true {
		err := errors.New("This Role has already exist : " + cid + " in " + r.Id + " _children .")
		return err
	} else {
		r._children = append(r._children, cid)
		r._children_changed = true
		return nil
	}
}

// 删除一个子角色
func (r *Role) DeleteChild(child string) error {
	exist := r.ExistChild(child)
	if exist != true {
		err := errors.New("Role has no exist : " + child + " in " + r.Id + " _children .")
		return err
	} else {
		var count int
		for i, v := range r._children {
			if v == child {
				count = i
				break
			}
		}
		r._children = append(r._children[:count], r._children[count+1:]...)
		r._children_changed = true
		return nil
	}
}

// 添加一个朋友关系
func (r *Role) AddFriend(id string, bind int64) error {
	// 检查这个friend是否存在
	ifexist, _ := r.ExistFriend(id)
	if ifexist == true {
		err := errors.New("This Role has already exist : " + id + " in " + r.Id + " friend .")
		return err
	}
	r._friends[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
	r._friends[id].Int[0] = bind
	r._friends_changed = true
	return nil
}

// 删除一个朋友关系
func (r *Role) DeleteFriend(id string) error {
	ifexist, _ := r.ExistFriend(id)
	if ifexist == true {
		//err := Error(1, id, "in", r.Id, "friend .");
		err := errors.New("Role has no exist : " + id + " in " + r.Id + " friend .")
		return err
	}
	delete(r._friends, id)
	r._friends_changed = true
	return nil
}

// 修改一个朋友关系，只是改关系远近
func (r *Role) ChangeFriend(id string, bind int64) error {
	ifexist, obind := r.ExistFriend(id)
	if ifexist == true {
		err := errors.New("Role has no exist : " + id + " in " + r.Id + " friend .")
		return err
	}
	if obind == bind {
		return nil
	}
	r._friends[id].Int[0] = bind
	r._friends_changed = true
	return nil
}

// 获取相同远近关系下的所有朋友的ID
func (r *Role) GetSameBindFriendsId(bind int64) []string {
	rstr := make([]string, 0)
	for id, binds := range r._friends {
		if binds.Int[0] == bind {
			rstr = append(rstr, id)
		}
	}
	return rstr
}

// 查看删除状态
func (r *Role) ReturnDelete() bool {
	return r._be_delete
}

// 设置删除状态，true则为删除
func (r *Role) SetDelete(del bool) {
	r._be_delete = del
}

// 设定全部上下文，存储实例调用
func (r *Role) SetContexts(context map[string]Context) {
	r._context = context
}

// 获取全部上下文，存储实例调用
func (r *Role) GetContexts() map[string]Context {
	return r._context
}

// 创建一个空的上下文
func (r *Role) NewContext(contextname string) (err error) {
	_, find := r._context[contextname]
	if find == false {
		r._context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	} else {
		err = fmt.Errorf("The context already exist.")
	}
	return
}

// 是否存在一个上下文
func (r *Role) ExistContext(contextname string) (have bool) {
	_, have = r._context[contextname]
	return have
}

// 设定一个上下文的上游
func (r *Role) AddContextUp(contextname, upname string, bind int64) {
	_, find := r._context[contextname]
	if find == false {
		r._context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		r._context[contextname].Up[upname] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
	}
	r._context[contextname].Up[upname].Int[0] = bind
	r._context_changed = true
}

// 设定一个上下文的下游
func (r *Role) AddContextDown(contextname, downname string, bind int64) {
	_, find := r._context[contextname]
	if find == false {
		r._context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		r._context[contextname].Down[downname] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
	}
	r._context[contextname].Down[downname].Int[0] = bind
	r._context_changed = true
}

// 删除一个上下文的上游
func (r *Role) DelContextUp(contextname, upname string) {
	_, find := r._context[contextname]
	if find == true {
		if _, find2 := r._context[contextname].Up[upname]; find2 == true {
			delete(r._context[contextname].Up, upname)
		}
	}
	r._context_changed = true
}

// 删除一个上下文的下游
func (r *Role) DelContextDown(contextname, downname string) {
	_, find := r._context[contextname]
	if find == true {
		if _, find2 := r._context[contextname].Down[downname]; find2 == true {
			delete(r._context[contextname].Down, downname)
		}
	}
	r._context_changed = true
}

// 清除一个上下文
func (r *Role) DelContext(contextname string) {
	_, find := r._context[contextname]
	if find == true {
		delete(r._context, contextname)
	}
}

// 找到一个上下文的上文，返回绑定值
func (r *Role) GetContextUp(contextname, upname string) (bind int64, have bool) {
	if _, have = r._context[contextname]; have == true {
		var binds Status
		if binds, have = r._context[contextname].Up[upname]; have == true {
			bind = binds.Int[0]
			return
		} else {
			have = false
			return
		}
	} else {
		have = false
		return
	}
}

// 找到一个上下文的下文，返回绑定值
func (r *Role) GetContextDown(contextname, downname string) (bind int64, have bool) {
	if _, have = r._context[contextname]; have == true {
		var binds Status
		if binds, have = r._context[contextname].Down[downname]; have == true {
			bind = binds.Int[0]
			return
		} else {
			have = false
			return
		}
	} else {
		have = false
		return
	}
}

// 返回某个上下文的全部信息
func (r *Role) GetContext(contextname string) (context Context, have bool) {
	if context, have = r._context[contextname]; have == true {
		return
	} else {
		have = false
		return
	}
}

// 设置某个上下文的全部信息
func (r *Role) SetContext(contextname string, context Context) {
	r._context[contextname] = context
	r._context_changed = true
}

// 返回某个上下文中的上游同样绑定值的所有
func (r *Role) GetContextUpSameBind(contextname string, bind int64) (rolesid []string, have bool) {
	if _, find := r._context[contextname]; find == true {
		rolesid = make([]string, 0)
		for id, binds := range r._context[contextname].Up {
			if binds.Int[0] == bind {
				rolesid = append(rolesid, id)
			}
		}
		have = true
		return
	} else {
		have = false
		return
	}
}

// 返回某个上下文中的下游同样绑定值的所有
func (r *Role) GetContextDownSameBind(contextname string, bind int64) (rolesid []string, have bool) {
	if _, find := r._context[contextname]; find == true {
		rolesid = make([]string, 0)
		for id, binds := range r._context[contextname].Down {
			if binds.Int[0] == bind {
				rolesid = append(rolesid, id)
			}
		}
		have = true
		return
	} else {
		have = false
		return
	}
}

// 返回所有上下文组的名称
func (r *Role) GetContextsName() (names []string) {
	lens := len(r._context)
	names = make([]string, lens)
	i := 0
	for name, _ := range r._context {
		names[i] = name
		i++
	}
	return names
}

// 设置朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) SetFriendStatus(id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles: SetFriendStatus: %v", e)
		}
	}()
	_, findf := r._friends[id]
	if findf == false {
		r._friends[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
	}
	if bit > 9 {
		return errors.New("The bit must less than 10.")
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int":
		r._friends[id].Int[bit] = valuer.Int()
	case "int64":
		r._friends[id].Int[bit] = valuer.Int()
	case "float":
		r._friends[id].Float[bit] = valuer.Float()
	case "float64":
		r._friends[id].Float[bit] = valuer.Float()
	case "complex64":
		r._friends[id].Complex[bit] = valuer.Complex()
	case "complex128":
		r._friends[id].Complex[bit] = valuer.Complex()
	case "string":
		r._friends[id].String[bit] = valuer.String()
	default:
		return errors.New("The value's type must int64, float64, complex128 or string.")
	}
	r._friends_changed = true
	return nil
}

// 获取朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) GetFriendStatus(id string, bit int, value interface{}) (have bool, err error) {
	have = true
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles: GetFriendStatus: %v", e)
		}
	}()
	_, findf := r._friends[id]
	if findf == false {
		have = false
		return
	}
	if bit > 9 {
		err = errors.New("The bit must less than 10.")
		return
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		valuer.SetInt(r._friends[id].Int[bit])
	case "float64":
		valuer.SetFloat(r._friends[id].Float[bit])
	case "complex128":
		valuer.SetComplex(r._friends[id].Complex[bit])
	case "string":
		valuer.SetString(r._friends[id].String[bit])
	default:
		err = errors.New("The value's type must int64, float64, complex128 or string.")
	}
	return
}

// 设置上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) SetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles: SetContextStatus: %v", e)
		}
	}()
	if bit > 9 {
		return errors.New("The bit must less than 10.")
	}
	_, findc := r._context[contextname]
	if findc == false {
		r._context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	}
	if upordown == CONTEXT_UP {
		_, findr := r._context[contextname].Up[id]
		if findr == false {
			r._context[contextname].Up[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int":
			r._context[contextname].Up[id].Int[bit] = valuer.Int()
		case "int64":
			r._context[contextname].Up[id].Int[bit] = valuer.Int()
		case "float":
			r._context[contextname].Up[id].Float[bit] = valuer.Float()
		case "float64":
			r._context[contextname].Up[id].Float[bit] = valuer.Float()
		case "complex64":
			r._context[contextname].Up[id].Complex[bit] = valuer.Complex()
		case "complex128":
			r._context[contextname].Up[id].Complex[bit] = valuer.Complex()
		case "string":
			r._context[contextname].Up[id].String[bit] = valuer.String()
		default:
			return errors.New("The value's type must int64, float64, complex128 or string.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r._context[contextname].Down[id]
		if findr == false {
			r._context[contextname].Down[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int":
			r._context[contextname].Down[id].Int[bit] = valuer.Int()
		case "int64":
			r._context[contextname].Down[id].Int[bit] = valuer.Int()
		case "float":
			r._context[contextname].Down[id].Float[bit] = valuer.Float()
		case "float64":
			r._context[contextname].Down[id].Float[bit] = valuer.Float()
		case "complex64":
			r._context[contextname].Down[id].Complex[bit] = valuer.Complex()
		case "complex128":
			r._context[contextname].Down[id].Complex[bit] = valuer.Complex()
		case "string":
			r._context[contextname].Down[id].String[bit] = valuer.String()
		default:
			return errors.New("The value's type must int64, float64, complex128 or string.")
		}
	} else {
		return errors.New("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return nil
}

// 获取上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) GetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (have bool, err error) {
	have = true
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles: GetContextStatus: %v", e)
		}
	}()
	if bit > 9 {
		err = errors.New("The bit must less than 10.")
		return
	}
	_, findc := r._context[contextname]
	if findc == false {
		err = errors.New("The Role have no context " + contextname + " in " + r.Id + " .")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := r._context[contextname].Up[id]
		if findr == false {
			//return errors.New("The Role have no up context relationship " + id + " in " + contextname + " in " + r.Id + " .")
			have = false
			return
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int64":
			valuer.SetInt(r._context[contextname].Up[id].Int[bit])
		case "float64":
			valuer.SetFloat(r._context[contextname].Up[id].Float[bit])
		case "complex128":
			valuer.SetComplex(r._context[contextname].Up[id].Complex[bit])
		case "string":
			valuer.SetString(r._context[contextname].Up[id].String[bit])
		default:
			err = errors.New("The value's type must int64, float64, complex128 or string.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r._context[contextname].Down[id]
		if findr == false {
			//return errors.New("The Role have no down context relationship " + id + " in " + contextname + " in " + r.Id + " .")
			have = false
			return
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int64":
			valuer.SetInt(r._context[contextname].Down[id].Int[bit])
		case "float64":
			valuer.SetFloat(r._context[contextname].Down[id].Float[bit])
		case "complex128":
			valuer.SetComplex(r._context[contextname].Down[id].Complex[bit])
		case "string":
			valuer.SetString(r._context[contextname].Down[id].String[bit])
		default:
			err = errors.New("The value's type must int64, float64, complex128 or string.")
		}
	} else {
		err = errors.New("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

// 为Gob注册角色类型
func RegInterfaceForGob() {
	gob.Register(&Role{})
}
