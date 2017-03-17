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

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// ExecTCP的读取角色，需要判断是否在事务中
//	<-- 发送DATA_PLEASE
//	--> 接收角色id
//	<-- Net_RoleSendAndReceive
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
	slave_receipt, err := SendPrefixStat(cprocess, conns[connrandom].code, prefix_stat.TransactionId, prefix_stat.InTransaction, roleid, OPERATE_READ_ROLE)
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

// 为ExecTCP的保存角色
//	--> DATA_PLEASE
//	<-- Net_RoleSendAndReceive
//	--> 交由ExecTCP发送DATA_ALL_OK
func (d *DRule) storeRole(prefix_stat Net_PrefixStat, conn_exec *nst.ConnExec) (err error) {
	// 看看有roleid吗
	roleid := prefix_stat.RoleId
	if len(roleid) == 0 {
		err = fmt.Errorf("Don't set the Role id.")
		return err
	}
	// 发送DATA_PLEASE
	err = d.serverDataReceipt(conn_exec, DATA_PLEASE, nil, nil)
	if err != nil {
		return err
	}

	// 接收Net_RoleSendAndReceive
	role_body_b, err := conn_exec.GetData()
	if err != nil {
		return err
	}
	// 查看角色应该在哪里保存
	connmode, conn := d.findConn(roleid)
	if connmode == CONN_IS_LOCAL {
		// 本地处理
		// 解码角色
		role_body := Net_RoleSendAndReceive{}
		err = nst.BytesGobStruct(role_body_b, &role_body)
		if err != nil {
			return err
		}
		role, err := hardstore.DecodeRole(role_body.RoleBody, role_body.RoleRela, role_body.RoleVer)
		if err != nil {
			return err
		}
		// 查看是不是在事务中
		if prefix_stat.InTransaction == true && len(prefix_stat.TransactionId) != 0 {
			// 在事务中
			tran, err := d.trule.getTransactionForDRule(prefix_stat.TransactionId)
			if err != nil {
				return err
			}
			err = tran.StoreRole(role)
			if err != nil {
				return err
			}
		} else {
			// 不在事务中
			err = d.trule.StoreRole(role)
			if err != nil {
				return err
			}
		}
	} else {
		// 在slave上的话
		err = d.storeRoleForSlaves(prefix_stat, role_body_b, conn)
		if err != nil {
			return err
		}
	}
	return
}

// 为所有slave的存储角色
func (d *DRule) storeRoleForSlaves(prefix_stat Net_PrefixStat, role_body_b []byte, conns []*slaveIn) (err error) {
	errarray := make([]string, 0)
	for _, conn := range conns {
		errone := d.storeRoleForOneSlave(prefix_stat, role_body_b, conn)
		if errone != nil {
			errarray = append(errarray, fmt.Sprint(errone))
		}
	}
	if len(errarray) != 0 {
		errstr := strings.Join(errarray, " | ")
		return fmt.Errorf(errstr)
	}
	return
}

// 为某一个slave存储角色
//	--> OPERATE_WRITE_ROLE(前导)
//	<-- DATA_PLEASE
//	--> Net_RoleSendAndReceive(byte)
//	<-- DATA_ALL_OK
func (d *DRule) storeRoleForOneSlave(prefix_stat Net_PrefixStat, role_body_b []byte, conn *slaveIn) (err error) {
	// 分配连接
	cprocess := conn.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 前导
	slave_receipt, err := SendPrefixStat(cprocess, conn.code, prefix_stat.TransactionId, prefix_stat.InTransaction, prefix_stat.RoleId, OPERATE_WRITE_ROLE)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		return fmt.Errorf(slave_receipt.Error)
	}
	// 发送Net_RoleSendAndReceive
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, role_body_b)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return nil
}
