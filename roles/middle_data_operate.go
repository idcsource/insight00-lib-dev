// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 角色（Role）概念封装的数据存储与数据关系
package roles

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"
)

// 获取自己的版本
func (r *RoleMiddleData) ReturnVersion() (version int) {
	return r.Version.Version
}

// 设置自己的版本（通常这个是在存储（如HardStore）来处理的时候才需要甬道）
func (r *RoleMiddleData) SetVersion(version int) {
	r.Version.Version = version
}

// 返回角色自身的ID
func (r *RoleMiddleData) ReturnId() string {
	return r.Version.Id
}

func (r *RoleMiddleData) SetId(id string) {
	r.Version.Id = id
}

// 返回自己的父亲是谁
func (r *RoleMiddleData) GetFather() string {
	return r.Relation.Father
}

// 返回整个子角色关系
func (r *RoleMiddleData) GetChildren() []string {
	return r.Relation.Children
}

// 返回整个朋友关系
func (r *RoleMiddleData) GetFriends() map[string]Status {
	return r.Relation.Friends
}

// 重置父关系，也就是将父关系清空
func (r *RoleMiddleData) ResetFather() {
	r.Relation.Father = ""
}

// 重置子关系，也就是将子关系清空
func (r *RoleMiddleData) ResetChilren() {
	r.Relation.Children = make([]string, 0)
}

// 重置朋友关系，也就是将朋友关系清空
func (r *RoleMiddleData) ResetFriends() {
	r.Relation.Friends = make(map[string]Status)
}

// 设置父关系
func (r *RoleMiddleData) SetFather(id string) {
	r.Relation.Father = id
}

// 设置整个子关系
func (r *RoleMiddleData) SetChildren(children []string) {
	r.Relation.Children = children
}

// 设置整个朋友关系
func (r *RoleMiddleData) SetFriends(friends map[string]Status) {
	r.Relation.Friends = friends
}

// 看是否存在某个子角色，如果存在返回true
func (r *RoleMiddleData) ExistChild(name string) bool {
	for _, v := range r.Relation.Children {
		if v == name {
			return true
			break
		}
	}
	return false
}

// 是否存在某个朋友，并且返回这个朋友的疏离关系
func (r *RoleMiddleData) ExistFriend(name string) (bool, int64) {
	v, FindV := r.Relation.Friends[name]
	if FindV == true {
		return true, v.Int[0]
	}
	return false, 0
}

// 将一个子角色添加进去
func (r *RoleMiddleData) AddChild(cid string) error {
	exist := r.ExistChild(cid)
	if exist == true {
		err := fmt.Errorf("roles[RoleMiddleData]AddChild: This Role has already exist : " + cid + " in " + r.Version.Id + ".")
		return err
	} else {
		r.Relation.Children = append(r.Relation.Children, cid)
		return nil
	}
}

// 删除一个子角色
func (r *RoleMiddleData) DeleteChild(child string) error {
	exist := r.ExistChild(child)
	if exist != true {
		err := errors.New("roles[RoleMiddleData]DeleteChild: Role has no exist : " + child + " in " + r.Version.Id + " .")
		return err
	} else {
		var count int
		for i, v := range r.Relation.Children {
			if v == child {
				count = i
				break
			}
		}
		r.Relation.Children = append(r.Relation.Children[:count], r.Relation.Children[count+1:]...)
		return nil
	}
}

