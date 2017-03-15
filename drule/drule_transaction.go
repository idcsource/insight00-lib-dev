// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 创建事务
// <-- 发送DATA_PLEASE（slave回执）
// --> 接收创建事务的ID
func (d *DRule) beginTransaction(conn_exec *nst.ConnExec) (err error) {
	// 发送回执
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 接收事务的创建ID
	tranid_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	tranid := string(tranid_b)

	// 如果模式是master，则向所有slave发送创建ID
	if d.dmode == DMODE_MASTER {
		var can []*slaveIn
		can, err = d.startTransactionForSlaves(tranid)
		if err != nil {
			// 向can发送关闭事务（Rollback）
			d.rollbackTransactionIfError(tranid, can)
			// 返回失败，交由ExecTCP发送失败信息
			return err
		}
	}
	// 生成本地自身的事务
	err = d.trule.beginForDRule(tranid)
	if err != nil && d.dmode == DMODE_MASTER {
		// 全部回滚事务
		d.rollbackTransactionAll(tranid)
	}
	return
}

// slave的事务创建
func (d *DRule) startTransactionForSlaves(tranid string) (can []*slaveIn, err error) {
	can = make([]*slaveIn, 0)
	errarray := make([]string, 0)
	for _, onec := range d.slavecpool {
		errn := d.startTransactionForOneSlave(tranid, onec)
		if errn != nil {
			errarray = append(errarray, errn.Error())
		} else {
			can = append(can, onec)
		}
	}
	if len(errarray) != 0 {
		errstr := strings.Join(errarray, " | ")
		err = fmt.Errorf(errstr)
	}
	return
}

// slave的单个事务创建
// --> 发送请求OPERATE_TRAN_BEGIN（前导）
// <-- DATA_PLEASE（回执）
// --> tranid
// <-- DATA_ALL_OK（回执）
func (d *DRule) startTransactionForOneSlave(tranid string, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, onec.code, tranid, OPERATE_TRAN_BEGIN, false)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送tranid
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, []byte(tranid))
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return
}

// 错误时候的部分回滚事务
func (d *DRule) rollbackTransactionIfError(tranid string, can []*slaveIn) {
	for _, onec := range can {
		d.rollbackOne(tranid, onec)
	}
}

// 全部回滚事务
func (d *DRule) rollbackTransactionAll(tranid string) (err error) {
	for _, onec := range d.slavecpool {
		d.rollbackOne(tranid, onec)
	}
	return
}

func (d *DRule) rollbackTransaction(conn_exec *nst.ConnExec) (err error) {
	return
}

// 回滚事务
func (d *DRule) rollbackOne(tranid string, onec *slaveIn) (err error) {
	return
}
