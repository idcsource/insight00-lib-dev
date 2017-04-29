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

// ExecTCP nst的ConnExecer接口
func (d *DRule) ExecTCP(conn_exec *nst.ConnExec) (err error) {
	// 接收operator发送
	o_send_b, err := conn_exec.GetData()
	if err != nil {
		return
	}
	// 解码接受的信息
	o_send := operator.O_OperatorSend{}
	err = iendecode.BytesGobStruct(o_send_b, &o_send)
	if err != nil {
		return d.sendReceipt(conn_exec, operator.DATA_NOT_EXPECT, "Data not expect.", nil)
	}
	// 查看那些无论是否被暂停都会执行的命令
	// 判断是否被暂停再看剩余的命令
	return
}

// 发送O_DRuleReceipt
func (d *DRule) sendReceipt(conn_exec *nst.ConnExec, datastat operator.DRuleReturnStatus, errstr string, data []byte) (err error) {
	drule_r := operator.O_DRuleReceipt{
		DataStat: datastat,
		Error:    errstr,
		Data:     data,
	}
	drule_r_b, err := iendecode.StructGobBytes(drule_r)
	if err != nil {
		return
	}
	err = conn_exec.SendData(drule_r_b)
	return
}
