// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 角色（Role）概念封装的数据存储与数据关系
package roles

import (
	"errors"
	"reflect"
	"fmt"
)

const (
	// Father be changed
	FATHER_CHANGED		= iota
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

const (
	// 上下文上游
	CONTEXT_UP		= iota
	// 上下文下游
	CONTEXT_DOWN
)

// Role为基本角色类型。
// 此类型实现了Roleer接口，
// 应被所有用户自定义角色类型所继承。
type Role struct {
	// 角色ID
	Id						string
	// 角色的版本号
	_role_version			int
	// 父角色（拓扑结构层面）
	_father					string
	// 虚拟的子角色群，只保存键名
	_children				[]string
	// 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	_friends				map[string]Status
	// 上下文关系列表
	_context				map[string]Context
	// 父角色被更改
	_father_changed			bool
	// 子角色关系被改变
	_children_changed		bool
	// 朋友角色被改变
	_friends_changed		bool
	// 上下文关系改变
	_context_changed		bool
	// 数据被改变
	_data_changed			bool
	// 是否被删除
	_be_delete				bool
}

// 句柄上下文的结构
type Context struct {
	// 上游
	Up						map[string]Status
	// 下游
	Down					map[string]Status
}

// 状态的数据结构
type Status struct {
	Int						[]int64
	Float					[]float64
	Complex					[]complex128
}

// 新建自己
func (r *Role) New (id string){
	r.Id = id;
	r._role_version = 2;
	r._children = make([]string,0);
	r._friends = make(map[string]Status);
	r._context = make(map[string]Context);
	r._father_changed = false;
	r._children_changed = false;
	r._friends_changed = false;
	r._context_changed = false;
	r._data_changed = true;
	r._be_delete = false;
}

// 获取自己的版本
func (r *Role) Version () (version int) {
	return r._role_version;
}

// 设置自己的版本（通常这个是在存储（如HardStore）来处理的时候才需要甬道）
func (r *Role) SetVersion (version int) {
	r._role_version = version;
}

// 返回status代表的部分是否被改变，被改变则返回true
func (r *Role) ReturnChanged (status uint8) bool {
	switch status {
		case FATHER_CHANGED:
			return r._father_changed;
		case CHILDREN_CHANGED:
			return r._children_changed;
		case FRIENDS_CHANGED:
			return r._friends_changed;
		case DATA_CHANGED:
			return r._data_changed;
		case CONTEXT_CHANGED:
			return r._context_changed;
		case SELF_CHANGED:
			if r._father_changed == true {
				return true;
			} else if r._children_changed == true {
				return true;
			} else if r._friends_changed == true {
				return true;
			} else if r._context_changed == true {
				return true;
			} else {
				return false;
			}
		default:
			return false;
	}
}

// 设置数据体被改变，这个方法应该由RolesControl来调用
func (r *Role) SetDataChanged () {
	r._data_changed = true;
}

// 返回角色自身的ID
func (r *Role) ReturnId() string {
	return r.Id;
}

// 返回自己的父亲是谁
func (r *Role) GetFather() string{
	return r._father;
}

// 返回整个子角色关系
func (r *Role) GetChildren() []string{
	return r._children;
}

// 返回整个朋友关系
func (r *Role) GetFriends() map[string]Status{
	return r._friends;
}

// 重置父关系，也就是将父关系清空
func (r *Role) ResetFather() {
	r._father = "";
	r._father_changed = true;
}

// 重置子关系，也就是将子关系清空
func (r *Role) ResetChilren() {
	r._children = make([]string,0);
	r._children_changed = true;
}

// 重置朋友关系，也就是将朋友关系清空
func (r *Role) ResetFriends() {
	r._friends = make(map[string]Status);
	r._friends_changed = true;
}

// 设置父关系
func (r *Role) SetFather (id string){
	r._father = id;
	r._father_changed = true;
}

// 设置整个子关系
func (r *Role) SetChildren (children []string){
	r._children = children;
	r._children_changed = true;
}

// 设置整个朋友关系
func (r *Role) SetFriends (friends map[string]Status){
	r._friends = friends;
	r._friends_changed = true;
}

// 看是否存在某个子角色，如果存在返回true
func (r *Role) ExistChild (name string) bool{
	for _, v := range r._children {
		if v == name {
			return true;
			break;
		}
	}
	return false;
}

// 是否存在某个朋友，并且返回这个朋友的疏离关系
func (r *Role) ExistFriend (name string) (bool, int64){
	v, FindV := r._friends[name];
	if FindV == true {
		return true, v.Int[0];
	}
	return false, 0;
}

// 将一个子角色添加进去
func (r *Role) AddChild (cid string) error {
	exist := r.ExistChild(cid);
	if exist == true {
		err := errors.New("This Role has already exist : " + cid + " in " + r.Id + " _children .");
		return err;
	}else{
		r._children = append(r._children, cid);
		r._children_changed = true;
		return nil;
	}
}

// 删除一个子角色
func (r *Role) DeleteChild (child string) error {
	exist := r.ExistChild(child);
	if exist != true {
		err := errors.New("Role has no exist : " + child + " in " + r.Id + " _children .");
		return err;
	}else{
		var count int;
		for i, v := range r._children {
			if v == child {
				count = i;
				break;
			}
		}
		r._children = append(r._children[:count],r._children[count+1:]...);
		r._children_changed = true;
		return nil;
	}
}

// 添加一个朋友关系
func (r *Role) AddFriend (id string, bind int64) error {
	// 检查这个friend是否存在
	ifexist, _:= r.ExistFriend(id);
	if ifexist == true {
		err := errors.New("This Role has already exist : " + id + " in " + r.Id + " friend .");
		return err;
	}
	r._friends[id] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
	r._friends[id].Int[0] = bind;
	r._friends_changed = true;
	return nil;
}

// 删除一个朋友关系
func (r *Role) DeleteFriend (id string) error {
	ifexist, _  := r.ExistFriend(id);
	if ifexist == true {
		//err := Error(1, id, "in", r.Id, "friend .");
		err := errors.New("Role has no exist : " + id + " in " + r.Id + " friend .");
		return err;
	}
	delete(r._friends, id);
	r._friends_changed = true;
	return nil;
}

// 修改一个朋友关系，只是改关系远近
func (r *Role) ChangeFriend (id string, bind int64) error {
	ifexist, obind  := r.ExistFriend(id);
	if ifexist == true {
		err := errors.New("Role has no exist : " + id + " in " + r.Id + " friend .");
		return err;
	}
	if obind == bind {
		return nil;
	}
	r._friends[id].Int[0] = bind;
	r._friends_changed = true;
	return nil;
}

// 获取相同远近关系下的所有朋友的ID
func (r *Role) GetSameBindFriendsId (bind int64) []string{
	rstr := make([]string,0);
	for id, binds := range r._friends {
		if binds.Int[0] == bind {
			rstr = append(rstr,id);
		}
	}
	return rstr;
}

// 查看删除状态
func (r *Role) ReturnDelete () bool {
	return r._be_delete;
}

// 设置删除状态，true则为删除
func (r *Role) SetDelete (del bool) {
	r._be_delete = del;
}

// 设定全部上下文，存储实例调用
func (r *Role) SetContexts (context map[string]Context) {
	r._context = context;
}

// 获取全部上下文，存储实例调用
func (r *Role) GetContexts () map[string]Context {
	return r._context;
}

// 创建一个空的上下文，如果已经存在则忽略
func (r *Role) NewContext (contextname string) {
	_ , find := r._context[contextname];
	if find == false {
		r._context[contextname] = Context{
			Up : make(map[string]Status),
			Down : make(map[string]Status),
		};
	}
}

// 设定一个上下文的上游
func (r *Role) AddContextUp (contextname, upname string, bind int64){
	_ , find := r._context[contextname];
	if find == false {
		r._context[contextname] = Context{
			Up : make(map[string]Status),
			Down : make(map[string]Status),
		};
		r._context[contextname].Up[upname] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
	}
	r._context[contextname].Up[upname].Int[0] = bind;
	r._context_changed = true;
}

// 设定一个上下文的下游
func (r *Role) AddContextDown (contextname, downname string, bind int64) {
	_ , find := r._context[contextname];
	if find == false {
		r._context[contextname] = Context{
			Up : make(map[string]Status),
			Down : make(map[string]Status),
		};
		r._context[contextname].Down[downname] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
	}
	r._context[contextname].Down[downname].Int[0] = bind;
	r._context_changed = true;
}

// 删除一个上下文的上游
func (r *Role) DelContextUp (contextname, upname string) {
	_ , find := r._context[contextname];
	if find == true {
		if _, find2 := r._context[contextname].Up[upname]; find2 == true {
			delete(r._context[contextname].Up, upname);
		}
	}
	r._context_changed = true;
}

// 删除一个上下文的下游
func (r *Role) DelContextDown (contextname, downname string) {
	_ , find := r._context[contextname];
	if find == true {
		if _, find2 := r._context[contextname].Down[downname]; find2 == true {
			delete(r._context[contextname].Down, downname);
		}
	}
	r._context_changed = true;
}

// 清除一个上下文
func (r *Role) DelContext (contextname string) {
	_ , find := r._context[contextname];
	if find == true {
		delete(r._context, contextname);
	}
}

// 找到一个上下文的上文，返回绑定值
func (r *Role) GetContextUp (contextname, upname string) (bind int64, have bool) {
	if _ , have = r._context[contextname]; have == true {
		var binds Status;
		if binds, have = r._context[contextname].Up[upname]; have == true {
			bind = binds.Int[0];
			return;
		}else{
			have = false;
			return;
		}
	} else {
		have = false;
		return;
	}
}

// 找到一个上下文的下文，返回绑定值
func (r *Role) GetContextDown (contextname, downname string) (bind int64, have bool) {
	if _ , have = r._context[contextname]; have == true {
		var binds Status;
		if binds, have = r._context[contextname].Down[downname]; have == true {
			bind = binds.Int[0];
			return;
		}else{
			have = false;
			return;
		}
	} else {
		have = false;
		return;
	}
}

// 返回某个上下文的全部信息
func (r *Role) GetContext (contextname string) (context Context, have bool) {
	if context , have = r._context[contextname]; have == true {
		return;
	} else {
		have = false;
		return;
	}
}

// 返回某个上下文中的上游同样绑定值的所有
func (r *Role) GetContextUpSameBind (contextname string, bind int64) (rolesid []string, have bool) {
	if _ , find := r._context[contextname]; find == true {
		rolesid = make([]string,0);
		for id, binds := range r._context[contextname].Up {
			if binds.Int[0] == bind {
				rolesid = append(rolesid, id);
			}
		}
		have = true;
		return;
	}else{
		have = false;
		return;
	}
}

// 返回某个上下文中的下游同样绑定值的所有
func (r *Role) GetContextDownSameBind (contextname string, bind int64) (rolesid []string, have bool) {
	if _ , find := r._context[contextname]; find == true {
		rolesid = make([]string,0);
		for id, binds := range r._context[contextname].Down {
			if binds.Int[0] == bind {
				rolesid = append(rolesid, id);
			}
		}
		have = true;
		return;
	}else{
		have = false;
		return;
	}
}

// 返回所有上下文组的名称
func (r *Role) GetContextsName () (names []string) {
	lens := len(r._context)
	names = make([]string,lens);
	i := 0;
	for name, _ := range r._context {
		names[i] = name;
		i++;
	}
	return names;
}


// 设置朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) SetFriendStatus (id string, bit int, value interface{}) (err error) {
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	_, findf := r._friends[id];
	if findf == false {
		r._friends[id] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
	}
	if bit > 9 {
		return errors.New("The bit must less than 10.");
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String();
	switch vname {
		case "int":
			r._friends[id].Int[bit] = valuer.Int();
		case "int64":
			r._friends[id].Int[bit] = valuer.Int();
		case "float":
			r._friends[id].Float[bit] = valuer.Float();
		case "float64":
			r._friends[id].Float[bit] = valuer.Float();
		case "complex64":
			r._friends[id].Complex[bit] = valuer.Complex();
		case "complex128":
			r._friends[id].Complex[bit] = valuer.Complex();
		default :
			return errors.New("The value's type must int64, float64 or complex128.");
	}
	return nil;
}

// 获取朋友的状态属性，id：朋友的ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) GetFriendStatus (id string, bit int, value interface{}) (err error) {
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	_, findf := r._friends[id];
	if findf == false {
		return errors.New("Role has no exist : " + id + " in " + r.Id + " friend .");
	}
	if bit > 9 {
		return errors.New("The bit must less than 10.");
	}
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String();
	switch vname {
		case "int64":
			valuer.SetInt(r._friends[id].Int[bit]);
		case "float64":
			valuer.SetFloat(r._friends[id].Float[bit]);
		case "complex128":
			valuer.SetComplex(r._friends[id].Complex[bit]);
		default :
			return errors.New("The value's type must int64, float64 or complex128.");
	}
	return nil;
}