// 添加一个朋友关系
func (r *RoleMiddleData) AddFriend(id string, bind int64) error {
	// 检查这个friend是否存在
	ifexist, _ := r.ExistFriend(id)
	if ifexist == true {
		err := errors.New("roles[RoleMiddleData]AddFriend: This Role has already exist : " + id + " in " + r.Version.Id + " friend .")
		return err
	}
	r.Relation.Friends[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
	r.Relation.Friends[id].Int[0] = bind
	return nil
}

// 删除一个朋友关系
func (r *RoleMiddleData) DeleteFriend(id string) error {
	ifexist, _ := r.ExistFriend(id)
	if ifexist == true {
		//err := Error(1, id, "in", r.Id, "friend .");
		err := errors.New("roles[RoleMiddleData]DeleteFriend: Role has no exist : " + id + " in " + r.Version.Id + " friend .")
		return err
	}
	delete(r.Relation.Friends, id)
	return nil
}

// 修改一个朋友关系，只是改关系远近
func (r *RoleMiddleData) ChangeFriend(id string, bind int64) error {
	ifexist, obind := r.ExistFriend(id)
	if ifexist == true {
		err := errors.New("roles[RoleMiddleData]ChangeFriend: Role has no exist : " + id + " in " + r.Version.Id + " friend .")
		return err
	}
	if obind == bind {
		return nil
	}
	r.Relation.Friends[id].Int[0] = bind
	return nil
}

// 获取相同远近关系下的所有朋友的ID
func (r *RoleMiddleData) GetSameBindFriendsId(bind int64) []string {
	rstr := make([]string, 0)
	for id, binds := range r.Relation.Friends {
		if binds.Int[0] == bind {
			rstr = append(rstr, id)
		}
	}
	return rstr
}

// 设定全部上下文，存储实例调用
func (r *RoleMiddleData) SetContexts(context map[string]Context) {
	r.Relation.Contexts = context
}

// 获取全部上下文，存储实例调用
func (r *RoleMiddleData) GetContexts() map[string]Context {
	return r.Relation.Contexts
}

// 创建一个空的上下文，如果已经存在则忽略
func (r *RoleMiddleData) NewContext(contextname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == false {
		r.Relation.Contexts[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	}
}

// 设定一个上下文的上游
func (r *RoleMiddleData) AddContextUp(contextname, upname string, bind int64) {
	_, find := r.Relation.Contexts[contextname]
	if find == false {
		r.Relation.Contexts[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		r.Relation.Contexts[contextname].Up[upname] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
	}
	r.Relation.Contexts[contextname].Up[upname].Int[0] = bind
}

// 设定一个上下文的下游
func (r *RoleMiddleData) AddContextDown(contextname, downname string, bind int64) {
	_, find := r.Relation.Contexts[contextname]
	if find == false {
		r.Relation.Contexts[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		r.Relation.Contexts[contextname].Down[downname] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
	}
	r.Relation.Contexts[contextname].Down[downname].Int[0] = bind
}

// 删除一个上下文的上游
func (r *RoleMiddleData) DelContextUp(contextname, upname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == true {
		if _, find2 := r.Relation.Contexts[contextname].Up[upname]; find2 == true {
			delete(r.Relation.Contexts[contextname].Up, upname)
		}
	}
}

// 删除一个上下文的下游
func (r *RoleMiddleData) DelContextDown(contextname, downname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == true {
		if _, find2 := r.Relation.Contexts[contextname].Down[downname]; find2 == true {
			delete(r.Relation.Contexts[contextname].Down, downname)
		}
	}
}

// 清除一个上下文
func (r *RoleMiddleData) DelContext(contextname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == true {
		delete(r.Relation.Contexts, contextname)
	}
}

// 找到一个上下文的上文，返回绑定值
func (r *RoleMiddleData) GetContextUp(contextname, upname string) (bind int64, have bool) {
	if _, have = r.Relation.Contexts[contextname]; have == true {
		var binds Status
		if binds, have = r.Relation.Contexts[contextname].Up[upname]; have == true {
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
func (r *RoleMiddleData) GetContextDown(contextname, downname string) (bind int64, have bool) {
	if _, have = r.Relation.Contexts[contextname]; have == true {
		var binds Status
		if binds, have = r.Relation.Contexts[contextname].Down[downname]; have == true {
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
func (r *RoleMiddleData) GetContext(contextname string) (context Context, have bool) {
	if context, have = r.Relation.Contexts[contextname]; have == true {
		return
	} else {
		have = false
		return
	}
}

// 返回某个上下文中的上游同样绑定值的所有
func (r *RoleMiddleData) GetContextUpSameBind(contextname string, bind int64) (rolesid []string, have bool) {
	if _, find := r.Relation.Contexts[contextname]; find == true {
		rolesid = make([]string, 0)
		for id, binds := range r.Relation.Contexts[contextname].Up {
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
func (r *RoleMiddleData) GetContextDownSameBind(contextname string, bind int64) (rolesid []string, have bool) {
	if _, find := r.Relation.Contexts[contextname]; find == true {
		rolesid = make([]string, 0)
		for id, binds := range r.Relation.Contexts[contextname].Down {
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
func (r *RoleMiddleData) GetContextsName() (names []string) {
	lens := len(r.Relation.Contexts)
	names = make([]string, lens)
	i := 0
	for name, _ := range r.Relation.Contexts {
		names[i] = name
		i++
	}
	return names
}

// 设置朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) SetFriendStatus(id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]SetFriendStatus: %v", e)
		}
	}()
	_, findf := r.Relation.Friends[id]
	if findf == false {
		r.Relation.Friends[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
	}
	if bit > 9 {
		return errors.New("The bit must less than 10.")
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int":
		r.Relation.Friends[id].Int[bit] = valuer.Int()
	case "int64":
		r.Relation.Friends[id].Int[bit] = valuer.Int()
	case "float":
		r.Relation.Friends[id].Float[bit] = valuer.Float()
	case "float64":
		r.Relation.Friends[id].Float[bit] = valuer.Float()
	case "complex64":
		r.Relation.Friends[id].Complex[bit] = valuer.Complex()
	case "complex128":
		r.Relation.Friends[id].Complex[bit] = valuer.Complex()
	default:
		return fmt.Errorf("roles[RoleMiddleData]SetFriendStatus: The value's type must int64, float64 or complex128.")
	}
	return nil
}

// 获取朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) GetFriendStatus(id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: %v", e)
		}
	}()
	_, findf := r.Relation.Friends[id]
	if findf == false {
		return fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: Role has no exist : " + id + " in " + r.Version.Id + " friend .")
	}
	if bit > 9 {
		return fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: The bit must less than 10.")
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		valuer.SetInt(r.Relation.Friends[id].Int[bit])
	case "float64":
		valuer.SetFloat(r.Relation.Friends[id].Float[bit])
	case "complex128":
		valuer.SetComplex(r.Relation.Friends[id].Complex[bit])
	default:
		return fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: The value's type must int64, float64 or complex128.")
	}
	return nil
}

// 设置上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) SetContextStatus(contextname string, upordown uint8, id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles:[RoleMiddleData]SetContextStatus: %v", e)
		}
	}()
	if bit > 9 {
		return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus: The bit must less than 10.")
	}
	_, findc := r.Relation.Contexts[contextname]
	if findc == false {
		r.Relation.Contexts[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	}
	if upordown == CONTEXT_UP {
		_, findr := r.Relation.Contexts[contextname].Up[id]
		if findr == false {
			r.Relation.Contexts[contextname].Up[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int":
			r.Relation.Contexts[contextname].Up[id].Int[bit] = valuer.Int()
		case "int64":
			r.Relation.Contexts[contextname].Up[id].Int[bit] = valuer.Int()
		case "float":
			r.Relation.Contexts[contextname].Up[id].Float[bit] = valuer.Float()
		case "float64":
			r.Relation.Contexts[contextname].Up[id].Float[bit] = valuer.Float()
		case "complex64":
			r.Relation.Contexts[contextname].Up[id].Complex[bit] = valuer.Complex()
		case "complex128":
			r.Relation.Contexts[contextname].Up[id].Complex[bit] = valuer.Complex()
		default:
			return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The value's type must int64, float64 or complex128.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r.Relation.Contexts[contextname].Down[id]
		if findr == false {
			r.Relation.Contexts[contextname].Down[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10)}
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int":
			r.Relation.Contexts[contextname].Down[id].Int[bit] = valuer.Int()
		case "int64":
			r.Relation.Contexts[contextname].Down[id].Int[bit] = valuer.Int()
		case "float":
			r.Relation.Contexts[contextname].Down[id].Float[bit] = valuer.Float()
		case "float64":
			r.Relation.Contexts[contextname].Down[id].Float[bit] = valuer.Float()
		case "complex64":
			r.Relation.Contexts[contextname].Down[id].Complex[bit] = valuer.Complex()
		case "complex128":
			r.Relation.Contexts[contextname].Down[id].Complex[bit] = valuer.Complex()
		default:
			return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The value's type must int64, float64 or complex128.")
		}
	} else {
		return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return nil
}

// 获取上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) GetContextStatus(contextname string, upordown uint8, id string, bit int, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]GetContextStatus: %v", e)
		}
	}()
	if bit > 9 {
		return errors.New("roles[RoleMiddleData]GetContextStatus: The bit must less than 10.")
	}
	_, findc := r.Relation.Contexts[contextname]
	if findc == false {
		return errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no context " + contextname + " in " + r.Version.Id + " .")
	}
	if upordown == CONTEXT_UP {
		_, findr := r.Relation.Contexts[contextname].Up[id]
		if findr == false {
			return errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no up context relationship " + id + " in " + contextname + " in " + r.Version.Id + " .")
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int64":
			valuer.SetInt(r.Relation.Contexts[contextname].Up[id].Int[bit])
		case "float64":
			valuer.SetFloat(r.Relation.Contexts[contextname].Up[id].Float[bit])
		case "complex128":
			valuer.SetComplex(r.Relation.Contexts[contextname].Up[id].Complex[bit])
		default:
			return errors.New("roles[RoleMiddleData]GetContextStatus: The value's type must int64, float64 or complex128.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r.Relation.Contexts[contextname].Down[id]
		if findr == false {
			return errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no down context relationship " + id + " in " + contextname + " in " + r.Version.Id + " .")
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String()
		switch vname {
		case "int64":
			valuer.SetInt(r.Relation.Contexts[contextname].Down[id].Int[bit])
		case "float64":
			valuer.SetFloat(r.Relation.Contexts[contextname].Down[id].Float[bit])
		case "complex128":
			valuer.SetComplex(r.Relation.Contexts[contextname].Down[id].Complex[bit])
		default:
			return errors.New("roles[RoleMiddleData]GetContextStatus: The value's type must int64, float64 or complex128.")
		}
	} else {
		return errors.New("roles[RoleMiddleData]GetContextStatus: The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return nil
}

func (r *RoleMiddleData) GetDataToInterface(name, typename string) (data interface{}, err error) {
	var find bool
	switch typename {
	case "time.Time":
		data, find = r.Normal.Time[name]
	case "[]byte":
		data, find = r.Normal.Byte[name]
	case "string":
		data, find = r.Normal.String[name]
	case "bool":
		data, find = r.Normal.Bool[name]
	case "uint8":
		data, find = r.Normal.Uint8[name]
	case "uint":
		data, find = r.Normal.Uint[name]
	case "uint64":
		data, find = r.Normal.Uint64[name]
	case "int8":
		data, find = r.Normal.Int8[name]
	case "int":
		data, find = r.Normal.Int[name]
	case "int64":
		data, find = r.Normal.Int64[name]
	case "float32":
		data, find = r.Normal.Float32[name]
	case "float64":
		data, find = r.Normal.Float64[name]
	case "complex64":
		data, find = r.Normal.Complex64[name]
	case "complex128":
		data, find = r.Normal.Complex128[name]

	case "[]string":
		data, find = r.Slice.String[name]
	case "[]bool":
		data, find = r.Slice.Bool[name]
	case "[]uint8":
		data, find = r.Slice.Uint8[name]
	case "[]uint":
		data, find = r.Slice.Uint[name]
	case "[]uint64":
		data, find = r.Slice.Uint64[name]
	case "[]int8":
		data, find = r.Slice.Int8[name]
	case "[]int":
		data, find = r.Slice.Int[name]
	case "[]int64":
		data, find = r.Slice.Int64[name]
	case "[]float32":
		data, find = r.Slice.Float32[name]
	case "[]float64":
		data, find = r.Slice.Float64[name]
	case "[]complex64":
		data, find = r.Slice.Complex64[name]
	case "[]complex128":
		data, find = r.Slice.Complex128[name]

	case "map[string]string":
		data, find = r.StringMap.String[name]
	case "map[string]bool":
		data, find = r.StringMap.Bool[name]
	case "map[string]uint8":
		data, find = r.StringMap.Uint8[name]
	case "map[string]uint":
		data, find = r.StringMap.Uint[name]
	case "map[string]uint64":
		data, find = r.StringMap.Uint64[name]
	case "map[string]int8":
		data, find = r.StringMap.Int8[name]
	case "map[string]int":
		data, find = r.StringMap.Int[name]
	case "map[string]int64":
		data, find = r.StringMap.Int64[name]
	case "map[string]float32":
		data, find = r.StringMap.Float32[name]
	case "map[string]float64":
		data, find = r.StringMap.Float64[name]
	case "map[string]complex64":
		data, find = r.StringMap.Complex64[name]
	case "map[string]complex128":
		data, find = r.StringMap.Complex128[name]

	default:
		err = fmt.Errorf("roles[RoleMiddleData]GetDataToInterface: Can't find the data.")
		return
	}
	if find == false {
		err = fmt.Errorf("roles[RoleMiddleData]GetDataToInterface: Can't find the data.")
		return
	}
	return
}

// 中间类型的获取数据
func (r *RoleMiddleData) GetData(name string, datas interface{}) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddle]GetData: %v", e)
		}
	}()

	// 获取data的数据类型
	data_v := reflect.Indirect(reflect.ValueOf(datas))
	data_t := data_v.Type()
	data_type_name := data_t.String()
	var find bool
	var value interface{}
	// 查看找哪个文件
	if m, _ := regexp.MatchString(`^map\[string\]`, data_type_name); m == true {
		data := r.StringMap
		if err != nil {
			err = fmt.Errorf("roles[RoleMiddle]GetData: %v", err)
			return err
		}
		switch data_type_name {
		case "map[string]string":
			value, find = data.String[name]
		case "map[string]bool":
			value, find = data.Bool[name]
		case "map[string]uint8":
			value, find = data.Uint8[name]
		case "map[string]uint":
			value, find = data.Uint[name]
		case "map[string]uint64":
			value, find = data.Uint64[name]
		case "map[string]int8":
			value, find = data.Int8[name]
		case "map[string]int":
			value, find = data.Int[name]
		case "map[string]int64":
			value, find = data.Int64[name]
		case "map[string]float32":
			value, find = data.Float32[name]
		case "map[string]float64":
			value, find = data.Float64[name]
		case "map[string]complex64":
			value, find = data.Complex64[name]
		case "map[string]complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("roles[RoleMiddle]GetData: Unsupported data type %v", data_type_name)
			return err
		}
	} else if m, _ := regexp.MatchString(`^\[\]`, data_type_name); m == true {
		data := r.Slice
		switch data_type_name {
		case "[]string":
			value, find = data.String[name]
		case "[]bool":
			value, find = data.Bool[name]
		case "[]uint8":
			value, find = data.Uint8[name]
		case "[]uint":
			value, find = data.Uint[name]
		case "[]uint64":
			value, find = data.Uint64[name]
		case "[]int8":
			value, find = data.Int8[name]
		case "[]int":
			value, find = data.Int[name]
		case "[]int64":
			value, find = data.Int64[name]
		case "[]float32":
			value, find = data.Float32[name]
		case "[]float64":
			value, find = data.Float64[name]
		case "[]complex64":
			value, find = data.Complex64[name]
		case "[]complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("roles[RoleMiddle]GetData: Unsupported data type %v", data_type_name)
			return err
		}
	} else {
		data := r.Normal
		switch data_type_name {
		case "time.Time":
			value, find = data.Time[name]
		case "[]byte":
			value, find = data.Byte[name]
		case "string":
			value, find = data.String[name]
		case "bool":
			value, find = data.Bool[name]
		case "uint8":
			value, find = data.Uint8[name]
		case "uint":
			value, find = data.Uint[name]
		case "uint64":
			value, find = data.Uint64[name]
		case "int8":
			value, find = data.Int8[name]
		case "int":
			value, find = data.Int[name]
		case "int64":
			value, find = data.Int64[name]
		case "float32":
			value, find = data.Float32[name]
		case "float64":
			value, find = data.Float64[name]
		case "complex64":
			value, find = data.Complex64[name]
		case "complex128":
			value, find = data.Complex128[name]
		default:
			err = fmt.Errorf("roles[RoleMiddle]GetData: Unsupported data type %v", data_type_name)
			return err
		}
	}
	if find == false {
		err = fmt.Errorf("roles[RoleMiddleData]GetData: Can not find the field %v", name)
		return
	}

	value_v := reflect.ValueOf(value)
	data_v.Set(value_v)
	return
}

// 往中间类型设置数据
func (r *RoleMiddleData) SetData(name string, datas interface{}) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]SetData: %v", e)
		}
	}()

	// 获取data的数据类型
	data_v := reflect.ValueOf(datas)
	data_t := reflect.TypeOf(datas)
	data_type_name := data_t.String()
	// 查看找哪个文件
	if m, _ := regexp.MatchString(`^map\[string\]`, data_type_name); m == true {
		switch data_type_name {
		case "map[string]string":
			r.StringMap.String[name] = data_v.Interface().(map[string]string)
		case "map[string]bool":
			r.StringMap.Bool[name] = data_v.Interface().(map[string]bool)
		case "map[string]uint8":
			r.StringMap.Uint8[name] = data_v.Interface().(map[string]uint8)
		case "map[string]uint":
			r.StringMap.Uint[name] = data_v.Interface().(map[string]uint)
		case "map[string]uint64":
			r.StringMap.Uint64[name] = data_v.Interface().(map[string]uint64)
		case "map[string]int8":
			r.StringMap.Int8[name] = data_v.Interface().(map[string]int8)
		case "map[string]int":
			r.StringMap.Int[name] = data_v.Interface().(map[string]int)
		case "map[string]int64":
			r.StringMap.Int64[name] = data_v.Interface().(map[string]int64)
		case "map[string]float32":
			r.StringMap.Float32[name] = data_v.Interface().(map[string]float32)
		case "map[string]float64":
			r.StringMap.Float64[name] = data_v.Interface().(map[string]float64)
		case "map[string]complex64":
			r.StringMap.Complex64[name] = data_v.Interface().(map[string]complex64)
		case "map[string]complex128":
			r.StringMap.Complex128[name] = data_v.Interface().(map[string]complex128)
		default:
			err = fmt.Errorf("roles[RoleMiddleData]SetData: Unsupported data type %v", data_type_name)
			return
		}
	} else if m, _ := regexp.MatchString(`^\[\]`, data_type_name); m == true {
		switch data_type_name {
		case "[]string":
			r.Slice.String[name] = data_v.Interface().([]string)
		case "[]bool":
			r.Slice.Bool[name] = data_v.Interface().([]bool)
		case "[]uint8":
			r.Slice.Uint8[name] = data_v.Interface().([]uint8)
		case "[]uint":
			r.Slice.Uint[name] = data_v.Interface().([]uint)
		case "[]uint64":
			r.Slice.Uint64[name] = data_v.Interface().([]uint64)
		case "[]int8":
			r.Slice.Int8[name] = data_v.Interface().([]int8)
		case "[]int":
			r.Slice.Int[name] = data_v.Interface().([]int)
		case "[]int64":
			r.Slice.Int64[name] = data_v.Interface().([]int64)
		case "[]float32":
			r.Slice.Float32[name] = data_v.Interface().([]float32)
		case "[]float64":
			r.Slice.Float64[name] = data_v.Interface().([]float64)
		case "[]complex64":
			r.Slice.Complex64[name] = data_v.Interface().([]complex64)
		case "[]complex128":
			r.Slice.Complex128[name] = data_v.Interface().([]complex128)
		default:
			err = fmt.Errorf("roles[RoleMiddleData]SetData: Unsupported data type %v", data_type_name)
			return
		}
	} else {
		switch data_type_name {
		case "time.Time":
			r.Normal.Time[name] = data_v.Interface().(time.Time)
		case "[]byte":
			r.Normal.Byte[name] = data_v.Interface().([]byte)
		case "string":
			r.Normal.String[name] = data_v.Interface().(string)
		case "bool":
			r.Normal.Bool[name] = data_v.Interface().(bool)
		case "uint8":
			r.Normal.Uint8[name] = data_v.Interface().(uint8)
		case "uint":
			r.Normal.Uint[name] = data_v.Interface().(uint)
		case "uint64":
			r.Normal.Uint64[name] = data_v.Interface().(uint64)
		case "int8":
			r.Normal.Int8[name] = data_v.Interface().(int8)
		case "int":
			r.Normal.Int[name] = data_v.Interface().(int)
		case "int64":
			r.Normal.Int64[name] = data_v.Interface().(int64)
		case "float32":
			r.Normal.Float32[name] = data_v.Interface().(float32)
		case "float64":
			r.Normal.Float64[name] = data_v.Interface().(float64)
		case "complex64":
			r.Normal.Complex64[name] = data_v.Interface().(complex64)
		case "complex128":
			r.Normal.Complex128[name] = data_v.Interface().(complex128)
		default:
			err = fmt.Errorf("roles[RoleMiddleData]SetData: Unsupported data type %v", data_type_name)
			return
		}
	}
	return
}

// 获取类型
func (r *RoleMiddleData) GetDataType(typename string) (data interface{}, err error) {
	switch typename {
	case "time.Time":
		d := time.Now()
		return d, nil
	case "[]byte":
		return make([]byte, 0), nil
	case "string":
		d := ""
		return d, nil
	case "bool":
		return true, nil
	case "uint8":
		var d uint8 = 0
		return d, nil
	case "uint":
		var d uint = 0
		return d, nil
	case "uint64":
		var d uint64 = 0
		return d, nil
	case "int8":
		var d int8 = 0
		return d, nil
	case "int":
		var d int = 0
		return d, nil
	case "int64":
		var d int64 = 0
		return d, nil
	case "float32":
		var d float32 = 0.1
		return d, nil
	case "float64":
		var d float64 = 0.1
		return d, nil
	case "complex64":
		var d complex64 = complex(0, 0)
		return d, nil
	case "complex128":
		var d complex128 = complex(0, 0)
		return d, nil

	case "[]string":
		d := make([]string, 0)
		return d, nil
	case "[]bool":
		d := make([]bool, 0)
		return d, nil
	case "[]uint8":
		d := make([]uint8, 0)
		return d, nil
	case "[]uint":
		d := make([]uint, 0)
		return d, nil
	case "[]uint64":
		d := make([]uint64, 0)
		return d, nil
	case "[]int8":
		d := make([]int8, 0)
		return d, nil
	case "[]int":
		d := make([]int, 0)
		return d, nil
	case "[]int64":
		d := make([]int64, 0)
		return d, nil
	case "[]float32":
		d := make([]float32, 0)
		return d, nil
	case "[]float64":
		d := make([]float64, 0)
		return d, nil
	case "[]complex64":
		d := make([]complex64, 0)
		return d, nil
	case "[]complex128":
		d := make([]complex128, 0)
		return d, nil

	case "map[string]string":
		d := make(map[string]string)
		return d, nil
	case "map[string]bool":
		d := make(map[string]bool)
		return d, nil
	case "map[string]uint8":
		d := make(map[string]uint8)
		return d, nil
	case "map[string]uint":
		d := make(map[string]uint)
		return d, nil
	case "map[string]uint64":
		d := make(map[string]uint64)
		return d, nil
	case "map[string]int8":
		d := make(map[string]int8)
		return d, nil
	case "map[string]int":
		d := make(map[string]int)
		return d, nil
	case "map[string]int64":
		d := make(map[string]int64)
		return d, nil
	case "map[string]float32":
		d := make(map[string]float32)
		return d, nil
	case "map[string]float64":
		d := make(map[string]float64)
		return d, nil
	case "map[string]complex64":
		d := make(map[string]complex64)
		return d, nil
	case "map[string]complex128":
		d := make(map[string]complex128)
		return d, nil

	default:
		err = fmt.Errorf("roles[RoleMiddleData]SetDataFromNoType: Can't find the data.")
		return
	}

	return
}

// 获取类型
func (r *RoleMiddleData) SetDataFromByte(name, typename string, data_b []byte) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]SetDataFromByte: %v", e)
		}
	}()

	b_buf := bytes.NewBuffer(data_b) //将[]byte放入bytes的buffer中
	b_go := gob.NewDecoder(b_buf)    //将buffer放入gob的decoder中
	switch typename {
	case "time.Time":
		d := time.Now()
		err = b_go.Decode(&d)
		r.Normal.Time[name] = d
	case "[]byte":
		d := make([]byte, 0)
		err = b_go.Decode(&d)
		r.Normal.Byte[name] = d
	case "string":
		d := ""
		err = b_go.Decode(&d)
		r.Normal.String[name] = d
	case "bool":
		d := true
		err = b_go.Decode(&d)
		r.Normal.Bool[name] = d
	case "uint8":
		var d uint8 = 0
		err = b_go.Decode(&d)
		r.Normal.Uint8[name] = d
	case "uint":
		var d uint = 0
		err = b_go.Decode(&d)
		r.Normal.Uint[name] = d
	case "uint64":
		var d uint64 = 0
		err = b_go.Decode(&d)
		r.Normal.Uint64[name] = d
	case "int8":
		var d int8 = 0
		err = b_go.Decode(&d)
		r.Normal.Int8[name] = d
	case "int":
		var d int = 0
		err = b_go.Decode(&d)
		r.Normal.Int[name] = d
	case "int64":
		var d int64 = 0
		err = b_go.Decode(&d)
		r.Normal.Int64[name] = d
	case "float32":
		var d float32 = 0.0
		err = b_go.Decode(&d)
		r.Normal.Float32[name] = d
	case "float64":
		var d float64 = 0.0
		err = b_go.Decode(&d)
		r.Normal.Float64[name] = d
	case "complex64":
		var d complex64 = complex(0, 0)
		err = b_go.Decode(&d)
		r.Normal.Complex64[name] = d
	case "complex128":
		var d complex128 = complex(0, 0)
		err = b_go.Decode(&d)
		r.Normal.Complex128[name] = d

	case "[]string":
		d := make([]string, 0)
		err = b_go.Decode(&d)
		r.Slice.String[name] = d
	case "[]bool":
		d := make([]bool, 0)
		err = b_go.Decode(&d)
		r.Slice.Bool[name] = d
	case "[]uint8":
		d := make([]uint8, 0)
		err = b_go.Decode(&d)
		r.Slice.Uint8[name] = d
	case "[]uint":
		d := make([]uint, 0)
		err = b_go.Decode(&d)
		r.Slice.Uint[name] = d
	case "[]uint64":
		d := make([]uint64, 0)
		err = b_go.Decode(&d)
		r.Slice.Uint64[name] = d
	case "[]int8":
		d := make([]int8, 0)
		err = b_go.Decode(&d)
		r.Slice.Int8[name] = d
	case "[]int":
		d := make([]int, 0)
		err = b_go.Decode(&d)
		r.Slice.Int[name] = d
	case "[]int64":
		d := make([]int64, 0)
		err = b_go.Decode(&d)
		r.Slice.Int64[name] = d
	case "[]float32":
		d := make([]float32, 0)
		err = b_go.Decode(&d)
		r.Slice.Float32[name] = d
	case "[]float64":
		d := make([]float64, 0)
		err = b_go.Decode(&d)
		r.Slice.Float64[name] = d
	case "[]complex64":
		d := make([]complex64, 0)
		err = b_go.Decode(&d)
		r.Slice.Complex64[name] = d
	case "[]complex128":
		d := make([]complex128, 0)
		err = b_go.Decode(&d)
		r.Slice.Complex128[name] = d

	case "map[string]string":
		d := make(map[string]string)
		err = b_go.Decode(&d)
		r.StringMap.String[name] = d
	case "map[string]bool":
		d := make(map[string]bool)
		err = b_go.Decode(&d)
		r.StringMap.Bool[name] = d
	case "map[string]uint8":
		d := make(map[string]uint8)
		err = b_go.Decode(&d)
		r.StringMap.Uint8[name] = d
	case "map[string]uint":
		d := make(map[string]uint)
		err = b_go.Decode(&d)
		r.StringMap.Uint[name] = d
	case "map[string]uint64":
		d := make(map[string]uint64)
		err = b_go.Decode(&d)
		r.StringMap.Uint64[name] = d
	case "map[string]int8":
		d := make(map[string]int8)
		err = b_go.Decode(&d)
		r.StringMap.Int8[name] = d
	case "map[string]int":
		d := make(map[string]int)
		err = b_go.Decode(&d)
		r.StringMap.Int[name] = d
	case "map[string]int64":
		d := make(map[string]int64)
		err = b_go.Decode(&d)
		r.StringMap.Int64[name] = d
	case "map[string]float32":
		d := make(map[string]float32)
		err = b_go.Decode(&d)
		r.StringMap.Float32[name] = d
	case "map[string]float64":
		d := make(map[string]float64)
		err = b_go.Decode(&d)
		r.StringMap.Float64[name] = d
	case "map[string]complex64":
		d := make(map[string]complex64)
		err = b_go.Decode(&d)
		r.StringMap.Complex64[name] = d
	case "map[string]complex128":
		d := make(map[string]complex128)
		err = b_go.Decode(&d)
		r.StringMap.Complex128[name] = d

	default:
		err = fmt.Errorf("Can't find the data.")
	}
	if err != nil {
		err = fmt.Errorf("roles[RoleMiddleData]SetDataFromByte: %v", err)
	}
	return
}
