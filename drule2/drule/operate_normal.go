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
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 创建事务
func (d *DRule) normalTranBigen(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
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
			if errd.IsError() != nil {
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
func (d *DRule) normalTranCommit(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
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
func (d *DRule) normalTranRollback(conn_exec *nst.ConnExec, o_send *operator.O_OperatorSend) (errs error) {
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
