// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package porter

import (
	"fmt"

	"github.com/idcsource/insight00-lib/cpool"
	"github.com/idcsource/insight00-lib/hardstore"
	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/nst"
	"github.com/idcsource/insight00-lib/rolesio"
)

// 创建一个接收者
func NewReceiver(config *cpool.Section, logs *ilogs.Logs) (receiver *Receiver, err error) {
	if config == nil {
		config = cpool.NewSection("Receiver", "Receiver")
		receiver = &Receiver{
			logs:   logs,
			config: config,
		}
		return receiver, nil
	} else {
		receiver = &Receiver{
			logs:   logs,
			config: config,
			optype: OPERATE_NOT_SET,
		}
		// 添加身份码
		code, err := config.GetConfig("code")
		if err != nil {
			return nil, err
		}
		receiver.code = code
		// 建立TcpServer连接
		listen, err := config.GetConfig("listen")
		if err != nil {
			return nil, err
		}
		err = receiver.setServerListen(listen)
		if err != nil {
			return nil, err
		}
		return receiver, nil
	}
}

// 运行时设置日志
func (r *Receiver) SetLog(logs *ilogs.Logs) {
	r.logs = logs
}

// 注册存储器，SetStorage和SetOperater两个方法，将使用最后一个设置的，不会同时使用
func (r *Receiver) SetStorage(store rolesio.RolesInOutManager) {
	r.store = store
	r.optype = OPERATE_TO_STORE
}

// 注册处理者函数
func (r *Receiver) SetOperater(function ReceiverOperater) {
	r.function = function
	r.optype = OPERATE_TO_FUNCTION
}

// 设置身份验证码
func (r *Receiver) SetCode(code string) {
	r.config.SetConfig("code", code, "身份验证码")
	r.code = code
}

// 运行时改变监听
func (r *Receiver) SetServerListen(listen string) (err error) {
	r.config.SetConfig("listen", listen, "监听端口")
	return r.setServerListen(listen)
}

// 内部的建立监听连接方法
func (r *Receiver) setServerListen(listen string) (err error) {
	r.listen, err = nst.NewTcpServer(r, listen, r.logs)
	return
}

// nst.TcpServer要求的接口
//
// 直接接收信息，看情况进行处理或怎么样
func (r *Receiver) ExecTCP(ce *nst.ConnExec) (err error) {
	getsend_b, err := ce.GetData()
	// 这里是货物
	cargo := Cargo{}
	err = nst.BytesGobStruct(getsend_b, &cargo)
	if err != nil {
		return r.receiverReceipt(ce, DATA_NOT_EXPECT, err)
	}
	// 查看code码
	if cargo.ReceiverCode != r.code {
		return r.receiverReceipt(ce, DATA_NOT_EXPECT, fmt.Errorf("The Receiver Code is wrong."))
	}
	// 开始分别处理
	switch r.optype {
	case OPERATE_TO_STORE:
		err = r.operateStore(cargo)
	case OPERATE_TO_FUNCTION:
		err = r.operateFunction(cargo)
	default:
		return r.receiverReceipt(ce, DATA_NOT_EXPECT, fmt.Errorf("The Receiver not ready to receive."))
	}
	if err != nil {
		return r.receiverReceipt(ce, DATA_NOT_EXPECT, err)
	}
	return r.receiverReceipt(ce, DATA_ALL_OK, nil)
}

// 通过存储器保存的方法
func (r *Receiver) operateStore(cargo Cargo) (err error) {
	if r.store == nil {
		return fmt.Errorf("The Storage not set.")
	}
	role, err := hardstore.DecodeRole(cargo.RoleBody, cargo.RoleRela, cargo.RoleVer)
	if err != nil {
		return err
	}
	err = r.store.StoreRole(role)
	return err
}

// 通过函数处理
func (r *Receiver) operateFunction(cargo Cargo) (err error) {
	if r.function == nil {
		return fmt.Errorf("The operate function not set.")
	}
	role, err := hardstore.DecodeRole(cargo.RoleBody, cargo.RoleRela, cargo.RoleVer)
	if err != nil {
		return err
	}
	err = r.function.Operate(cargo.SenderName, role)
	return err
}

// 编码回执
func (r *Receiver) receiverReceipt(ce *nst.ConnExec, stat uint8, toerr error) (err error) {
	receipt := Net_ReceiverReceipt{
		DataStat: stat,
		Error:    fmt.Sprint(toerr),
	}
	receipt_b, err := nst.StructGobBytes(receipt)
	if err != nil {
		return err
	}
	err = ce.SendData(receipt_b)
	return err
}

// 处理错误日志
func (r *Receiver) logerr(err interface{}) {
	if err == nil {
		return
	}
	if r.logs != nil {
		r.logs.ErrLog(fmt.Errorf("porter[Receiver]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (r *Receiver) logrun(err interface{}) {
	if err == nil {
		return
	}
	if r.logs != nil {
		r.logs.RunLog(fmt.Errorf("porter[Receiver]: %v", err))
	} else {
		fmt.Println(err)
	}
}
