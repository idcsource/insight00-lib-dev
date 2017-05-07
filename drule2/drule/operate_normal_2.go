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
	"github.com/idcsource/Insight-0-0-lib/roles"
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

// 写children
func (d *DRule) normalWriteChildren(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndChildren{}
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
			err = d.trule.WriteChildren(ds.Area, ds.Id, ds.Children)
		} else {
			err = tran.tran.WriteChildren(ds.Area, ds.Id, ds.Children)
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
				errd = d.operators[one].WriteChildren(ds.Area, ds.Id, ds.Children)
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
				errd = tran.operators[one].WriteChildren(ds.Area, ds.Id, ds.Children)
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

// 写child
func (d *DRule) normalWriteChild(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndChild{}
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
			err = d.trule.WriteChild(ds.Area, ds.Id, ds.Child)
		} else {
			err = tran.tran.WriteChild(ds.Area, ds.Id, ds.Child)
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
				errd = d.operators[one].WriteChild(ds.Area, ds.Id, ds.Child)
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
				errd = tran.operators[one].WriteChild(ds.Area, ds.Id, ds.Child)
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

// 删child
func (d *DRule) normalDeleteChild(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndChild{}
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
			err = d.trule.DeleteChild(ds.Area, ds.Id, ds.Child)
		} else {
			err = tran.tran.DeleteChild(ds.Area, ds.Id, ds.Child)
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
				errd = d.operators[one].DeleteChild(ds.Area, ds.Id, ds.Child)
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
				errd = tran.operators[one].DeleteChild(ds.Area, ds.Id, ds.Child)
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

// 有child吗
func (d *DRule) normalExistChild(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndChild{}
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
			ds.Exist, err = d.trule.ExistChild(ds.Area, ds.Id, ds.Child)
		} else {
			ds.Exist, err = tran.tran.ExistChild(ds.Area, ds.Id, ds.Child)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Exist, errd = o.ExistChild(ds.Area, ds.Id, ds.Child)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Exist, errd = o.ExistChild(ds.Area, ds.Id, ds.Child)
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

// 读朋友们
func (d *DRule) normalReadFriends(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndFriends{}
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
			ds.Friends, err = d.trule.ReadFriends(ds.Area, ds.Id)
		} else {
			ds.Friends, err = tran.tran.ReadFriends(ds.Area, ds.Id)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Friends, errd = o.ReadFriends(ds.Area, ds.Id)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Friends, errd = o.ReadFriends(ds.Area, ds.Id)
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

// 写friends
func (d *DRule) normalWriteFriends(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndFriends{}
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
			err = d.trule.WriteFriends(ds.Area, ds.Id, ds.Friends)
		} else {
			err = tran.tran.WriteFriends(ds.Area, ds.Id, ds.Friends)
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
				errd = d.operators[one].WriteFriends(ds.Area, ds.Id, ds.Friends)
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
				errd = tran.operators[one].WriteFriends(ds.Area, ds.Id, ds.Friends)
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

// 写friend状态
func (d *DRule) normalWriteFriendStatus(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndFriend{}
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
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				err = d.trule.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				err = d.trule.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				err = d.trule.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Complex)
			}
		} else {
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				err = tran.tran.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				err = tran.tran.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				err = tran.tran.WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Complex)
			}
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
				switch ds.Single {
				case roles.STATUS_VALUE_TYPE_INT:
					errd = d.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Int)
				case roles.STATUS_VALUE_TYPE_FLOAT:
					errd = d.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Float)
				case roles.STATUS_VALUE_TYPE_COMPLEX:
					errd = d.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Complex)
				}
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
				switch ds.Single {
				case roles.STATUS_VALUE_TYPE_INT:
					errd = tran.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Int)
				case roles.STATUS_VALUE_TYPE_FLOAT:
					errd = tran.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Float)
				case roles.STATUS_VALUE_TYPE_COMPLEX:
					errd = tran.operators[one].WriteFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, ds.Complex)
				}
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

// 读friend状态
func (d *DRule) normalReadFriendStatus(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndFriend{}
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
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				err = d.trule.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				err = d.trule.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				err = d.trule.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Complex)
			}
		} else {
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				err = tran.tran.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				err = tran.tran.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				err = tran.tran.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Complex)
			}
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Complex)
			}
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			switch ds.Single {
			case roles.STATUS_VALUE_TYPE_INT:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Int)
			case roles.STATUS_VALUE_TYPE_FLOAT:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Float)
			case roles.STATUS_VALUE_TYPE_COMPLEX:
				errd = o.ReadFriendStatus(ds.Area, ds.Id, ds.Friend, ds.Bit, &ds.Complex)
			}
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

// 删朋友
func (d *DRule) normalDeleteFriend(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndFriend{}
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
			err = d.trule.DeleteFriend(ds.Area, ds.Id, ds.Friend)
		} else {
			err = tran.tran.DeleteFriend(ds.Area, ds.Id, ds.Friend)
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
				errd = d.operators[one].DeleteFriend(ds.Area, ds.Id, ds.Friend)
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
				errd = tran.operators[one].DeleteFriend(ds.Area, ds.Id, ds.Friend)
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

// 创建空的上下文
func (d *DRule) normalCreateContext(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndContext{}
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
			err = d.trule.CreateContext(ds.Area, ds.Id, ds.Context)
		} else {
			err = tran.tran.CreateContext(ds.Area, ds.Id, ds.Context)
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
				errd = d.operators[one].CreateContext(ds.Area, ds.Id, ds.Context)
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
				errd = tran.operators[one].CreateContext(ds.Area, ds.Id, ds.Context)
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

// 有context吗
func (d *DRule) normalExistContext(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndContext{}
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
			ds.Exist, err = d.trule.ExistContext(ds.Area, ds.Id, ds.Context)
		} else {
			ds.Exist, err = tran.tran.ExistContext(ds.Area, ds.Id, ds.Context)
		}
	} else {
		var errd operator.DRuleError
		if tran == nil {
			o, f := d.randomOneOperator(o_s)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Exist, errd = o.ExistContext(ds.Area, ds.Id, ds.Context)
		} else {
			o, f := d.randomOneOTransaction(o_s, tran)
			if f == false {
				errs = d.sendReceipt(conn_exec, operator.DATA_DRULE_OPERATOR_NO_EXIST, "Can not find remote operator.", nil)
				return
			}
			ds.Exist, errd = o.ExistContext(ds.Area, ds.Id, ds.Context)
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

// 删除上下文
func (d *DRule) normalDropContext(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend, tran *transactionMap) (errs error) {
	var err error
	// 解码，ds = data struct
	ds := operator.O_RoleAndContext{}
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
			err = d.trule.DropContext(ds.Area, ds.Id, ds.Context)
		} else {
			err = tran.tran.DropContext(ds.Area, ds.Id, ds.Context)
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
				errd = d.operators[one].DropContext(ds.Area, ds.Id, ds.Context)
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
				errd = tran.operators[one].DropContext(ds.Area, ds.Id, ds.Context)
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
