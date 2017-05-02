// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"time"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 用户登录
func (d *DRule) man_userLogin(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 解码
	var login operator.O_DRuleUser
	err = iendecode.BytesGobStruct(o_send.Data, &login)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	user_id := USER_PREFIX + login.UserName
	// 查看有没有这个用户
	user_have := d.trule.ExistRole(INSIDE_DMZ, user_id)
	if user_have == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_EXIST, "", nil)
		return
	}
	// 查看密码
	var password string
	err = d.trule.ReadData(INSIDE_DMZ, user_id, "Password", &password)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	if password != login.Password {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_EXIST, "", nil)
		return
	}
	// 规整权限
	var auth operator.UserAuthority
	d.trule.ReadData(INSIDE_DMZ, user_id, "Authority", &auth)
	wrable := make(map[string]bool)
	d.trule.ReadData(INSIDE_DMZ, user_id, "WRable", &wrable)

	unid := random.Unid(1, time.Now().String(), login.UserName)

	// 查看是否已经有登录的了，并写入登录的乱七八糟
	_, find := d.loginuser[login.UserName]
	if find {
		d.loginuser[login.UserName].wrable = wrable
		d.loginuser[login.UserName].unid[unid] = time.Now()
	} else {
		loginuser := &loginUser{
			username:  login.UserName,
			unid:      make(map[string]time.Time),
			authority: auth,
			wrable:    wrable,
		}
		loginuser.unid[unid] = time.Now()
		d.loginuser[login.UserName] = loginuser
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 用户续命
func (d *DRule) man_userAddLife(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	yes := d.checkUserLogin(o_send.User, o_send.Unid)
	if yes == true {
		errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
		return
	} else {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
}

// 新建用户
func (d *DRule) man_userAdd(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var newuser operator.O_DRuleUser
	err = iendecode.BytesGobStruct(o_send.Data, &newuser)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	user_id := USER_PREFIX + newuser.UserName

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
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = tran.WriteChild(INSIDE_DMZ, USER_PREFIX+ROOT_USER, user_id)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = tran.Commit()
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 修改密码
func (d *DRule) man_userPassword(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	// 解码
	var theuser operator.O_DRuleUser
	err = iendecode.BytesGobStruct(o_send.Data, &theuser)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	if auth != operator.USER_AUTHORITY_ROOT || theuser.UserName != o_send.User {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 修改密码
	userid := USER_PREFIX + theuser.UserName
	err = d.trule.WriteData(INSIDE_DMZ, userid, "Password", theuser.Password)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 修改邮箱
func (d *DRule) man_userEmail(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	// 解码
	var theuser operator.O_DRuleUser
	err = iendecode.BytesGobStruct(o_send.Data, &theuser)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	if auth != operator.USER_AUTHORITY_ROOT || theuser.UserName != o_send.User {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 修改密码
	userid := USER_PREFIX + theuser.UserName
	err = d.trule.WriteData(INSIDE_DMZ, userid, "Email", theuser.Email)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除用户
func (d *DRule) man_userDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var username string
	err = iendecode.BytesGobStruct(o_send.Data, &username)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 检查是否为可以删除的
	if username == ROOT_USER || username == o_send.User {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "Can not delete this user.", nil)
		return
	}
	userid := USER_PREFIX + username

	tran, _ := d.trule.Begin()
	err = tran.DeleteRole(INSIDE_DMZ, userid)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		tran.Rollback()
		return
	}
	err = tran.DeleteChild(INSIDE_DMZ, USER_PREFIX+ROOT_USER, userid)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		tran.Rollback()
		return
	}
	err = tran.Commit()
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 用户登出
func (d *DRule) man_userLogout(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	// 查看用户权限
	_, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
	}
	// 删除相应登陆的信息
	delete(d.loginuser[o_send.User].unid, o_send.Unid)
	if len(d.loginuser[o_send.User].unid) == 0 {
		delete(d.loginuser, o_send.User)
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 用户列表
func (d *DRule) man_userList(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	list := make([]operator.O_DRuleUser, 0)

	tran, _ := d.trule.Begin()
	children, err := tran.ReadChildren(INSIDE_DMZ, USER_PREFIX+ROOT_USER)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		tran.Rollback()
		return
	}
	for _, child := range children {
		one := operator.O_DRuleUser{}
		err = tran.ReadData(INSIDE_DMZ, child, "UserName", &one.UserName)
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			tran.Rollback()
			return
		}
		err = tran.ReadData(INSIDE_DMZ, child, "Email", &one.Email)
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			tran.Rollback()
			return
		}
		err = tran.ReadData(INSIDE_DMZ, child, "Authority", &one.Authority)
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			tran.Rollback()
			return
		}
		list = append(list, one)
	}
	tran.Commit()
	// 编码
	list_b, err := iendecode.StructGobBytes(list)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 发送
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", list_b)
	return
}

// 新建区域
func (d *DRule) man_areaAdd(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var area operator.O_Area
	err = iendecode.BytesGobStruct(o_send.Data, &area)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = d.trule.AreaInit(area.AreaName)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除区域
func (d *DRule) man_areaDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var area operator.O_Area
	err = iendecode.BytesGobStruct(o_send.Data, &area)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = d.trule.AreaDelete(area.AreaName)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 区域改名
func (d *DRule) man_areaRename(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var area operator.O_Area
	err = iendecode.BytesGobStruct(o_send.Data, &area)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = d.trule.AreaReName(area.AreaName, area.Rename)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 区域是否存在
func (d *DRule) man_areaExist(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	var area operator.O_Area
	err = iendecode.BytesGobStruct(o_send.Data, &area)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	area.Exist = d.trule.AreaExist(area.AreaName)
	// 编码
	area_b, err := iendecode.StructGobBytes(area)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", area_b)
	return
}

// 所有区域的列表
func (d *DRule) man_areaList(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	list, err := d.trule.AreaList()
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 编码
	list_b, err := iendecode.StructGobBytes(list)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", list_b)
	return
}

// 用户和区域的关系
func (d *DRule) man_areaAndUser(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	au := operator.O_Area_User{}
	err = iendecode.BytesGobStruct(o_send.Data, &au)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看是否有这个用户
	userid := USER_PREFIX + au.UserName
	if have := d.trule.ExistRole(INSIDE_DMZ, userid); have == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_EXIST, "user not exist.", nil)
		return
	}
	// 查看是否有这个区域
	if have := d.trule.AreaExist(au.Area); have == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_AREA_NO_EXIST, "area not exist.", nil)
		return
	}
	// 查看用户是不是root级别
	var user_auth operator.UserAuthority
	err = d.trule.ReadData(INSIDE_DMZ, userid, "Authority", &user_auth)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 是root就直接返回，不需要加什么
	if user_auth == operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
		return
	}
	// 读取用户的权限表
	wrable := make(map[string]bool)
	err = d.trule.ReadData(INSIDE_DMZ, userid, "WRable", &wrable)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看是删除还是添加
	if au.Add == true {
		wrable[au.Area] = au.WRable
	} else {
		delete(wrable, au.Area)
	}
	// 重新写进去
	err = d.trule.WriteData(INSIDE_DMZ, userid, "WRable", wrable)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 看有没有正在登录的，有的话就改
	if _, find := d.loginuser[au.UserName]; find == true {
		d.loginuser[au.UserName].wrable = wrable
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 设置operator
func (d *DRule) man_operatorSet(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 解码
	odo := operator.O_DRuleOperator{}
	err = iendecode.BytesGobStruct(o_send.Data, &odo)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
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
	err = tran.StoreRole(INSIDE_DMZ, odo_r)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	have, err := tran.ExistChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	if have == false {
		err = tran.WriteChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
		if err != nil {
			tran.Rollback()
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	}
	tran.Commit()
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除operator
func (d *DRule) man_operatorDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	name := string(o_send.Data)
	odo_id := OPERATOR_PREFIX + name
	// 开始执行
	tran, _ := d.trule.Begin()
	have := tran.ExistRole(INSIDE_DMZ, odo_id)
	if have == true {
		err = tran.DeleteRole(INSIDE_DMZ, odo_id)
		if err != nil {
			tran.Rollback()
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
		err = tran.DeleteChild(INSIDE_DMZ, OPERATOR_ROOT, odo_id)
		if err != nil {
			tran.Rollback()
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	}
	tran.Commit()
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 列出operator
func (d *DRule) man_operatorList(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	children, err := d.trule.ReadChildren(INSIDE_DMZ, OPERATOR_ROOT)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	list := make([]operator.O_DRuleOperator, 0)
	for _, child := range children {
		one := &DRuleOperator{}
		err = d.trule.ReadRole(INSIDE_DMZ, child, one)
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
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
	// 编码
	list_b, err := iendecode.StructGobBytes(list)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", list_b)
	return
}

// 远端路由的设置
func (d *DRule) man_areaRouterSet(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	// 解码
	ars := operator.O_AreasRouter{}
	err = iendecode.BytesGobStruct(o_send.Data, &ars)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
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
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	have, err := tran.ExistChild(INSIDE_DMZ, AREA_DRULE_ROOT, ars_id)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	if have == false {
		err = tran.WriteChild(INSIDE_DMZ, AREA_DRULE_ROOT, ars_id)
		if err != nil {
			tran.Rollback()
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	}
	tran.Commit()
	d.areas[ars.AreaName] = ars_r
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除远端路由的设置
func (d *DRule) man_areaRouterDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	areaname := string(o_send.Data)
	tran, _ := d.trule.Begin()
	err = tran.DeleteRole(INSIDE_DMZ, AREA_DRULE_PREFIX+areaname)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	err = tran.DeleteChild(INSIDE_DMZ, AREA_DRULE_ROOT, AREA_DRULE_PREFIX+areaname)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	tran.Commit()
	delete(d.areas, areaname)
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 列出远端路由的设置
func (d *DRule) man_areaRouterList(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}

	tran, _ := d.trule.Begin()
	children, err := tran.ReadChildren(INSIDE_DMZ, AREA_DRULE_ROOT)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	list := make([]operator.O_AreasRouter, 0)
	for _, child := range children {
		one_r := &AreasRouter{}
		err = tran.ReadRole(INSIDE_DMZ, child, one_r)
		if err != nil {
			tran.Rollback()
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
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
	// 编码
	list_b, err := iendecode.StructGobBytes(list)
	if err != nil {
		tran.Rollback()
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", list_b)
	return
}
