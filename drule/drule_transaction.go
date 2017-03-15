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

	"github.com/idcsource/Insight-0-0-lib/random"
)

// 创建事务
func (d *DRule) beginTransaction() (err error) {
	// 生成事务id
	tranid := random.GetRand(40)
	// 如果模式是master
	if d.dmode == DMODE_MASTER {
		var can []*slaveIn
		can, err = d.startTransactionForSlaves(tranid)
		if err != nil {
			// 向can发送关闭事务（Rollback）
			d.rollbackTransactionIfError(tranid, can)
			return
		}
	}
	// 生成本地的事务
	err = d.trule.beginForDRule(tranid)
	if err != nil {
		// 全部回滚事务
		d.rollbackTransaction(tranid)
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
func (d *DRule) startTransactionForOneSlave(tranid string, onec *slaveIn) (err error) {
	return
}

// 错误时候的部分回滚事务
func (d *DRule) rollbackTransactionIfError(tranid string, can []*slaveIn) {
	for _, onec := range can {
		d.rollbackOne(tranid, onec)
	}
}

// 回滚事务
func (d *DRule) rollbackTransaction(tranid string) (err error) {
	for _, onec := range d.slavecpool {
		d.rollbackOne(tranid, onec)
	}
	return
}

// 回滚事务
func (d *DRule) rollbackOne(tranid string, onec *slaveIn) (err error) {
	return
}
