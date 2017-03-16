// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// ExecTCP的读取角色，需要判断是否在事务中
//	<-- 发送DATA_PLEASE
//	--> 接收角色id
//	<-- 判断是否在事务中
func (d *DRule) readRole(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}
	// 角色id
	role_id_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	role_id := string(role_id_b)
	// 查看位置
	connmode, conn := d.findConn(role_id)
	if connmode == CONN_IS_LOCAL {
		//本地处理
		var role roles.Roleer
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 在事务中
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			role, err = tran.ReadRole(role_id)
			if err != nil {
				return err
			}
		} else {
			// 不在事务中
			role, err = d.trule.ReadRole(role_id)
			if err != nil {
				return err
			}
		}
		err = d.sendRoleForExecTCP(role, conn_exec)
		if err != nil {
			return err
		}
	} else {
		// 在slave上
		role_send_b, err := d.getRoleFromSlaves(prefix_stat, role_id, conn)
		if err != nil {
			return err
		}
		// 发送数据
		err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, role_send_b, nil)
		return err
	}
	return
}

// 向请求方发送角色的本体
func (d *DRule) sendRoleForExecTCP(role roles.Roleer, conn_exec *nst.ConnExec) (err error) {
	// 构造发送结构
	roleb, relab, verb, err := hardstore.EncodeRole(role)
	if err != nil {
		return err
	}
	role_send := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	role_send_b, err := nst.StructGobBytes(role_send)
	if err != nil {
		return err
	}
	// 发送数据
	err = d.serverDataReceipt(conn_exec, DATA_ALL_OK, role_send_b, nil)
	return err
}

// 向slave发送请求角色
//	--> OPERATE_READ_ROLE（前导）
//	<-- 角色id
func (d *DRule) getRoleFromSlaves(prefix_stat Net_PrefixStat, roleid string, conns []*slaveIn) (role_send_b []byte, err error) {
	// 找一个随机的连接
	conncount := len(conns)
	connrandom := random.GetRandNum(conncount - 1)
	cprocess := conns[connrandom].tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_receipt, err := SendPrefixStat(cprocess, conns[connrandom].code, prefix_stat.TransactionId, OPERATE_READ_ROLE, prefix_stat.InTransaction)
	if err != nil {
		return
	}
	// 角色ID
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, []byte(roleid))
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	role_send_b = slave_receipt.Data
	return
}
