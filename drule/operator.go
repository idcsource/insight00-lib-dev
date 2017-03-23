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

	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个操作机，addr和code是默认的drule的地址（含端口号）和身份码，conn_num为连接池的个数
func NewOperator(selfname string, addr, code string, conn_num int, logs *ilogs.Logs) (operator *Operator, err error) {
	operator = &Operator{
		selfname:      selfname,
		slaves:        make([]*slaveIn, 0),
		inTransaction: false,
		logs:          logs,
	}
	slave, err := nst.NewTcpClient(addr, conn_num, logs)
	if err != nil {
		return nil, err
	}
	oneSlaveIn := &slaveIn{
		name:    addr,
		code:    code,
		tcpconn: slave,
	}
	operator.slaves = append(operator.slaves, oneSlaveIn)
	return operator, nil
}

// 增加一个服务器到控制器，addr和code是默认的drule的地址（含端口号）和身份码，conn_num为连接池的个数
func (o *Operator) AddServer(addr, code string, conn_num int) (err error) {
	slave, err := nst.NewTcpClient(addr, conn_num, o.logs)
	if err != nil {
		return err
	}
	oneSlaveIn := &slaveIn{
		name:    addr,
		code:    code,
		tcpconn: slave,
	}
	o.slaves = append(o.slaves, oneSlaveIn)
	return nil
}

// 创建事务
func (o *Operator) Begin() (operator *Operator, err error) {
	if o.inTransaction == true {
		return nil, fmt.Errorf("There's no function Begin.")
	}
	return o.beginTransaction()
}

// 创建事务
func (o *Operator) Transaction() (operator *Operator, err error) {
	if o.inTransaction == true {
		return nil, fmt.Errorf("There's no function Transaction.")
	}
	return o.beginTransaction()
}

// 内部的创建事务
func (o *Operator) beginTransaction() (operator *Operator, err error) {
	// 生成事务ID
	tranid := random.GetRand(40)
	// 对所有镜像开启
	can := make([]*slaveIn, 0)
	errall := make([]string, 0)
	for key := range o.slaves {
		errone := o.startTransactionForOne(tranid, o.slaves[key])
		if errone != nil {
			errall = append(errall, errone.Error())
		} else {
			can = append(can, o.slaves[key])
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
		o.rollbackTransactionIfError(tranid, can)
		return
	}
	// 对本地开启
	operator = &Operator{
		selfname:      o.selfname,
		inTransaction: true,
		transactionId: tranid,
		slaves:        o.slaves,
		logs:          o.logs,
	}
	return
}

// 错误时候的部分回滚事务
func (o *Operator) rollbackTransactionIfError(tranid string, can []*slaveIn) {
	for _, onec := range can {
		o.rollbackSlaveOne(tranid, onec)
	}
}

// 向某一个slave发送的回滚事务
// --> 发送请求OPERATE_TRAN_ROLLBACK（前导）
// <-- DATA_ALL_OK，接收回执
func (o *Operator) rollbackSlaveOne(tranid string, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()

	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, o.selfname, onec.code, tranid, true, "", OPERATE_TRAN_ROLLBACK)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return
}

// 开启一个的事务
func (o *Operator) startTransactionForOne(tranid string, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, o.selfname, onec.code, tranid, false, "", OPERATE_TRAN_BEGIN)
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

// 回滚事务
func (o *Operator) Rollback() (err error) {
	if o.inTransaction == false {
		err = fmt.Errorf("There's no function Rollback.")
		return
	}
	errall := make([]string, 0)
	for key := range o.slaves {
		errone := o.rollbackSlaveOne(o.transactionId, o.slaves[key])
		if errone != nil {
			errall = append(errall, errone.Error())
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
		return
	}
	return
}

// 执行事务
func (o *Operator) Commit() (err error) {
	if o.inTransaction == false {
		err = fmt.Errorf("There's no function Rollback.")
		return
	}
	errall := make([]string, 0)
	for key := range o.slaves {
		errone := o.commitSlaveOne(o.transactionId, o.slaves[key])
		if errone != nil {
			errall = append(errall, errone.Error())
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
		return
	}
	return
}

// 针对某一个slave的执行事务
func (o *Operator) commitSlaveOne(tranid string, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()

	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, o.selfname, onec.code, tranid, true, "", OPERATE_TRAN_COMMIT)
	if err != nil {
		return err
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		return fmt.Errorf(slave_receipt.Error)
	}
	return
}

// 运行时保存，用的话就弹错误
func (o *Operator) ToStore() (err error) {
	err = fmt.Errorf("drule[Operator]ToStore: Transaction does not provide this method.")
	return
}

// 发送前导并返回解码后的data数据
// slavercode为服务端的验证字符串，roleid为涉及到的角色ID，operate为操作的类型（OPERATE_*），senddata为发送的数据，returndata为要求装入的内容
func (o *Operator) sendReadAndDecodeData(roleid string, operate int, senddata []byte, returndata interface{}) (slave_receipt Net_SlaveReceipt_Data, err error) {
	// 随机一个连接
	onec := o.randomLink()
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err = SendPrefixStat(cprocess, o.selfname, onec.code, o.transactionId, o.inTransaction, roleid, operate)
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	// 发送数据
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, senddata)
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	// 解码装入
	err = nst.BytesGobStruct(slave_receipt.Data, returndata)
	if err != nil {
		return
	}
	return
}

// 向服务器发送写入命令
// roleid为涉及的角色ID，operate为操作类型（OPERATE_*），senddata为发送的数据
func (o *Operator) sendWriteToServer(roleid string, operate int, senddata []byte) (err error) {
	errall := make([]string, 0)
	for key := range o.slaves {
		errone := o.sendWriteToOneServer(roleid, operate, senddata, o.slaves[key])
		if errone != nil {
			errall = append(errall, errone.Error())
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
		return
	}
	return
}

func (o *Operator) sendWriteToOneServer(roleid string, operate int, senddata []byte, onec *slaveIn) (err error) {
	// 分配连接
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送前导
	slave_receipt, err := SendPrefixStat(cprocess, o.selfname, onec.code, o.transactionId, o.inTransaction, roleid, operate)
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_PLEASE {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	// 发送数据
	slave_receipt, err = SendAndDecodeSlaveReceiptData(cprocess, senddata)
	if err != nil {
		return
	}
	if slave_receipt.DataStat != DATA_ALL_OK {
		err = fmt.Errorf(slave_receipt.Error)
		return
	}
	return
}

// 随机一个连接
func (o *Operator) randomLink() (conn *slaveIn) {
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	return o.slaves[connrandom]
}
