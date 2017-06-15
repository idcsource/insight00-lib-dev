// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"strings"
	"time"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst2"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 创建事务
func (d *DRule) normalTranBigen(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 解码
	o_t := operator.O_Transaction{}
	err = iendecode.BytesGobStruct(o_send.Data, &o_t)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}

	// 生成本地事务
	d_t := &transactionMap{
		tran_unid: o_t.TransactionId,
		operators: make(map[string]*operator.OTransaction),
	}

	itok := make([]string, 0)

	// 去蔓延
	if d.dmode == operator.DRULE_OPERATE_MODE_MASTER {
		for key, _ := range d.operators {
			var errd operator.DRuleError
			d_t.operators[key], errd = d.operators[key].Begin()
			if errd.IsError() == nil {
				itok = append(itok, key)
			} else {
				for _, k := range itok {
					d_t.operators[k].Rollback()
				}
				errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, errd.String(), nil)
				return
			}
		}
	}

	// 生成本地的事务
	d_t.tran, _ = d.trule.Begin()
	d_t.alivetime = time.Now()

	d.transaction_map[o_t.TransactionId] = d_t

	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 执行事务
func (d *DRule) normalTranCommit(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 如果不事务中
	if o_send.InTransaction == false || len(o_send.TransactionId) == 0 {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Not in a transaction.", nil)
		return
	}
	tran_map, find := d.transaction_map[o_send.TransactionId]
	if find == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Can not find transaction.", nil)
		return
	}

	errd_a := make([]string, 0)
	// 去蔓延
	if d.dmode == operator.DRULE_OPERATE_MODE_MASTER {
		for key, _ := range tran_map.operators {
			errd := tran_map.operators[key].Commit()
			if errd.IsError() != nil {
				errd_a = append(errd_a, errd.String())
			}
		}
	}

	// 本地的
	err = tran_map.tran.Commit()
	if err != nil {
		errd_a = append(errd_a, err.Error())
	}

	// 删除
	delete(d.transaction_map, o_send.TransactionId)

	if len(errd_a) != 0 {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, strings.Join(errd_a, " | "), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)

	return
}

// 回滚事务
func (d *DRule) normalTranRollback(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error
	// 如果不事务中
	if o_send.InTransaction == false || len(o_send.TransactionId) == 0 {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Not in a transaction.", nil)
		return
	}
	tran_map, find := d.transaction_map[o_send.TransactionId]
	if find == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Can not find transaction.", nil)
		return
	}

	errd_a := make([]string, 0)
	// 去蔓延
	if d.dmode == operator.DRULE_OPERATE_MODE_MASTER {
		for key, _ := range tran_map.operators {
			errd := tran_map.operators[key].Rollback()
			if errd.IsError() != nil {
				errd_a = append(errd_a, errd.String())
			}
		}
	}

	// 本地的
	err = tran_map.tran.Rollback()
	if err != nil {
		errd_a = append(errd_a, err.Error())
	}

	// 删除
	delete(d.transaction_map, o_send.TransactionId)

	if len(errd_a) != 0 {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, strings.Join(errd_a, " | "), nil)
		return
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)

	return
}

// 锁定角色
func (d *DRule) normalLockRole(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var err error

	// 如果不事务中
	if o_send.InTransaction == false || len(o_send.TransactionId) == 0 {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Not in a transaction.", nil)
		return
	}
	// 获取到事务
	tran_map, find := d.transaction_map[o_send.TransactionId]
	if find == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Can not find transaction.", nil)
		return
	}
	// 解码
	ot := operator.O_Transaction{}
	err = iendecode.BytesGobStruct(o_send.Data, &ot)
	if err != nil {
		errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
		return
	}
	// 查看区域的权限
	have := d.checkUserNormalPower(o_send.User, ot.Area, true)
	if have == false {
		errs = d.sendReceipt(conn_exec, operator.DATA_USER_NO_AUTHORITY, "", nil)
		return
	}
	// 查看运行模式
	if d.dmode == operator.DRULE_OPERATE_MODE_MASTER {
		err_a := make([]string, 0)
		for _, roleid := range ot.PrepareIDs {
			// 获得角色的问题
			position, o := d.getRolePosition(ot.Area, roleid)
			if position == ROLE_POSITION_IN_REMOTE {
				for _, name := range o {
					errd := tran_map.operators[name].LockRole(ot.Area, roleid)
					err_a = append(err_a, errd.String())
				}
			} else {
				err = tran_map.tran.LockRole(ot.Area, roleid)
				err_a = append(err_a, err.Error())
			}
		}
		if len(err_a) != 0 {
			errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, strings.Join(err_a, " | "), nil)
			return
		}
	} else {
		err = tran_map.tran.LockRole(ot.Area, ot.PrepareIDs...)
		if err != nil {
			errs = d.sendReceipt(conn_exec, operator.DATA_RETURN_ERROR, err.Error(), nil)
			return
		}
	}
	errs = d.sendReceipt(conn_exec, operator.DATA_ALL_OK, "", nil)
	return
}

