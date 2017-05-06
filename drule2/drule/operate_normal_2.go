// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"strings"

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

// 写角色
func (d *DRule) normalStoreRole(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleSendAndReceive{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.RoleID, true)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			err = d.trule.StoreRoleFromMiddleData(ds.Area, ds.RoleBody)
		} else {
			err = tran.tran.StoreRoleFromMiddleData(ds.Area, ds.RoleBody)
		}
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	} else {
		var errd operator.DRuleError
		erra := make([]string, 0)
		if tran == nil {
			for _, one := range o_s {
				if _, find := d.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = d.operators[one].StoreRoleFromMiddleData(ds.Area, ds.RoleBody)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		} else {
			for _, one := range o_s {
				if _, find := tran.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = tran.operators[one].StoreRoleFromMiddleData(ds.Area, ds.RoleBody)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		}
		if len(erra) != 0 {
			errstr := strings.Join(erra, " | ")
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, errstr, nil)
			return
		}
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 删除角色
func (d *DRule) normalDeleteRole(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleSendAndReceive{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.RoleID, true)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			err = d.trule.DeleteRole(ds.Area, ds.RoleID)
		} else {
			err = tran.tran.DeleteRole(ds.Area, ds.RoleID)
		}
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	} else {
		var errd operator.DRuleError
		erra := make([]string, 0)
		if tran == nil {
			for _, one := range o_s {
				if _, find := d.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = d.operators[one].DeleteRole(ds.Area, ds.RoleID)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		} else {
			for _, one := range o_s {
				if _, find := tran.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = tran.operators[one].DeleteRole(ds.Area, ds.RoleID)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		}
		if len(erra) != 0 {
			errstr := strings.Join(erra, " | ")
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, errstr, nil)
			return
		}
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 写father
func (d *DRule) normalWriteFather(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleFatherChange{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.Id, true)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			err = d.trule.WriteFather(ds.Area, ds.Id, ds.Father)
		} else {
			err = tran.tran.WriteFather(ds.Area, ds.Id, ds.Father)
		}
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	} else {
		var errd operator.DRuleError
		erra := make([]string, 0)
		if tran == nil {
			for _, one := range o_s {
				if _, find := d.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = d.operators[one].WriteFather(ds.Area, ds.Id, ds.Father)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		} else {
			for _, one := range o_s {
				if _, find := tran.operators[one]; find == false {
					erra = append(erra, "Can not find remote operator "+one)
					continue
				}
				errd = tran.operators[one].WriteFather(ds.Area, ds.Id, ds.Father)
				if errd.IsError() != nil {
					erra = append(erra, errd.String())
				}
			}
		}
		if len(erra) != 0 {
			errstr := strings.Join(erra, " | ")
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, errstr, nil)
			return
		}
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 读father
func (d *DRule) normalReadFather(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleFatherChange{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.Id, false)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			ds.Father, err = d.trule.ReadFather(ds.Area, ds.Id)
		} else {
			ds.Father, err = tran.tran.ReadFather(ds.Area, ds.Id)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Father, errd = o.ReadFather(ds.Area, ds.Id)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Father, errd = o.ReadFather(ds.Area, ds.Id)
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

// 读children
func (d *DRule) normalReadChildren(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndChildren{}
	err = iendecode.BytesGobStruct(o_send.Data, &ds)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看读写权限和角色位置，o_s = operator service
	havepower, position, o_s := d.getAreaPowerAndRolePosition(o_send.User, ds.Area, ds.Id, false)
	if havepower == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AREA_AUTHORITY, "", nil)
		return
	}
	if position == ROLE_POSITION_IN_LOCAL {
		if tran == nil {
			ds.Children, err = d.trule.ReadChildren(ds.Area, ds.Id)
		} else {
			ds.Children, err = tran.tran.ReadChildren(ds.Area, ds.Id)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Children, errd = o.ReadChildren(ds.Area, ds.Id)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Children, errd = o.ReadChildren(ds.Area, ds.Id)
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
