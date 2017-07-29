// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package roles

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/idcsource/insight00-lib/iendecode"
)

// 获取自己的版本
func (r *RoleMiddleData) ReturnVersion() (version int) {
	return r.Version.Version
}

// 设置自己的版本（通常这个是在存储（如HardStore）来处理的时候才需要）
func (r *RoleMiddleData) SetVersion(version int) {
	r.Version.Version = version
	r.VersionChange = true
}

// 返回角色自身的ID
func (r *RoleMiddleData) ReturnId() string {
	return r.Version.Id
}

// 设置角色自身的ID
func (r *RoleMiddleData) SetId(id string) {
	r.Version.Id = id
	r.VersionChange = true
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
	r.RelationChange = true
}

// 重置子关系，也就是将子关系清空
func (r *RoleMiddleData) ResetChilren() {
	r.Relation.Children = make([]string, 0)
	r.RelationChange = true
}

// 重置朋友关系，也就是将朋友关系清空
func (r *RoleMiddleData) ResetFriends() {
	r.Relation.Friends = make(map[string]Status)
	r.RelationChange = true
}

// 设置父关系
func (r *RoleMiddleData) SetFather(id string) {
	r.Relation.Father = id
	r.RelationChange = true
}

// 设置整个子关系
func (r *RoleMiddleData) SetChildren(children []string) {
	r.Relation.Children = children
	r.RelationChange = true
}

// 设置整个朋友关系
func (r *RoleMiddleData) SetFriends(friends map[string]Status) {
	r.Relation.Friends = friends
	r.RelationChange = true
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
		r.RelationChange = true
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
		r.RelationChange = true
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
	r.RelationChange = true
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
	r.RelationChange = true
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
	r.RelationChange = true
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
	r.RelationChange = true
}

// 获取全部上下文，存储实例调用
func (r *RoleMiddleData) GetContexts() map[string]Context {
	return r.Relation.Contexts
}

// 创建一个空的上下文，如果已经存在则忽略
func (r *RoleMiddleData) NewContext(contextname string) (err error) {
	_, find := r.Relation.Contexts[contextname]
	if find == false {
		r.Relation.Contexts[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	} else {
		//err = fmt.Errorf("The context already exist.")
	}
	r.RelationChange = true
	return
}

// 是否存在一个上下文
func (r *RoleMiddleData) ExistContext(contextname string) (have bool) {
	_, have = r.Relation.Contexts[contextname]
	return have
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
	r.RelationChange = true
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
	r.RelationChange = true
}

// 删除一个上下文的上游
func (r *RoleMiddleData) DelContextUp(contextname, upname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == true {
		if _, find2 := r.Relation.Contexts[contextname].Up[upname]; find2 == true {
			delete(r.Relation.Contexts[contextname].Up, upname)
		}
	}
	r.RelationChange = true
}

// 删除一个上下文的下游
func (r *RoleMiddleData) DelContextDown(contextname, downname string) {
	_, find := r.Relation.Contexts[contextname]
	if find == true {
		if _, find2 := r.Relation.Contexts[contextname].Down[downname]; find2 == true {
			delete(r.Relation.Contexts[contextname].Down, downname)
		}
	}
	r.RelationChange = true
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
		r.Relation.Friends[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
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
	case "string":
		r.Relation.Friends[id].String[bit] = valuer.String()
	default:
		return fmt.Errorf("roles[RoleMiddleData]SetFriendStatus: The value's type must int64, float64, complex128 or string.")
	}
	r.RelationChange = true
	return nil
}

// 获取朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) GetFriendStatus(id string, bit int, value interface{}) (have bool, err error) {
	have = true
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: %v", e)
		}
	}()
	_, findf := r.Relation.Friends[id]
	if findf == false {
		//return fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: Role has no exist : " + id + " in " + r.Version.Id + " friend .")
		have = false
		return
	}
	if bit > 9 {
		err = fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: The bit must less than 10.")
		return
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
	case "string":
		valuer.SetString(r.Relation.Friends[id].String[bit])
	default:
		err = fmt.Errorf("roles[RoleMiddleData]GetFriendStatus: The value's type must int64, float64, complex128 or string.")
	}
	return
}