// 查看是否在事务中
func (d *DRule) checkTranOrNoTran(conn_exec *nst2.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
	var tran *transactionMap

	// 查看事务情况
	if o_send.InTransaction == true && len(o_send.TransactionId) != 0 {
		// 找到tran
		var find bool
		tran, find = d.transaction_map[o_send.TransactionId]
		if find == false {
			errs = d.sendReceipt(conn_exec, operator.DATA_TRAN_ERROR, "Can not find transaction.", nil)
			return
		}
	}

	switch o_send.Operate {
	case operator.OPERATE_EXIST_ROLE:
		// 是否存在角色
		errs = d.normalExitRole(conn_exec, o_send, tran)
	case operator.OPERATE_READ_ROLE:
		// 读取角色
		errs = d.normalReadRole(conn_exec, o_send, tran)
	case operator.OPERATE_WRITE_ROLE:
		// 写角色
		errs = d.normalStoreRole(conn_exec, o_send, tran)
	case operator.OPERATE_DEL_ROLE:
		// 删角色
		errs = d.normalDeleteRole(conn_exec, o_send, tran)
	case operator.OPERATE_SET_FATHER:
		// 写father
		errs = d.normalWriteFather(conn_exec, o_send, tran)
	case operator.OPERATE_GET_FATHER:
		// 读father
		errs = d.normalReadFather(conn_exec, o_send, tran)
	case operator.OPERATE_GET_CHILDREN:
		// 读children
		errs = d.normalReadChildren(conn_exec, o_send, tran)
	case operator.OPERATE_SET_CHILDREN:
		// 写children
		errs = d.normalWriteChildren(conn_exec, o_send, tran)
	case operator.OPERATE_ADD_CHILD:
		// 写child
		errs = d.normalWriteChild(conn_exec, o_send, tran)
	case operator.OPERATE_DEL_CHILD:
		// 删child
		errs = d.normalDeleteChild(conn_exec, o_send, tran)
	case operator.OPERATE_EXIST_CHILD:
		// 存在child
		errs = d.normalExistChild(conn_exec, o_send, tran)
	case operator.OPERATE_GET_FRIENDS:
		// 读friends
		errs = d.normalReadFriends(conn_exec, o_send, tran)
	case operator.OPERATE_SET_FRIENDS:
		// 写friends
		errs = d.normalWriteFriends(conn_exec, o_send, tran)
	case operator.OPERATE_SET_FRIEND_STATUS:
		// 写friend状态
		errs = d.normalWriteFriendStatus(conn_exec, o_send, tran)
	case operator.OPERATE_GET_FRIEND_STATUS:
		// 读friend状态
		errs = d.normalReadFriendStatus(conn_exec, o_send, tran)
	case operator.OPERATE_DEL_FRIEND:
		// 删friend
		errs = d.normalDeleteFriend(conn_exec, o_send, tran)
	case operator.OPERATE_ADD_CONTEXT:
		// 创建上下文
		errs = d.normalCreateContext(conn_exec, o_send, tran)
	case operator.OPERATE_EXIST_CONTEXT:
		// 是否存在上下文
		errs = d.normalExistContext(conn_exec, o_send, tran)
	case operator.OPERATE_DROP_CONTEXT:
		// 删除上下文
		errs = d.normalDropContext(conn_exec, o_send, tran)
	case operator.OPERATE_READ_CONTEXT:
		// 读上下文
		errs = d.normalReadContext(conn_exec, o_send, tran)
	case operator.OPERATE_DEL_CONTEXT_BIND:
		// 删上下文绑定
		errs = d.normalDeleteContextBind(conn_exec, o_send, tran)
	case operator.OPERATE_CONTEXT_SAME_BIND:
		// 上下文同样绑定
		errs = d.normalReadContextSameBind(conn_exec, o_send, tran)
	case operator.OPERATE_GET_CONTEXTS_NAME:
		// 上下文名称
		errs = d.normalReadContextsName(conn_exec, o_send, tran)
	case operator.OPERATE_SET_CONTEXT_STATUS:
		// 设定上下文状态
		errs = d.normalWriteContextStatus(conn_exec, o_send, tran)
	case operator.OPERATE_GET_CONTEXT_STATUS:
		// 读取上下文状态
		errs = d.normalReadContextStatus(conn_exec, o_send, tran)
	case operator.OPERATE_SET_CONTEXTS:
		// 写contexts
		errs = d.normalWriteContexts(conn_exec, o_send, tran)
	case operator.OPERATE_GET_CONTEXTS:
		// 读contexts
		errs = d.normalReadContexts(conn_exec, o_send, tran)
	case operator.OPERATE_SET_DATA:
		// 写数据
		errs = d.normalWriteData(conn_exec, o_send, tran)
	case operator.OPERATE_GET_DATA:
		// 读数据
		errs = d.normalReadData(conn_exec, o_send, tran)
	default:
		return d.sendReceipt(conn_exec, operator.DATA_NOT_EXPECT, "No operate.", nil)
	}
	return
}

// 从给出的名字中随机获取一个oprator
func (d *DRule) randomOneOperator(o_s []string) (o *operator.Operator, find bool) {
	lens := len(o_s)
	var r_i int
	if lens == 1 {
		r_i = 0
	} else {
		r_i = random.GetRandNum(lens - 1)
	}
	o, find = d.operators[o_s[r_i]]
	return
}

// 从给出的名字中随机获取一个事务的operator
func (d *DRule) randomOneOTransaction(o_s []string, trano *transactionMap) (o *operator.OTransaction, find bool) {
	lens := len(o_s)
	var r_i int
	if lens == 1 {
		r_i = 0
	} else {
		r_i = random.GetRandNum(lens - 1)
	}
	o, find = trano.operators[o_s[r_i]]
	return
}
