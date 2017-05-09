// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst"
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

	unid, _, errd := d.UserLogin(login.UserName, login.Password)
	if errd.IsError() != nil {
		errs = d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}

	login.Unid = unid
	// 编码
	login_b, err := iendecode.StructGobBytes(login)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", login_b)
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

	errd := d.UserAdd(&newuser)

	errs = d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
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
	// 修改邮箱
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
	username := string(o_send.Data)
	// 检查是否为可以删除的
	if username == ROOT_USER || username == o_send.User {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "Can not delete this user.", nil)
		return
	}
	errd := d.UserDelete(username)
	if errd.IsError() != nil {
		errs = d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
	} else {
		errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	}
	return
}

// 用户登出
func (d *DRule) man_userLogout(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	// 查看用户权限
	_, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
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

	list, errd := d.UserList()
	if errd.IsError() != nil {
		errs = d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}
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
	errd := d.AreaAddUser(au.UserName, au.Area, au.Add, au.WRable)
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 设置operator
func (d *DRule) man_operatorSet(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	if d.closed != true {
		errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_NOT_PAUSED, "The DRule must be paused to do this.", nil)
		return
	}

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
	errd := d.OperatorSet(&odo)
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除operator
func (d *DRule) man_operatorDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	if d.closed != true {
		errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_NOT_PAUSED, "The DRule must be paused to do this.", nil)
		return
	}

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
	errd := d.OperatorDelete(name)
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}
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

	list, errd := d.OperatorList()
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
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

// 远端路由的设置
func (d *DRule) man_areaRouterSet(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	if d.closed != true {
		errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_NOT_PAUSED, "The DRule must be paused to do this.", nil)
		return
	}

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

	errd := d.AreaRouterSet(&ars)
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}

	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除远端路由的设置
func (d *DRule) man_areaRouterDel(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	if d.closed != true {
		errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_NOT_PAUSED, "The DRule must be paused to do this.", nil)
		return
	}

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

	errd := d.AreaRouterDelete(areaname)
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
		return
	}

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

	list, errd := d.AreaRouterList()
	if errd.IsError() != nil {
		d.sendReceipt(conn_exec, errd.Code, errd.String(), nil)
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
