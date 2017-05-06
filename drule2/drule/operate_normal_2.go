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

// 存在角色吗
func (d *DRule) normalExitRole(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleSendAndReceive{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.RoleID, false)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}

	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			ds.IfHave = d.trule.ExistRole(ds.Area, ds.RoleID)
		} else {
			ds.IfHave = tran.tran.ExistRole(ds.Area, ds.RoleID)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.IfHave, errd = o.ExistRole(ds.Area, ds.RoleID)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.IfHave, errd = o.ExistRole(ds.Area, ds.RoleID)
		}
		if errd.IsError() != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, errd.String(), nil)
			return
		}
	}
	// 编码
	ds_b, err := iendecode.StructGobBytes(ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 发送
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", ds_b)
	return
}

// 读角色
func (d *DRule) normalReadRole(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleSendAndReceive{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.RoleID, false)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			ds.RoleBody, err = d.trule.ReadRoleMiddleData(ds.Area, ds.RoleID)
		} else {
			ds.RoleBody, err = tran.tran.ReadRoleMiddleData(ds.Area, ds.RoleID)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.RoleBody, errd = o.ReadRoleToMiddleData(ds.Area, ds.RoleID)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.RoleBody, errd = o.ReadRoleToMiddleData(ds.Area, ds.RoleID)
		}
		if errd.IsError() != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, errd.String(), nil)
			return
		}
	}
	// 编码
	ds_b, err := iendecode.StructGobBytes(ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 发送
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", ds_b)
	return
}
