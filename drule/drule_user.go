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

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

const (
	USER_ROOT_USER_NAME             = "root"   // 根用户用户名
	USER_ROOT_USER_DEFAULT_PASSWORD = "123456" // 根用户默认密码
	USER_ALIVE_TIME                 = 3600     // 用户的登录生存期
)

const (
	USER_AUTHORITY_NO     = iota // 没有权限
	USER_AUTHORITY_ROOT          // 根权限
	USER_AUTHORITY_NORMAL        // 普通权限
)

// Drule和Operator的用户
type DRuleUser struct {
	roles.Role        // 角色
	UserName   string // 用户名
	Password   string // 密码
	Email      string // 邮箱
	Authority  uint8  // 权限，USER_AUTHORITY_*
}

// 来往网络的用户信息
type Net_DRuleUser struct {
	UserName  string // 用户名
	Password  string // 密码
	Email     string // 邮箱
	Authority uint8  // 权限，USER_AUTHORITY_*
	Unid      string // 唯一码
}

// 登录进DRule中
func LoginToDRule(process *nst.ProgressData, selfname string, slaveIn *slaveIn) (err error) {
	tolog := Net_PrefixStat{
		ClientName: selfname,
		Operate:    OPERATE_USER_LOGIN,
	}
	tolog_b, err := iendecode.StructGobBytes(tolog)
	if err != nil {
		return
	}
	rdata, err := process.SendAndReturn(tolog_b)
	if err != nil {
		return
	}
	receipt := Net_SlaveReceipt_Data{}
	err = iendecode.BytesGobStruct(rdata, &receipt)
	if receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(receipt.Error)
	}
	// 构建login信息
	login := Net_DRuleUser{
		UserName: slaveIn.username,
		Password: slaveIn.password,
	}
	login_b, err := iendecode.StructGobBytes(login)
	if err != nil {
		return
	}
	rdata, err = process.SendAndReturn(login_b)
	err = iendecode.BytesGobStruct(rdata, &receipt)
	if receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(receipt.Error)
	}
	err = iendecode.BytesGobStruct(receipt.Data, &login)
	if err != nil {
		return
	}
	if login.Unid != "" && login.Unid != "0" {
		slaveIn.unid = login.Unid
		return nil
	} else {
		err = fmt.Errorf("The login is wrong.")
		return err
	}
	return
}

// 清理登录超时
func (d *DRule) userLoginTimeOutDel() {
	for {
		time.Sleep(USER_ALIVE_TIME * time.Second)
		for key, log := range d.loginuser {
			if log.logtime.Unix()+USER_ALIVE_TIME > time.Now().Unix() {
				delete(d.loginuser, key)
			}
		}
	}
}

// 用户登录
func (d *DRule) userLogin(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 获取编码的结构体数据
	druleuser_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 还原
	druleuser := Net_DRuleUser{}
	err = iendecode.BytesGobStruct(druleuser_b, &druleuser)
	if err != nil {
		return err
	}
	// 找有无这个用户以及用户名是否正确
	var password string
	err = d.trule.ReadData(d.selfname+"_"+druleuser.UserName, "Password", &password)
	if err != nil {
		druleuser.Unid = "0"
	} else {
		if password != druleuser.Password {
			druleuser.Unid = "0"
		} else {
			var authority uint8
			err = d.trule.ReadData(d.selfname+"_"+druleuser.UserName, "Authority", &authority)
			if err != nil {
				return err
			}
			druleuser.Unid = random.Unid(1, druleuser.UserName)
			// 写入登录表
			login := &loginUser{
				username:  druleuser.UserName,
				unid:      druleuser.Unid,
				authority: authority,
				logtime:   time.Now(),
			}
			d.loginuser[druleuser.Unid] = login
		}
	}
	// 编码发送结果
	druleuser_b, err = iendecode.StructGobBytes(druleuser)
	if err != nil {
		return
	}
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, druleuser_b, nil)
	return
}

// 添加用户
func (d *DRule) userAdd(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 检查是否为超级管理员
	if d.loginuser[prefix_stat.Unid].authority != USER_AUTHORITY_ROOT {
		err = fmt.Errorf("The user not a root.")
		return
	}
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 获取编码的结构体数据
	druleuser_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 还原
	druleuser := Net_DRuleUser{}
	err = iendecode.BytesGobStruct(druleuser_b, &druleuser)
	if err != nil {
		return err
	}
	userid := d.selfname + "_" + druleuser.UserName
	// 查看是否有重复的
	have := d.trule.ExistRole(userid)
	if have == true {
		err = d.serverDataReceipt(conn_exec, DATA_USER_EXIST, nil, nil)
		return
	}
	// 创建新用户
	newuser := &DRuleUser{
		UserName:  druleuser.UserName,
		Password:  druleuser.Password,
		Email:     druleuser.Email,
		Authority: druleuser.Authority,
	}
	newuser.New(userid)
	err = d.trule.StoreRole(newuser)
	if err != nil {
		err = d.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, err)
		return
	}
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, nil, nil)
	return
}
