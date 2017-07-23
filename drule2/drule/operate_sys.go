// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"github.com/idcsource/insight00-lib/drule2/operator"
	"github.com/idcsource/insight00-lib/iendecode"
	"github.com/idcsource/insight00-lib/nst2"
)

// 启动
func (d *DRule) sys_druleStart(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 查看用户权限
	auth, login := d.getUserAuthority(o_send.User, o_send.Unid)
	if login == false {
		d.sendReceipt(conn_exec, operator.DATA_USER_NOT_LOGIN, "", nil)
		return
	}
	if auth != operator.USER_AUTHORITY_ROOT {
		d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 执行
	err = d.Start()
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	} else {
		errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
		return
	}
	return
}

// 暂停
func (d *DRule) sys_drulePause(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	// var err error
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
	// 执行
	d.Pause()
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 运行模式
func (d *DRule) sys_druleMode(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
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
	// 执行
	mode := d.dmode
	mode_b, err := iendecode.StructGobBytes(mode)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", mode_b)
	return
}