// 设置上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) SetContextStatus (contextname string, upordown uint8, id string, bit int, value interface{}) (err error) {
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	if bit > 9 {
		return errors.New("The bit must less than 10.");
	}
	_, findc := r._context[contextname];
	if findc == false {
		r._context[contextname] = Context{
			Up : make(map[string]Status),
			Down : make(map[string]Status),
		};
	}
	if upordown == CONTEXT_UP {
		_, findr := r._context[contextname].Up[id];
		if findr == false {
			r._context[contextname].Up[id] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String();
		switch vname {
			case "int":
				r._context[contextname].Up[id].Int[bit] = valuer.Int();
			case "int64":
				r._context[contextname].Up[id].Int[bit] = valuer.Int();
			case "float":
				r._context[contextname].Up[id].Float[bit] = valuer.Float();
			case "float64":
				r._context[contextname].Up[id].Float[bit] = valuer.Float();
			case "complex64":
				r._context[contextname].Up[id].Complex[bit] = valuer.Complex();
			case "complex128":
				r._context[contextname].Up[id].Complex[bit] = valuer.Complex();
			default :
				return errors.New("The value's type must int64, float64 or complex128.");
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r._context[contextname].Down[id];
		if findr == false {
			r._context[contextname].Down[id] = Status{Int: make([]int64,10), Float: make([]float64,10), Complex: make([]complex128, 10)};
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String();
		switch vname {
			case "int":
				r._context[contextname].Down[id].Int[bit] = valuer.Int();
			case "int64":
				r._context[contextname].Down[id].Int[bit] = valuer.Int();
			case "float":
				r._context[contextname].Down[id].Float[bit] = valuer.Float();
			case "float64":
				r._context[contextname].Down[id].Float[bit] = valuer.Float();
			case "complex64":
				r._context[contextname].Down[id].Complex[bit] = valuer.Complex();
			case "complex128":
				r._context[contextname].Down[id].Complex[bit] = valuer.Complex();
			default :
				return errors.New("The value's type must int64, float64 or complex128.");
		}
	} else {
		return errors.New("The upordown must CONTEXT_UP or CONTEXT_DOWN.");
	}
	return nil;
}

// 获取上下文的状态属性，contextname：上下文名称；upordown：上游还是下游，使用常量CONTEXT_UP或CONTEXT_DOWN；id：角色ID；bit：位，0到9；value值，只接收int64、float64、complex128。
func (r *Role) GetContextStatus (contextname string, upordown uint8, id string, bit int, value interface{}) (err error) {
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	if bit > 9 {
		return errors.New("The bit must less than 10.");
	}
	_, findc := r._context[contextname];
	if findc == false {
		return errors.New("The Role have no context " + contextname + " in " + r.Id + " .");
	}
	if upordown == CONTEXT_UP {
		_, findr := r._context[contextname].Up[id];
		if findr == false {
			return errors.New("The Role have no up context relationship " + id + " in " + contextname + " in " + r.Id + " .");
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String();
		switch vname {
			case "int64":
				valuer.SetInt(r._context[contextname].Up[id].Int[bit]);
			case "float64":
				valuer.SetFloat(r._context[contextname].Up[id].Float[bit]);
			case "complex128":
				valuer.SetComplex(r._context[contextname].Up[id].Complex[bit]);
			default :
				return errors.New("The value's type must int64, float64 or complex128.");
		}
	} else if upordown == CONTEXT_DOWN {
		_, findr := r._context[contextname].Down[id];
		if findr == false {
			return errors.New("The Role have no down context relationship " + id + " in " + contextname + " in " + r.Id + " .");
		}
		valuer := reflect.Indirect(reflect.ValueOf(value))
		vname := valuer.Type().String();
		switch vname {
			case "int64":
				valuer.SetInt(r._context[contextname].Down[id].Int[bit]);
			case "float64":
				valuer.SetFloat(r._context[contextname].Down[id].Float[bit]);
			case "complex128":
				valuer.SetComplex(r._context[contextname].Down[id].Complex[bit]);
			default :
				return errors.New("The value's type must int64, float64 or complex128.");
		}
	} else {
		return errors.New("The upordown must CONTEXT_UP or CONTEXT_DOWN.");
	}
	return nil;
}
