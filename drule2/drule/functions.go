// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"time"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 工作状态
func (d *DRule) WorkStatus() (status bool) {
	if d.closed == true {
		status = false
	} else {
		status = true
	}
	return
}

// 登录
func (d *DRule) UserLogin(username, password string) (unid string, authority operator.UserAuthority, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	user_id := USER_PREFIX + username
	// 查看有没有这个用户
	user_have := d.trule.ExistRole(INSIDE_DMZ, user_id)
	if user_have == false {
		errs.Code = operator.DATA_USER_NO_EXIST
		errs.Err = fmt.Errorf("User not exist.")
		return
	}
	// 查看密码
	var password_2 string
	err = d.trule.ReadData(INSIDE_DMZ, user_id, "Password", &password_2)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	if password_2 != password {
		errs.Code = operator.DATA_USER_PASSWORD_WRONG
		errs.Err = fmt.Errorf("User password wrong.")
		return
	}
	// 规整权限
	var auth operator.UserAuthority
	d.trule.ReadData(INSIDE_DMZ, user_id, "Authority", &auth)
	wrable := make(map[string]bool)
	d.trule.ReadData(INSIDE_DMZ, user_id, "WRable", &wrable)

	authority = auth
	unid = random.Unid(1, time.Now().String(), username)

	// 查看是否已经有登录的了，并写入登录的乱七八糟
	_, find := d.loginuser[username]
	if find {
		d.loginuser[username].wrable = wrable
		d.loginuser[username].unid[unid] = time.Now()
	} else {
		loginuser := &loginUser{
			username:  username,
			unid:      make(map[string]time.Time),
			authority: auth,
			wrable:    wrable,
		}
		loginuser.unid[unid] = time.Now()
		d.loginuser[username] = loginuser
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 用户续命
func (d *DRule) UserAddLife(username, unid string) (errs operator.DRuleError) {
	errs = operator.NewDRuleError()
	yes := d.checkUserLogin(username, unid)
	if yes == true {
		errs.Code = operator.DATA_ALL_OK
		return
	} else {
		errs.Code = operator.DATA_USER_NOT_LOGIN
		errs.Err = fmt.Errorf("User not login.")
		return
	}
}

// 新建用户（没有权限验证）
func (d *DRule) UserAdd(newuser *operator.O_DRuleUser) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	user_id := USER_PREFIX + newuser.UserName

	// 查看是否有重名
	if have := d.trule.ExistRole(INSIDE_DMZ, user_id); have == true {
		errs.Code = operator.DATA_USER_EXIST
		errs.Err = fmt.Errorf("user already exist.")
		return
	}

	user_role := &DRuleUser{
		UserName:  newuser.UserName,
		Password:  newuser.Password,
		Email:     newuser.Email,
		Authority: newuser.Authority,
		WRable:    make(map[string]bool),
	}
	user_role.New(user_id)

	tran, _ := d.trule.Begin()
	err = tran.StoreRole(INSIDE_DMZ, user_role)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.WriteChild(INSIDE_DMZ, USER_PREFIX+ROOT_USER, user_id)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.Commit()
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 修改密码（没有权限验证）
func (d *DRule) UserPassword(theuser *operator.O_DRuleUser) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	// 修改密码
	userid := USER_PREFIX + theuser.UserName
	err = d.trule.WriteData(INSIDE_DMZ, userid, "Password", theuser.Password)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 修改邮箱（没有权限验证）
func (d *DRule) UserEmail(theuser *operator.O_DRuleUser) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	// 修改密码
	userid := USER_PREFIX + theuser.UserName
	err = d.trule.WriteData(INSIDE_DMZ, userid, "Email", theuser.Email)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 删除用户（没有权限验证）
func (d *DRule) UserDelete(username string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	userid := USER_PREFIX + username

	tran, _ := d.trule.Begin()
	err = tran.DeleteRole(INSIDE_DMZ, userid)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.DeleteChild(INSIDE_DMZ, USER_PREFIX+ROOT_USER, userid)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.Commit()
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 用户登出
func (d *DRule) UserLogout(o_send *operator.O_OperatorSend) (errs operator.DRuleError) {
	errs = operator.NewDRuleError()

	// 查看用户权限
	_, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs.Code = operator.DATA_USER_NOT_LOGIN
		return
	}
	// 删除相应登陆的信息
	delete(d.loginuser[o_send.User].unid, o_send.Unid)
	if len(d.loginuser[o_send.User].unid) == 0 {
		delete(d.loginuser, o_send.User)
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 当前用户
func (d *DRule) UserNow(username string) (user operator.O_DRuleUser, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()
	user = operator.O_DRuleUser{}

	user_id := USER_PREFIX + username

	tran, _ := d.trule.Begin()
	err = tran.ReadData(INSIDE_DMZ, user_id, "UserName", &user.UserName)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.ReadData(INSIDE_DMZ, user_id, "Email", &user.Email)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.ReadData(INSIDE_DMZ, user_id, "Authority", &user.Authority)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	tran.Commit()
	errs.Code = operator.DATA_ALL_OK
	return
}

// 用户列表（没有权限验证）
func (d *DRule) UserList() (list []operator.O_DRuleUser, errs operator.DRuleError) {
	list = make([]operator.O_DRuleUser, 0)
	errs = operator.NewDRuleError()

	tran, _ := d.trule.Begin()
	children, err := tran.ReadChildren(INSIDE_DMZ, USER_PREFIX+ROOT_USER)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	for _, child := range children {
		one := operator.O_DRuleUser{}
		err = tran.ReadData(INSIDE_DMZ, child, "UserName", &one.UserName)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.ReadData(INSIDE_DMZ, child, "Email", &one.Email)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.ReadData(INSIDE_DMZ, child, "Authority", &one.Authority)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		list = append(list, one)
	}
	tran.Commit()
	errs.Code = operator.DATA_ALL_OK
	return
}

// 新建区域（没有权限验证）
func (d *DRule) AreaAdd(areaname string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	err = d.trule.AreaInit(areaname)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
	} else {
		errs.Code = operator.DATA_ALL_OK
	}
	return
}

// 删除区域（没有权限验证）
func (d *DRule) AreaDelete(areaname string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()
	if areaname == INSIDE_DMZ {
		errs.Code = operator.DATA_USER_NO_AUTHORITY
		errs.Err = fmt.Errorf("cannot delete this area.")
		return
	}
	err = d.trule.AreaDelete(areaname)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
	} else {
		errs.Code = operator.DATA_ALL_OK
	}
	return
}

// 重命名区域（没有权限验证）
func (d *DRule) AreaRename(areaname string, newname string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()
	if areaname == INSIDE_DMZ {
		errs.Code = operator.DATA_USER_NO_AUTHORITY
		errs.Err = fmt.Errorf("cannot rename this area.")
		return
	}

	err = d.trule.AreaReName(areaname, newname)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
	} else {
		errs.Code = operator.DATA_ALL_OK
	}
	return
}

// 是否存在区域（没有权限验证）
func (d *DRule) AreaExist(areaname string) (exist bool) {
	exist = d.trule.AreaExist(areaname)
	return
}

// 区域列出（没有权限验证）
func (d *DRule) AreaList() (list []string, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	list, err = d.trule.AreaList()
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
	}
	return
}

