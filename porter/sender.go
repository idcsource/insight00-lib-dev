// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package porter

import (
	"errors"
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 创建一个发送者
func NewSender(config *cpool.Block, logs *ilogs.Logs) (sender *Sender, err error) {
	sender = &Sender{
		logs: logs,
	}
	// 如果config是nil，则建立一个空的config
	if config == nil {
		sender.config = cpool.NewBlock("Sender", "Sender")
	}
	// 如果配置不是nil，则要处理许多事情了
	if config != nil {
		sender.name, _ = sender.config.GetConfig("main.name")
		// 创造连接
		err = sender.doConnect()
		if err != nil {
			err = fmt.Errorf("porter[Sender]NewSender: %v", err)
			return nil, err
		}
	}
	return
}

// 向所有接收者发送角色
func (s *Sender) SendRole(role roles.Roleer) (err error) {
	var errstr string
	for _, one := range s.receivers {
		err1 := s.sendTo(one, role)
		if err1 != nil {
			errstr += one.name + ": " + fmt.Sprint(err1) + " | "
		}
	}
	if len(errstr) != 0 {
		return fmt.Errorf("porter[Sender]NewSender: %v", errstr)
	}
	return nil
}

// 向某一个接收者发送角色
func (s *Sender) SendRoleToReceiver(receiver string, role roles.Roleer) (err error) {
	// 先看一眼有没有这个接收者
	receiver_conn, have := s.receivers[receiver]
	if have == false {
		return fmt.Errorf("porter[Sender]SendRoleToReceiver: Can not find the receiver %v.", receiver)
	}
	return s.sendTo(receiver_conn, role)
}

// 内部的角色发送
func (s *Sender) sendTo(receiver oneReceiver, role roles.Roleer) (err error) {
	role_body, role_rela, role_ver, err := hardstore.EncodeRole(role)
	if err != nil {
		return err
	}
	// 构造发送
	cargo := Cargo{
		Id:           role.ReturnId(),
		SenderName:   s.name,
		ReceiverCode: receiver.code,
		RoleBody:     role_body,
		RoleRela:     role_rela,
		RoleVer:      role_ver,
	}
	// 编码
	cargo_b, err := nst.StructGobBytes(cargo)
	if err != nil {
		return err
	}
	// 分配连接
	cprocess := receiver.tcpconn.OpenProgress()
	defer cprocess.Close()
	// 发送数据
	return_byte, err := cprocess.SendAndReturn(cargo_b)
	// 转码收到的信息
	return_data := Net_ReceiverReceipt{}
	err = nst.BytesGobStruct(return_byte, &return_data)
	if err != nil {
		return err
	}
	if return_data.DataStat != DATA_ALL_OK {
		return errors.New(return_data.Error)
	}
	return nil
}

// 运行时设置配置
func (s *Sender) SetConfig(config *cpool.Block) (err error) {
	if config == nil {
		return fmt.Errorf("porter[Sender]SetConfig: Can not set nil config.")
	}
	s.name, err = s.config.GetConfig("main.name")
	if err != nil {
		return fmt.Errorf("porter[Sender]SetConfig: Can not find the name config.")
	}
	// 创造连接
	err = s.doConnect()
	if err != nil {
		err = fmt.Errorf("porter[Sender]SetConfig: %v", err)
		return err
	}
	return nil
}

// 运行时设置日志
func (s *Sender) SetLog(logs *ilogs.Logs) {
	s.logs = logs
}

// 内部的重新建立连接的机制
func (s *Sender) doConnect() (err error) {
	s.receivers = make(map[string]oneReceiver)
	// 获得所有的列表
	receivers, err := s.config.GetEnum("main.receiver")
	if err != nil {
		return err
	}
	for _, one := range receivers {
		onecfg, err2 := s.config.GetSection(one)
		if err2 != nil {
			s.Close()
			return err2
		}
		// 获取地址
		address, err2 := onecfg.GetConfig("address")
		if err2 != nil {
			s.Close()
			return err2
		}
		// 获取身份码
		code, err2 := onecfg.GetConfig("code")
		if err2 != nil {
			s.Close()
			return err2
		}
		// 获取连接数
		var conn_num int
		conn_num64, err2 := onecfg.TranInt64("conn_num")
		if err2 != nil {
			conn_num = 1
		} else {
			conn_num = int(conn_num64)
		}
		// 创建连接
		sconn, err2 := nst.NewTcpClient(address, conn_num, s.logs)
		if err != nil {
			s.Close()
			return err2
		}
		s.receivers[one] = oneReceiver{
			name:    one,
			code:    code,
			address: address,
			tcpconn: sconn,
		}
	}
	return
}

// 关闭连接
func (s *Sender) Close() {
	for _, one := range s.receivers {
		one.tcpconn.Close()
	}
}

// 处理错误日志
func (s *Sender) logerr(err interface{}) {
	if err == nil {
		return
	}
	if s.logs != nil {
		s.logs.ErrLog(fmt.Errorf("porter[Sender]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (s *Sender) logrun(err interface{}) {
	if err == nil {
		return
	}
	if s.logs != nil {
		s.logs.RunLog(fmt.Errorf("porter[Sender]: %v", err))
	} else {
		fmt.Println(err)
	}
}
