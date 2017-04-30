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
