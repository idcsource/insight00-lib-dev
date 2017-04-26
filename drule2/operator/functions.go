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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_USER_PASSWORD, user_b)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_USER_ADD, user_b)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_USER_DEL, []byte(username))
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

// 列出所有用户
func (o *Operator) UserList() (list []O_DRuleUser, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_USER_DEL, nil)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_AREA_ADD, a_b)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_AREA_DEL, a_b)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_AREA_DEL, a_b)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_AREA_LIST, nil)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]AreaList: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]AreaList: %v", drule_r.Error)
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
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_TRAN_BEGIN, tsend_b)
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
