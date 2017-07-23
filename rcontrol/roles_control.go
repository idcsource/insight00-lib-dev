// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// RolesControl是一个角色管理器。
//
// 主要职责包括：
//     1. 角色的生成
//     2. 角色间父子、朋友关系的一致性统一管理
//     3. 角色内数据访问
//     4. 角色及其关系的永久硬存储（磁盘）的处理
// 务必通过RolesControl且通过同一个RolesControl来管理一组相关联的角色。
// 只有通过RolesControl才能保证角色间关系的统一，且它是线程安全的。
//
// 注意一定要定期ToStore，只有ToStore的过程才能将信息保存进永久存储中，
// 否则内存中的数据会因停电等故障而消失。
// ToStore也是线程安全的。
//
// TODO: 更小的锁粒度
package rcontrol

import(
	"reflect"
	"fmt"
	"errors"
	"sync"
	
	"github.com/idcsource/insight00-lib/roles"
	"github.com/idcsource/insight00-lib/rolesio"
	"github.com/idcsource/insight00-lib/ilogs"
)


type RolesControl struct {
	// 角色的永久硬存储的IO方法，RolesInOutManager接口
	readWrite				rolesio.RolesInOutManager

	// 最大缓存角色数
	cacheMax				int64
	// 缓存数量
	rolesCount				int64
	// 缓存满的触发
	cacheIsFull				chan bool
	
	// 角色缓存
	rolesCache				map[string]roles.Roleer
	// 删除缓存
	deleteCache				[]string
	
	logs					*ilogs.Logs
	
	// 读写锁
	lock					*sync.RWMutex
	// 检查缓存数量中
	checkCacheNumOn			bool
}

// 新建一个角色管理器。
// 需要提供符合RolesInOutManager接口的存储实例。
// max为缓存的大小，也就是缓存多个角色对象，0为不限制。当缓存满了之后，将自动调用存储对缓存进行保存，并清空缓存。
func NewRolesControl (ro rolesio.RolesInOutManager, max int64, logs *ilogs.Logs) *RolesControl{
	rb := &RolesControl{
		readWrite: ro,
		rolesCache : make(map[string]roles.Roleer),
		deleteCache: make([]string,0),
		cacheMax : max,
		rolesCount : 0,
		cacheIsFull : make(chan bool),
		logs : logs,
		lock: new(sync.RWMutex),
		checkCacheNumOn: false,
	};
	if rb.cacheMax > 0 {
		go rb.runtimeStore();
	}
	return rb;
}

// 查看是否被标记删除，标记删除则返回true。
func (rc *RolesControl) checkDel (role roles.Roleer) bool {
	if role == nil {
		return true;
	}
	id := role.ReturnId();
	del := rc.checkDelById(id);
	if del == true {
		return del;
	}
	del = role.ReturnDelete();
	return del;
}

func (rc *RolesControl) checkDelById (id string) bool {
	for _, v := range rc.deleteCache {
		if v == id {
			return true;
		}
	}
	return false;
}

// 检查缓存，如果缓存里没有，则纳入到缓存里
func (rc *RolesControl) checkCache (role roles.Roleer) {
	id := role.ReturnId();
	_, find := rc.rolesCache[id];
	if find == false {
		rc.rolesCache[id] = role;
		rc.rolesCount++;
	}
}

// 检查缓存数，如果超出则执行运行时保存
func (rc *RolesControl) checkCacheNum () {
	if rc.cacheMax > 0 && rc.rolesCount >= rc.cacheMax && rc.checkCacheNumOn == false {
		rc.cacheIsFull <- true;
	}
}

// 新建立一个角色，需使用者自行确保id的非重复性，否则同id的角色可能会被覆盖
func (rc *RolesControl) NewRole (id string, new roles.Roleer) roles.Roleer {
	rc.lock.RLock();
	defer rc.lock.RUnlock();
	go rc.checkCacheNum();
	new.New(id);
	rc.rolesCache[id] = new;
	rc.rolesCount++;
	return new;
}

// 根据id从存储中获取一个角色的本体，如果缓存中有则直接调用缓存，找不到则返回错误
func (rc *RolesControl) GetRole (id string) (roles.Roleer, error) {
	rc.lock.RLock();
	defer rc.lock.RUnlock();
	go rc.checkCacheNum();
	var find bool;
	if find := rc.checkDelById(id); find == true {
		err := errors.New("rcontrol: GetRole: The role " + id + " can't be found, maybe it be deleted !");
		return nil, err;
	}
	role, find := rc.rolesCache[id];
	if find == true {
		return role, nil;
	} else {
		role, err := rc.readWrite.ReadRole(id);
		if err != nil {
			return nil , err;
		}
		rc.rolesCache[id] = role;
		rc.rolesCount++;
		return role, nil;
	}
}

