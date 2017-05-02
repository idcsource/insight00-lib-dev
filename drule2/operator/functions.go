// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"fmt"
	"time"
)

import (
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 设置或修改用户的密码
func (o *Operator) Password(username, password string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	user := O_DRuleUser{
		UserName: username,
		Password: random.GetSha1Sum(password),
	}
	user_b, err := iendecode.StructGobBytes(user)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]Password: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_PASSWORD, user_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]Password: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]Password: %v", drule_r.Error)
	}
	return
}

// 增加用户
func (o *Operator) UserAdd(username, password, email string, authority UserAuthority) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	user := O_DRuleUser{
		UserName:  username,
		Password:  random.GetSha1Sum(password),
		Email:     email,
		Authority: authority,
	}
	user_b, err := iendecode.StructGobBytes(user)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAdd: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_ADD, user_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAdd: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserAdd: %v", drule_r.Error)
	}
	return
}

// 删除用户
func (o *Operator) UserDel(username string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_DEL, []byte(username))
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserDel: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserDel: %v", drule_r.Error)
	}
	return
}

// 用户登陆
func (o *Operator) UserLogin(username, password string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	o.drule.username = username
	o.drule.password = password
	login := O_DRuleUser{
		UserName: username,
		Password: random.GetSha1Sum(password),
	}
	// 编码
	login_b, err := iendecode.StructGobBytes(login)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserLogin: %v", err)
		return
	}
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_LOGIN, login_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserLogin: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserLogin: %v", drule_r.Error)
	}
	// 解码
	err = iendecode.BytesGobStruct(login_b, &login)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserLogin: %v", err)
		return
	}
	o.login = true
	o.drule.unid = login.Unid

	return
}

// 用户登出
func (o *Operator) UserLogout() (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_LOGOUT, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserLogout: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserLogout: %v", drule_r.Error)
		return
	}
	o.login = false
	return
}

// 列出所有用户
func (o *Operator) UserList() (list []O_DRuleUser, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_DEL, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserList: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf(drule_r.Error)
	}
	// 解码
	err = iendecode.BytesGobStruct(drule_r.Data, &list)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserList: %v", err)
		return
	}
	return
}

// 新建区域
func (o *Operator) AreaAdd(area string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	a := O_Area{
		AreaName: area,
	}
	a_b, err := iendecode.StructGobBytes(a)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaAdd: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_ADD, a_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaAdd: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaAdd: %v", drule_r.Error)
	}
	return
}

// 删除区域
func (o *Operator) AreaDel(area string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	a := O_Area{
		AreaName: area,
	}
	a_b, err := iendecode.StructGobBytes(a)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaDel: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_DEL, a_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaDel: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaDel: %v", drule_r.Error)
	}
	return
}

// 区域是否存在
func (o *Operator) AreaExist(area string) (exist bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	a := O_Area{
		AreaName: area,
	}
	a_b, err := iendecode.StructGobBytes(a)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaExist: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_DEL, a_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaExist: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaExist: %v", drule_r.Error)
	}
	// 解码返回
	err = iendecode.BytesGobStruct(drule_r.Data, &a)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaExist: %v", err)
		return
	}
	exist = a.Exist
	return
}

// 重命名区域
func (o *Operator) AreaRename(oldname, newname string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	a := O_Area{
		AreaName: oldname,
		Rename:   newname,
	}
	a_b, err := iendecode.StructGobBytes(a)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRename: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_DEL, a_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRename: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaRename: %v", drule_r.Error)
	}
	return
}

// 区域列表
func (o *Operator) AreaList() (list []string, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_LIST, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaList: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaList: %v", drule_r.Error)
	}
	list = make([]string, 0)
	// 解码
	err = iendecode.BytesGobStruct(drule_r.Data, &list)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaList: %v", err)
	}
	return
}

// 添加(更新)一个用户对区域的访问权限
func (o *Operator) UserAreaAlert(userid, areaname string, rw UserAreaVisit) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 构建
	au := O_Area_User{
		UserName: userid,
		Area:     areaname,
		Add:      true,
	}
	if rw == USER_AREA_VISIT_READONLY {
		au.WRable = false
	} else {
		au.WRable = true
	}
	au_b, err := iendecode.StructGobBytes(au)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaAlert: %v", err)
		return
	}
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_AREA, au_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaAlert: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaAlert: %v", drule_r.Error)
	}
	return
}

