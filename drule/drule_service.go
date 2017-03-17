// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/nst"
)

// ExecTCP nst的ConnExecer接口
func (d *DRule) ExecTCP(conn_exec *nst.ConnExec) (err error) {
	// 接收前导
	prefix_stat_b, err := conn_exec.GetData()
	if err != nil {
		d.logerr(fmt.Errorf("Get the Prefix Stat err : %v", err))
		return err
	}
	// 解码前导
	prefix_stat := Net_PrefixStat{}
	err = nst.BytesGobStruct(prefix_stat_b, &prefix_stat)
	if err != nil {
		d.logerr(fmt.Errorf("Get the Prefix Stat err : %v", err))
		return err
	}
	// 检查身份验证码
	if prefix_stat.Code != d.code {
		// 发送错误
		d.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, fmt.Errorf("The service code is wrong : %v .", prefix_stat.Code))
		conn_exec.SendClose()
		return fmt.Errorf("The service code is wrong : %v .", prefix_stat.Code)
	}
	// 开始遍历操作
	switch prefix_stat.Operate {
	case OPERATE_TRAN_BEGIN:
		// 开启事务，如果出错则自己负责回滚掉开启的slave的事务
		err = d.beginTransaction(conn_exec)
	case OPERATE_TRAN_ROLLBACK:
		// 回滚事务
		if prefix_stat.InTransaction == false || len(prefix_stat.TransactionId) == 0 {
			err = fmt.Errorf("This is not in a transaction.")
		} else {
			err = d.rollbackTransaction(prefix_stat.TransactionId, conn_exec)
		}
	case OPERATE_TRAN_COMMIT:
		// 执行事务
		if prefix_stat.InTransaction == false || len(prefix_stat.TransactionId) == 0 {
			err = fmt.Errorf("This is not in a transaction.")
		} else {
			err = d.commitTransaction(prefix_stat.TransactionId, conn_exec)
		}
	case OPERATE_READ_ROLE:
		err = d.readRole(prefix_stat,conn_exec)
		// 读取角色
	case OPERATE_WRITE_ROLE:
		err = d.storeRole(prefix_stat,conn_exec)

	default:
		err = d.serverDataReceipt(conn_exec, DATA_NOT_EXPECT, nil, fmt.Errorf("The oprerate can not found."))
		conn_exec.SendClose()
		return nil
	}
	if err != nil {
		err = d.serverDataReceipt(conn_exec, DATA_RETURN_ERROR, nil, err)
		return nil
	} else {
		//统一发送DATA_ALL_OK的回执
		err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, nil, nil)
	}
	return
}

// 向请求方返回带有数据体的回执信息
func (d *DRule) serverDataReceipt(conn_exec *nst.ConnExec, stat uint8, data []byte, err error) (err2 error) {
	// 构建回执
	slave_receipt := Net_SlaveReceipt_Data{
		DataStat: stat,
		Error:    fmt.Sprint(err),
		Data:     data,
	}
	slave_receipt_b, err2 := nst.StructGobBytes(slave_receipt)
	if err2 != nil {
		return err2
	}
	err2 = conn_exec.SendData(slave_receipt_b)
	if err2 != nil {
		return err2
	}
	return nil
}
