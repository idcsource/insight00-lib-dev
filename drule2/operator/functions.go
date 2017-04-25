// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"fmt"
)

import (
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 设置或修改用户的密码
func (o *Operator) Password(username, password string) (err error) {
	user := O_DRuleUser{
		UserName: username,
		Password: random.GetSha1Sum(password),
	}
	user_b, err := iendecode.StructGobBytes(user)
	if err != nil {
		err = fmt.Errorf("operator[Operator]Password: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, o.drule.unid, o.selfname, "", false, "", OPERATE_USER_PASSWORD, user_b)
	if err != nil {
		err = fmt.Errorf("operator[Operator]Password: %v", err)
		return
	}
	if drule_r.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(drule_r.Error)
	}
	return
}

// 增加用户
func (o *Operator) UserAdd(username, password, email string, authority UserAuthority) (err error) {
	user := O_DRuleUser{
		UserName:  username,
		Password:  random.GetSha1Sum(password),
		Email:     email,
		Authority: authority,
	}
	user_b, err := iendecode.StructGobBytes(user)
	if err != nil {
		err = fmt.Errorf("operator[Operator]UserAdd: %v", err)
		return
	}
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, o.drule.unid, o.selfname, "", false, "", OPERATE_USER_ADD, user_b)
	if err != nil {
		err = fmt.Errorf("operator[Operator]UserAdd: %v", err)
		return
	}
	if drule_r.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(drule_r.Error)
	}
	return
}

// 删除用户
func (o *Operator) UserDel(username string) (err error) {
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, o.drule.unid, o.selfname, "", false, "", OPERATE_USER_DEL, []byte(username))
	if err != nil {
		err = fmt.Errorf("operator[Operator]UserDel: %v", err)
		return
	}
	if drule_r.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(drule_r.Error)
	}
	return
}

// 列出所有用户
func (o *Operator) UserList() (list []O_DRuleUser, err error) {
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, o.drule.unid, o.selfname, "", false, "", OPERATE_USER_DEL, nil)
	if err != nil {
		err = fmt.Errorf("operator[Operator]UserList: %v", err)
		return
	}
	if drule_r.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(drule_r.Error)
	}
	// 解码
	err = iendecode.BytesGobStruct(drule_r.Data, &list)
	if err != nil {
		err = fmt.Errorf("operator[Operator]UserList: %v", err)
		return
	}
	return
}
