// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"fmt"
	"sync"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 新建一个控制器，addr和code是默认的ZrStorage的地址（含端口号）和身份码，conn_num为连接池的个数
func NewOperator(addr, code string, conn_num int, logs *ilogs.Logs) (operator *Operator, err error) {
	operator = &Operator{
		slaves: make([]*slaveIn, 0),
		logs:   logs,
		lock:   new(sync.RWMutex),
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

// 增加一个服务器到控制器
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

// 运行时保存
func (o *Operator) ToStore() (err error) {
	for _, onec := range o.slaves {
		// 分配连接
		cprocess := onec.tcpconn.OpenProgress()
		defer cprocess.Close()
		// 发送前导
		slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_TOSTORE)
		if err != nil {
			o.logerr(err)
			//return err
		}
		if slave_receipt.DataStat != DATA_ALL_OK {
			o.logerr(slave_receipt.Error)
			//return slave_receipt.Error
		}
	}
	return nil
}

// 读取角色
func (o *Operator) ReadRole(id string) (role roles.Roleer, err error) {
	// 随机一个连接
	conncount := len(o.slaves)
	connrandom := random.GetRandNum(conncount - 1)
	role, err = o.readRole(id, o.slaves[connrandom])
	if err != nil {
		err = fmt.Errorf("drcm[Operator]ReadRole: %v", err)
		return nil, err
	}
	return role, nil
}

func (o *Operator) readRole(id string, slave *slaveIn) (role roles.Roleer, err error) {
	cprocess := slave.tcpconn.OpenProgress()
	defer cprocess.Close()
	slavereceipt, err := SendPrefixStat(cprocess, slave.code, OPERATE_READ_ROLE)
	if err != nil {
		return nil, err
	}
	// 如果获取到的DATA_ALL_OK则说明认证已经通过
	if slavereceipt.DataStat != DATA_PLEASE {
		return nil, slavereceipt.Error
	}
	// 发送想要的id，并接收slave的返回
	sreb, err := cprocess.SendAndReturn([]byte(id))
	if err != nil {
		return nil, err
	}
	// 解码返回值
	slavereceipt, err = DecodeSlaveReceipt(sreb)
	if err != nil {
		return nil, err
	}
	// 如果回执状态不是DATA_WILL_SEND，因为我们希望slave是应该把role发送给我们的
	if slavereceipt.DataStat != DATA_WILL_SEND {
		return nil, slavereceipt.Error
	}
	// 请求对方发送数据，使用DATA_PLEASE状态，并接收角色的byte流，这是一个Net_RoleSendAndReceive的值。
	dataplace := nst.Uint8ToBytes(DATA_PLEASE)
	rdata, err := cprocess.SendAndReturn(dataplace)
	if err != nil {
		return nil, err
	}
	// 解码Net_RoleSendAndReceive。
	rolegetstruct := Net_RoleSendAndReceive{}
	err = nst.BytesGobStruct(rdata, &rolegetstruct)
	if err != nil {
		return nil, err
	}
	// 合成出role来
	role, err = hardstore.DecodeRole(rolegetstruct.RoleBody, rolegetstruct.RoleRela, rolegetstruct.RoleVer)
	return role, err
}

// 存储角色
func (o *Operator) StoreRole(role roles.Roleer) (err error) {
	// 角色编码
	roleb, relab, verb, err := hardstore.EncodeRole(role)
	if err != nil {
		return err
	}
	roleS := Net_RoleSendAndReceive{
		RoleBody: roleb,
		RoleRela: relab,
		RoleVer:  verb,
	}
	roleS_b, err := nst.StructGobBytes(roleS)
	if err != nil {
		return err
	}
	// 遍历slave的连接，如果slave出现错误就输出，继续下一个结点
	var errstring string
	for _, onec := range o.slaves {
		err = o.storeRole(roleS_b, onec)
		if err != nil {
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring)
	}
	return nil
}

func (o *Operator) storeRole(roleS_b []byte, onec *slaveIn) (err error) {
	cprocess := onec.tcpconn.OpenProgress()
	defer cprocess.Close()
	//发送前导
	slavereceipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_WRITE_ROLE)
	if err != nil {
		return err
	}
	// 如果slave请求发送数据
	if slavereceipt.DataStat == DATA_PLEASE {
		srb, err := cprocess.SendAndReturn(roleS_b)
		if err != nil {
			return err
		}
		sr, err := DecodeSlaveReceipt(srb)
		if err != nil {
			return err
		}
		if sr.DataStat != DATA_ALL_OK {
			return sr.Error
		}
		return nil
	} else {
		return slavereceipt.Error
	}
}

// 处理错误日志
func (o *Operator) logerr(err interface{}) {
	if err == nil {
		return
	}
	if o.logs != nil {
		o.logs.ErrLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (o *Operator) logrun(err interface{}) {
	if err == nil {
		return
	}
	if o.logs != nil {
		o.logs.RunLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}