// 删除一个角色，但这里只是从缓存中删除并标记为删除，真正的删除需要等到tostore时才会进行。
// 此方法永远都会返回nil，也就是只要你给予了，它都将在tostore时提交给永久存储去删除。
func (rc *RolesControl) DeleteRole (role roles.Roleer) (err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	rid := role.ReturnId();
	if rc.checkDel(role) == true {
		return;
	}
	if _, find := rc.rolesCache[rid]; find == true {
		delete(rc.rolesCache, rid);
	}
	role.SetDelete(true);
	rc.deleteCache = append(rc.deleteCache, rid);
	role = nil;
	return;
}

// 向role中注册father关系，如果role之前有father关系，则返回错误。
func (rc *RolesControl) RegFather (role, father roles.Roleer) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: RegFather: Cant set selationship for a deleted role .");
	}
	if rc.checkDel(father) == true {
		return errors.New("rcontrol: RegFather: Cant set selationship for a deleted father .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(father);
	
	rid := role.ReturnId();
	fid := father.ReturnId();
	ofa := role.GetFather();
	if len(ofa) != 0  {
		return errors.New("rcontrol: RegFather: The Role " + rid + "already have father ! Please use ChangeFather() !");
	}
	
	role.SetFather(fid);
	if father.ExistChild(rid) == false {
		father.AddChild(rid);
	}
	return nil;
}

// 向role中注册一个子角色关系，如果child已经有了father关系，则返回错误。
func (rc *RolesControl) RegChild (role, child roles.Roleer) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	
	cid := child.ReturnId();
	fid := role.ReturnId();
	ofa := child.GetFather();
	if len(ofa) != 0  {
		return errors.New("rcontrol: RegChild: The Role " + cid + " already have father ! Please use ChangeFather() !");
	}
	
	if rc.checkDel(child) == true {
		return errors.New("rcontrol: RegChild: Cant set selationship for a deleted child .");
	}
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: RegChild: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(child);
	
	if role.ExistChild(cid) == false {
		role.AddChild(cid);
	}
	
	child.SetFather(fid);
	return nil;
}

// 从father角色中删除child关系，如果二者不存在父子关系，将返回错误。
func (rc *RolesControl) DeleteChild (father, child roles.Roleer) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(child) == true {
		return errors.New("rcontrol: DeleteChild: Cant set selationship for a deleted child .");
	}
	if rc.checkDel(father) == true {
		return errors.New("rcontrol: DeleteChild: Cant set selationship for a deleted father .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(father);
	rc.checkCache(child);
	
	fid := father.ReturnId();
	cid := child.ReturnId();
	c_h_f := child.GetFather();		// c_h_f : child have father
	if c_h_f != fid {
		return errors.New("rcontrol: DeleteChild: The child role's father not the given father !");
	}
	f_h_c := father.ExistChild(cid);	// f_h_c : father have child
	if f_h_c == false {
		return errors.New("rcontrol: DeleteChild: The father role not have the given child !");
	}
	child.ResetFather();
	err := father.DeleteChild(cid);
	if err != nil {
		return err;
	}
	return nil;
}

// 将role的father从from变成to，如果role的Father没有配置过，将返回错误。
func (rc *RolesControl) ChangeFather (role, from, to roles.Roleer) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: ChangeFather: Cant set selationship for a deleted role .");
	}
	if rc.checkDel(to) == true {
		return errors.New("rcontrol: ChangeFather: Cant set selationship for a deleted father .");
	}
	if rc.checkDel(from) == true {
		return errors.New("rcontrol: ChangeFather: Cant set selationship for a deleted old father .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(to);
	rc.checkCache(from);
	
	rid := role.ReturnId();
	fid := from.ReturnId();
	tid := to.ReturnId();
	
	r_n_f := role.GetFather();
	if len(r_n_f) == 0 {
		return errors.New("rcontrol: ChangeFather: The Role " + rid + " have no father to change !");
	}
	if r_n_f != fid {
		return errors.New("rcontrol: ChangeFather: The Role's now father not the Father, can't change !");
	}
	
	role.SetFather(tid);
	if to.ExistChild(rid) == false {
		to.AddChild(rid);
	}
	if from.ExistChild(rid) == true {
		from.DeleteChild(rid);
	}
	
	return nil;
}

// 向role中注册一个朋友角色关系，已经有的则忽略。
func (rc *RolesControl) RegFriend (role, friend roles.Roleer, bind int64) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(friend) == true {
		return errors.New("rcontrol: RegFriend: Cant set selationship for a deleted friend .");
	}
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: RegFriend: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(friend);
	
	fid := friend.ReturnId();
	if have, _ := role.ExistFriend(fid) ; have == false {
		role.AddFriend(fid, bind);
	}
	rid := role.ReturnId();
	if have, _ := friend.ExistFriend(rid); have == false{
		friend.AddFriend(rid, bind);
	}
	return nil;
}