// 删除一个用户对区域的访问权限
func (o *Operator) UserAreaDelete(userid, areaname string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 构建
	au := O_Area_User{
		UserName: userid,
		Area:     areaname,
		Add:      false,
	}
	au_b, err := iendecode.StructGobBytes(au)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaDelete: %v", err)
		return
	}
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_USER_AREA, au_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaDelete: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]UserAreaDelete: %v", drule_r.Error)
	}
	return
}

// 给DRule添加一个Operator
func (o *Operator) DRuleOperatorSet(name, address string, connnum int, tls bool, username, password string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 构建
	do := O_DRuleOperator{
		Name:     name,
		Address:  address,
		ConnNum:  connnum,
		TLS:      tls,
		Username: username,
		Password: random.GetSha1Sum(password),
	}
	do_b, err := iendecode.StructGobBytes(do)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorSet: %v", err)
		return
	}
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_DRULE_OPERATOR_SET, do_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorSet: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorSet: %v", drule_r.Error)
	}
	return
}

// 删除DRule的一个远程Operator
func (o *Operator) DRuleOperatorDelete(name string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_DRULE_OPERATOR_DELETE, []byte(name))
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorDelete: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorDelete: %v", drule_r.Error)
	}
	return
}

// 返回DRule上远端Operator的列表
func (o *Operator) DRuleOperatorList() (list []O_DRuleOperator, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_DRULE_OPERATOR_LIST, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorList: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorList: %v", drule_r.Error)
	}
	// 解码返回
	list = make([]O_DRuleOperator, 0)
	err = iendecode.BytesGobStruct(drule_r.Data, &list)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleOperatorList: %v", err)
		return
	}
	return
}

// 设置远端区域的路由
func (o *Operator) AreaRouterSet(set O_AreasRouter) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 编码
	set_b, err := iendecode.StructGobBytes(set)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterAdd: %v", err)
		return
	}
	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_ROUTER_SET, set_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterAdd: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterAdd: %v", drule_r.Error)
	}
	return
}

// 删除远端区域的路由
func (o *Operator) AreaRouterDelete(areaname string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_ROUTER_DELETE, []byte(areaname))
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterDelete: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterDelete: %v", drule_r.Error)
	}
	return
}

// 返回所有区域路由的信息
func (o *Operator) AreaRouterList() (list []O_AreasRouter, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_MANAGE, OPERATE_AREA_ROUTER_LIST, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterList: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterList: %v", drule_r.Error)
	}
	// 解码返回
	list = make([]O_AreasRouter, 0)
	err = iendecode.BytesGobStruct(drule_r.Data, &list)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaRouterList: %v", err)
		return
	}
	return
}

// 返回Drule的工作模式
func (o *Operator) DRuleModeGet() (mode DRuleOperateMode, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_SYSTEM, OPERATE_DRULE_OPERATE_MODE, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleModeGet: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRuleModeGet: %v", drule_r.Error)
	}
	// 解码返回
	err = iendecode.BytesGobStruct(drule_r.Data, &mode)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleModeGet: %v", err)
		return
	}
	return
}

// 启动DRule
func (o *Operator) DRuleStart() (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_SYSTEM, OPERATE_DRULE_START, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRuleStart: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRuleStart: %v", drule_r.Error)
	}
	return
}

// 暂停Drule
func (o *Operator) DRulePause() (errs DRuleError) {
	var err error
	errs = NewDRuleError()

	// 发送
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_SYSTEM, OPERATE_DRULE_PAUSE, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DRulePause: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DRulePause: %v", drule_r.Error)
	}
	return
}

// 创建事务
func (o *Operator) Begin() (t *OTransaction, errs DRuleError) {
	var err error
	errs = NewDRuleError()

	unid := random.Unid(1, o.selfname, time.Now().String())

	tsend := O_Transaction{
		TransactionId: unid,
	}
	tsend_b, err := iendecode.StructGobBytes(tsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]Begin: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_NORMAL, OPERATE_TRAN_BEGIN, tsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]Begin: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]Begin: %v", drule_r.Error)
		return
	}

	t = &OTransaction{
		selfname:       o.selfname,
		transaction_id: unid,
		drule:          o.drule,
		service:        o.service,
		logs:           o.logs,
		bedelete:       false,
		activetime:     time.Now(),
	}
	o.transaction[unid] = t
	return
}