// 用户区域列出（没有权限验证）
func (d *DRule) AreaUserList(username string) (rw map[string]bool, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	// 查看是否有这个用户
	userid := USER_PREFIX + username
	if have := d.trule.ExistRole(INSIDE_DMZ, userid); have == false {
		errs.Code = operator.DATA_USER_NO_EXIST
		errs.Err = fmt.Errorf("user not exist.")
		return
	}

	// 读取用户的权限表
	rw = make(map[string]bool)
	err = d.trule.ReadData(INSIDE_DMZ, userid, "WRable", &rw)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 用户增加区域（没有权限验证）
func (d *DRule) AreaAddUser(username, areaname string, add, wrable bool) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	if areaname == INSIDE_DMZ {
		errs.Code = operator.DATA_USER_NO_AUTHORITY
		errs.Err = fmt.Errorf("cannot operate this area.")
		return
	}

	// 查看是否有这个用户
	userid := USER_PREFIX + username
	if have := d.trule.ExistRole(INSIDE_DMZ, userid); have == false {
		errs.Code = operator.DATA_USER_NO_EXIST
		errs.Err = fmt.Errorf("user not exist.")
		return
	}
	// 查看是否有这个区域
	if have := d.trule.AreaExist(areaname); have == false {
		errs.Code = operator.DATA_AREA_NO_EXIST
		errs.Err = fmt.Errorf("area not exist.")
		return
	}
	// 查看用户是不是root级别
	var user_auth operator.UserAuthority
	err = d.trule.ReadData(INSIDE_DMZ, userid, "Authority", &user_auth)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	// 是root就直接返回，不需要加什么
	if user_auth == operator.USER_AUTHORITY_ROOT {
		errs.Code = operator.DATA_ALL_OK
		return
	}
	// 读取用户的权限表
	wrable_a := make(map[string]bool)
	err = d.trule.ReadData(INSIDE_DMZ, userid, "WRable", &wrable_a)
	if err != nil {
		fmt.Println("这里吗")
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	// 查看是删除还是添加
	if add == true {
		wrable_a[areaname] = wrable
	} else {
		delete(wrable_a, areaname)
	}
	// 重新写进去
	err = d.trule.WriteData(INSIDE_DMZ, userid, "WRable", wrable_a)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	// 看有没有正在登录的，有的话就改
	if _, find := d.loginuser[username]; find == true {
		d.loginuser[username].wrable = wrable_a
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 设置一个远程operator（没有权限验证）
func (d *DRule) OperatorSet(odo *operator.O_DRuleOperator) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	// id
	odo_id := OPERATOR_PREFIX + odo.Name
	odo_r := &DRuleOperator{
		Name:     odo.Name,
		Address:  odo.Address,
		ConnNum:  odo.ConnNum,
		TLS:      odo.TLS,
		Username: odo.Username,
		Password: odo.Password,
	}
	odo_r.New(odo_id)
	// 开始保存
	tran, _ := d.trule.Begin()
	if have := tran.ExistRole(INSIDE_DMZ, odo_id); have == false {
		err = tran.StoreRole(INSIDE_DMZ, odo_r)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
	} else {
		err = tran.WriteData(INSIDE_DMZ, odo_id, "Name", odo.Name)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.WriteData(INSIDE_DMZ, odo_id, "Address", odo.Address)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.WriteData(INSIDE_DMZ, odo_id, "ConnNum", odo.ConnNum)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.WriteData(INSIDE_DMZ, odo_id, "TLS", odo.TLS)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.WriteData(INSIDE_DMZ, odo_id, "Username", odo.Username)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		if len(odo.Password) != 0 {
			err = tran.WriteData(INSIDE_DMZ, odo_id, "Password", odo.Password)
			if err != nil {
				tran.Rollback()
				errs.Code = operator.DATA_RETURN_ERROR
				errs.Err = err
				return
			}
		}
	}

	have, err := tran.ExistChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	if have == false {
		err = tran.WriteChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
	}
	tran.Commit()
	errs.Code = operator.DATA_ALL_OK
	return
}

// 删除一个远程operator（没有权限验证）
func (d *DRule) OperatorDelete(oname string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	odo_id := OPERATOR_PREFIX + oname
	// 开始执行
	tran, _ := d.trule.Begin()
	have := tran.ExistRole(INSIDE_DMZ, odo_id)
	if have == true {
		err = tran.DeleteRole(INSIDE_DMZ, odo_id)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		err = tran.DeleteChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
	}
	tran.Commit()
	errs.Code = operator.DATA_ALL_OK
	return
}

// 列出全部远程operator（没有权限验证）
func (d *DRule) OperatorList() (list []operator.O_DRuleOperator, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	children, err := d.trule.ReadChildren(INSIDE_DMZ, OPERATOR_ROOT)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	list = make([]operator.O_DRuleOperator, 0)
	for _, child := range children {
		one := &DRuleOperator{}
		err = d.trule.ReadRole(INSIDE_DMZ, child, one)
		if err != nil {
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		onel := operator.O_DRuleOperator{
			Name:     one.Name,
			Address:  one.Address,
			ConnNum:  one.ConnNum,
			TLS:      one.TLS,
			Username: one.Username,
		}
		list = append(list, onel)
	}
	errs.Code = operator.DATA_ALL_OK
	return
}

// 是否存在这个operator
func (d *DRule) OperatorExist(o string) (exist bool, set operator.O_DRuleOperator, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	oname := OPERATOR_PREFIX + o
	exist = d.trule.ExistRole(INSIDE_DMZ, oname)

	if exist == true {
		one := &DRuleOperator{}
		err = d.trule.ReadRole(INSIDE_DMZ, oname, one)
		if err != nil {
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		set = operator.O_DRuleOperator{
			Name:     one.Name,
			Address:  one.Address,
			ConnNum:  one.ConnNum,
			TLS:      one.TLS,
			Username: one.Username,
		}
	}
	return
}

// 区域路由设置（没有权限验证）
func (d *DRule) AreaRouterSet(ars *operator.O_AreasRouter) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	ars_id := AREA_DRULE_PREFIX + ars.AreaName
	// 开始
	tran, _ := d.trule.Begin()
	ars_r := &AreasRouter{
		AreaName: ars.AreaName,
		Mirror:   ars.Mirror,
		Mirrors:  ars.Mirrors,
		Chars:    ars.Chars,
	}
	ars_r.New(ars_id)
	err = tran.StoreRole(INSIDE_DMZ, ars_r)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	have, err := tran.ExistChild(INSIDE_DMZ, AREA_DRULE_ROOT, ars_id)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	if have == false {
		err = tran.WriteChild(INSIDE_DMZ, AREA_DRULE_ROOT, ars_id)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
	}
	tran.Commit()
	d.areas[ars.AreaName] = ars_r
	errs.Code = operator.DATA_ALL_OK
	return
}

// 区域路由设置删除（没有权限验证）
func (d *DRule) AreaRouterDelete(areaname string) (errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	tran, _ := d.trule.Begin()
	err = tran.DeleteRole(INSIDE_DMZ, AREA_DRULE_PREFIX+areaname)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	err = tran.DeleteChild(INSIDE_DMZ, AREA_DRULE_ROOT, AREA_DRULE_PREFIX+areaname)
	if err != nil {
		tran.Rollback()
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	tran.Commit()
	delete(d.areas, areaname)
	errs.Code = operator.DATA_ALL_OK
	return
}

// 区域路由设置列出（没有权限验证）
func (d *DRule) AreaRouterList() (list []operator.O_AreasRouter, errs operator.DRuleError) {
	var err error
	errs = operator.NewDRuleError()

	tran, _ := d.trule.Begin()
	children, err := tran.ReadChildren(INSIDE_DMZ, AREA_DRULE_ROOT)
	if err != nil {
		errs.Code = operator.DATA_RETURN_ERROR
		errs.Err = err
		return
	}
	list = make([]operator.O_AreasRouter, 0)
	for _, child := range children {
		one_r := &AreasRouter{}
		err = tran.ReadRole(INSIDE_DMZ, child, one_r)
		if err != nil {
			tran.Rollback()
			errs.Code = operator.DATA_RETURN_ERROR
			errs.Err = err
			return
		}
		one := operator.O_AreasRouter{
			AreaName: one_r.AreaName,
			Mirror:   one_r.Mirror,
			Mirrors:  one_r.Mirrors,
			Chars:    one_r.Chars,
		}
		list = append(list, one)
	}
	tran.Commit()
	errs.Code = operator.DATA_ALL_OK
	return
}