// 只是改它们直接的疏离bind关系数值，如果它们不是朋友关系，则返回错误。
func (rc *RolesControl) ChangeFriend (role, friend roles.Roleer, bind int64) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(friend) == true {
		return errors.New("rcontrol: ChangeFriend: Cant set selationship for a deleted friend .");
	}
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: ChangeFriend: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(friend);
	
	rid := role.ReturnId();
	fid := friend.ReturnId();
	rhave , _ := role.ExistFriend(fid);
	fhave , _ := friend.ExistFriend(rid);
	if rhave == false || fhave == false {
		return errors.New("rcontrol: ChangeFriend: The role " + rid + " and the role " + fid + " are not friend .");
	}
	err1 := role.ChangeFriend(fid, bind);
	if err1 != nil { return fmt.Errorf("rcontrol: DeleteFriend: %v", err1) ; }
	err2 := friend.ChangeFriend(rid, bind);
	if err2 != nil { return fmt.Errorf("rcontrol: DeleteFriend: %v", err2) ; }
	return nil;
}

// 将role和friend之间的朋友关系删除，如果它们本不是朋友，则返回错误。
func (rc *RolesControl) DeleteFriend (role, friend roles.Roleer) error {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(friend) == true {
		return errors.New("rcontrol: DeleteFriend: Cant set selationship for a deleted friend .");
	}
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: DeleteFriend: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	rc.checkCache(friend);
	
	rid := role.ReturnId();
	fid := friend.ReturnId();
	rhave , _ := role.ExistFriend(fid);
	fhave , _ := friend.ExistFriend(rid);
	if rhave == false || fhave == false {
		return errors.New("rcontrol: DeleteFriend: The role " + rid + " and the role " + fid + " are not friend .");
	}
	err1 := role.DeleteFriend(fid);
	if err1 != nil { return fmt.Errorf("rcontrol: DeleteFriend: %v", err1) ; }
	err2 := friend.DeleteFriend(rid);
	if err2 != nil { return fmt.Errorf("rcontrol: DeleteFriend: %v", err2) ; }
	return nil;
}

// 从Role存储的数据中根据name取出并装入data中，找不到name的数据，则返回错误。
func (rc *RolesControl) GetData (role roles.Roleer, name string, data interface{}) (err error) {
	rc.lock.RLock();
	defer rc.lock.RUnlock();
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: GetData: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	
	rv := reflect.Indirect(reflect.ValueOf(role)).FieldByName(name);
	rv_type := rv.Type();
	dv := reflect.Indirect(reflect.ValueOf(data));
	dv_type := dv.Type();
	if rv_type != dv_type {
		err = errors.New("rcontrol: GetData: The Data type " + fmt.Sprint(rv_type) + " not assignable to type " + fmt.Sprint(dv_type) + ".");
		return err;
	}
	if dv.CanSet() != true {
		err = errors.New("rcontrol: GetData: The Data type " + fmt.Sprint(dv_type) + " not the type not be set.");
		return err;
	}
	dv.Set(rv);
	return nil;
}

