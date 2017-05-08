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

	return
}