// 设置上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) SetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (err error) {
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
			r.Relation.Contexts[contextname].Up[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
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
		case "string":
			r.Relation.Contexts[contextname].Up[id].String[bit] = valuer.String()
		default:
			return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The value's type must int64, float64, complex128 or string.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r.Relation.Contexts[contextname].Down[id]
		if findr == false {
			r.Relation.Contexts[contextname].Down[id] = Status{Int: make([]int64, 10), Float: make([]float64, 10), Complex: make([]complex128, 10), String: make([]string, 10)}
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
		case "string":
			r.Relation.Contexts[contextname].Down[id].String[bit] = valuer.String()
		default:
			return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The value's type must int64, float64, complex128 or string.")
		}
	} else {
		return fmt.Errorf("roles:[RoleMiddleData]SetContextStatus:The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	r.RelationChange = true
	return nil
}

// 获取上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *RoleMiddleData) GetContextStatus(contextname string, upordown ContextUpDown, id string, bit int, value interface{}) (have bool, err error) {
	have = true
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]GetContextStatus: %v", e)
		}
	}()
	if bit > 9 {
		err = errors.New("roles[RoleMiddleData]GetContextStatus: The bit must less than 10.")
		return
	}
	_, findc := r.Relation.Contexts[contextname]
	if findc == false {
		err = errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no context " + contextname + " in " + r.Version.Id + " .")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := r.Relation.Contexts[contextname].Up[id]
		if findr == false {
			//return errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no up context relationship " + id + " in " + contextname + " in " + r.Version.Id + " .")
			have = false
			return
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
		case "string":
			valuer.SetString(r.Relation.Contexts[contextname].Up[id].String[bit])
		default:
			err = errors.New("roles[RoleMiddleData]GetContextStatus: The value's type must int64, float64, complex128 or string.")
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r.Relation.Contexts[contextname].Down[id]
		if findr == false {
			//return errors.New("roles[RoleMiddleData]GetContextStatus: The Role have no down context relationship " + id + " in " + contextname + " in " + r.Version.Id + " .")
			have = false
			return
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
		case "string":
			valuer.SetString(r.Relation.Contexts[contextname].Down[id].String[bit])
		default:
			err = errors.New("roles[RoleMiddleData]GetContextStatus: The value's type must int64, float64, complex128 or string.")
		}
	} else {
		err = errors.New("roles[RoleMiddleData]GetContextStatus: The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

//func (r *RoleMiddleData) GetDataToByte(name string) (b []byte, err error) {
//	var find bool
//	b, find = r.Data.Point[name]
//	if find == false {
//		err = fmt.Errorf("roles[RoleMiddleData]GetDataToByte: Can not find the data.")
//	}
//	return
//}

// 中间类型的获取数据
func (r *RoleMiddleData) GetData(name string, datas interface{}) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddle]GetData: %v", e)
		}
	}()

	// 获取data的数据类型
	_, find := r.Data.Point[name]
	if find == false {
		err = fmt.Errorf("roles[RoleMiddleData]GetData: Can not find the field %v", name)
		return
	}
	datas_v := reflect.ValueOf(datas)
	datas_t := datas_v.Type().String()
	if in := typeWithIn(datas_t); in == true {
		fv := reflect.ValueOf(r.Data.Point[name])
		datas_v.Set(fv)
	} else {
		err = iendecode.BytesGobStruct(r.Data.Point[name].([]byte), datas)
	}
	if err != nil {
		err = fmt.Errorf("roles[RoleMiddleData]GetData: %v", err)
	}
	return
}

// 往中间类型设置数据
func (r *RoleMiddleData) SetData(name string, datas interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddle]SetData: %v", e)
		}
	}()
	datas_v := reflect.ValueOf(datas)
	datas_t := datas_v.Type().String()
	if in := typeWithIn(datas_t); in == true {
		r.Data.Point[name] = datas
	} else {
		data_b, err := iendecode.StructGobBytes(datas)
		if err != nil {
			err = fmt.Errorf("roles[RoleMiddleData]SetData: %v", err)
			return err
		}
		r.Data.Point[name] = data_b
	}
	r.DataChange = true
	return
}

// 从[]byte设置数据点数据
//func (r *RoleMiddleData) SetDataFromByte(name string, data_b []byte) (err error) {
//	r.Data.Point[name] = data_b
//	r.DataChange = true
//	return
//}