func (rc *RolesControl) GetData2 (role roles.Roleer, name string ) (data interface{}, err error) {
	rc.lock.RLock();
	defer rc.lock.RUnlock();
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: GetData: %v", e);
		}
	}()
	if rc.checkDel(role) == true {
		return data, errors.New("rcontrol: GetData: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	
	rv := reflect.Indirect(reflect.ValueOf(role)).FieldByName(name);
	//rv_type := rv.Type();
	//dv := reflect.Indirect(reflect.ValueOf(data));
	//dv_type := dv.Type();
	/*
	if rv_type != dv_type {
		err = errors.New("rcontrol: GetData: The Data type " + fmt.Sprint(rv_type) + " not assignable to type " + fmt.Sprint(dv_type) + ".");
		return data, err;
	}
	if dv.CanSet() != true {
		err = errors.New("rcontrol: GetData: The Data type " + fmt.Sprint(dv_type) + " not the type not be set.");
		return data, err;
	}*/
	data = rv.Interface();
	//dv.Set(rv);
	return data, nil;
}

// 把data的数据装入role的name值下，如果找不到name，则返回错误。
func (rc *RolesControl) SetData (role roles.Roleer, name string, data interface{}) (err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	defer func(){
		if e := recover(); e != nil {
			err = fmt.Errorf("rcontrol: SetData: %v", e);
		}
	}()
	if rc.checkDel(role) == true {
		return errors.New("rcontrol: SetData: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	rc.checkCache(role);
	
	rv := reflect.Indirect(reflect.ValueOf(role)).FieldByName(name);
	rv_type := rv.Type();
	dv := reflect.Indirect(reflect.ValueOf(data));
	dv_type := dv.Type();
	if rv_type != dv_type {
		err = errors.New("rcontrol: SetData: The Data type " + fmt.Sprint(dv_type) + " not assignable to type " + fmt.Sprint(rv_type) + ".");
		return err;
	}
	if rv.CanSet() != true {
		err = errors.New("rcontrol: SetData:  The Data type " + fmt.Sprint(rv_type) + " not the type not be set.");
		return err;
	}
	rv.Set(dv);
	role.SetDataChanged();
	return nil;
}

// 绑定平等上下文，也就是都绑定在一个context下并且绑定值也是一样的。
// uprole为上层的角色，downrole为下层的角色。
func (rc *RolesControl) ContextBindEqual(contextname string, uprole, downrole roles.Roleer, bind int64) (err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(uprole) == true || rc.checkDel(downrole) == true {
		return errors.New("rcontrol: ContextBindEqual: Cant set selationship for a deleted role .");
	}
	rc.checkCache(uprole);
	rc.checkCache(downrole);
	go rc.checkCacheNum();
	
	uprole.AddContextDown(contextname, downrole.ReturnId(), bind);
	downrole.AddContextUp(contextname, uprole.ReturnId(), bind);
	return nil;
}

// 绑定非平等上下文，也就是都绑定在一个context下并且绑定值不一样。
// uprole为上层的角色，downrole为下层的角色。upbind为uprole对downrole的绑定值，downbind是downrole对uprole的绑定值。
func (rc *RolesControl) ContextBindUnequal(contextname string, uprole, downrole roles.Roleer, upbind, downbind int64) (err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	if rc.checkDel(uprole) == true || rc.checkDel(downrole) == true {
		return errors.New("rcontrol: ContextBindUnequal: Cant set selationship for a deleted role .");
	}
	rc.checkCache(uprole);
	rc.checkCache(downrole);
	go rc.checkCacheNum();
	
	uprole.AddContextDown(contextname, downrole.ReturnId(), upbind);
	downrole.AddContextUp(contextname, uprole.ReturnId(), downbind);
	return nil;
}

// 将目前载入的角色关系全部保存起来，但不会删除缓存
func (rc *RolesControl) ToStore ()  {
	rc.toStore();
}

func (rc *RolesControl) deferCheckCacheNumOn() {
	rc.checkCacheNumOn = false;
}

// 将目前载入的角色关系全部保存起来，内部使用，无锁
func (rc *RolesControl) toStore ()  {
	rc.checkCacheNumOn = true;
	rc.lock.Lock();
	defer rc.deferCheckCacheNumOn();
	defer rc.lock.Unlock();
	// 检查在children和friends关系里面是否有被delete的成分，如果有则在保存前剔除
	for _, role := range rc.rolesCache {
		children := role.GetChildren();
		delchild := make([]string,0);
		for _, name := range children {
			bedel := rc.checkDelById(name);
			if bedel == true {
				delchild = append(delchild, name);
			}
		}
		for _, name := range delchild {
			role.DeleteChild(name);
		}
		friends := role.GetFriends();
		delfriend := make([]string,0);
		for name, _ := range friends {
			bedel := rc.checkDelById(name);
			if bedel == true {
				delfriend = append(delfriend, name);
			}
		}
		for _, name := range delfriend {
			role.DeleteFriend(name);
		}
	}
	
	// 开始调用永久存储方法
	// 首先就是删除要删除的东西
	for _, id := range rc.deleteCache {
		err := rc.readWrite.DeleteRole(id);
		if err != nil && rc.logs != nil {
			rc.logs.ErrLog(err);
		}
	}
	rc.deleteCache = make([]string, 0);		// 重置删除缓存
	// 然后写入每一个Role的本体，以及各种关系
	for _, role := range rc.rolesCache {
		err := rc.readWrite.StoreRole(role);		// 写入本体
		if err != nil && rc.logs != nil {
			rc.logs.ErrLog(err);
		}
	}
	return;
}

// 运行时存储，当缓存超过时执行
func (rc *RolesControl) runtimeStore () {
	for {
		full := <-rc.cacheIsFull;
		if full == true {
			rc.toStore();
			rc.rolesCache = make(map[string]roles.Roleer);
			rc.rolesCount = 0;
		}
	}
}
