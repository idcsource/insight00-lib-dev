// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs

import (
	"time"
	"fmt"
	
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/bridges"
)

// NodeSmcs 节点状态的设置以及定期的发送。
// 将通过指定nst.TcpClient发送，
// 中心的返回值将由nst.TcpClient提供的桥发送给TcpReturn()方法。
// 而TcpReturn()将继续通过与外部连接的桥将数据发送出去，此桥的接收者需要提供SmcsReturn()方法。
// 方法原型见rolesplus包描述。
//
// NodeSmcs继承自rolesplus.RolePlus，因此拥有了处理桥的功能。
// 但新建和管理并不能使用rcontrol，请使用NewNodeSmcs()函数。
type NodeSmcs struct {
	rolesplus.RolePlus
	runtimeid		string							// 运行时UNID
	outbridge		*bridges.Bridge					// 输出通讯桥
	tcpc			*nst.TcpClient					// 作为Client的发送方需要知道交给谁来发送
	nodesend		NodeSend						// 发送出去的类型
	logs			*ilogs.Logs						// 运行日志
}

// 新建一个节点使用的自身状态发送方法。
// tcpc为一个建立好的中心的连接。
func NewNodeSmcs (tcpc *nst.TcpClient, logs *ilogs.Logs) *NodeSmcs {
	runtimeid := random.Unid(1,"NODE_SMCS");
	outbridge := bridges.NewBridge(logs);		// 建立外传的桥
	nsmcs := &NodeSmcs{
		runtimeid : runtimeid,
		outbridge: outbridge,
		tcpc : tcpc,
		nodesend : NodeSend{ 
			WorkSet : WORK_SET_STOP,
			Status : NODE_STATUS_NO_CONFIG,
			RunLog : make([]string, 0),
			ErrLog : make([]string, 0),
		} ,
		logs : logs,
	}
	nsmcs.SetLog(logs);
	nsmcs.New(runtimeid);
	nsmcs.BridgeBind("NODE_SMCS", tcpc.ReturnBridge());		// 将自己注册进tcpc中的桥
	nsmcs.BridgeBind("OUT_SMCS_DATA", outbridge);		// 注册进外传的桥
	
	rolesplus.StartBridge(nsmcs);			// 启动桥
	
	return nsmcs;
}

func (ns *NodeSmcs) Start () {
	go ns.GoMonitor();
}

// 返回外传信息的通讯桥
func (ns *NodeSmcs) ReturnOutBridge () *bridges.Bridge {
	return ns.outbridge;
}

// 设置发送的WorkSet
func (ns *NodeSmcs) SetSendWorkSet (s uint8) {
	ns.nodesend.WorkSet = s;
}

// 设置发送的Type
func (ns *NodeSmcs) SetSendType (s string) {
	ns.nodesend.Type = s;
}

// 设置发送的Name
func (ns *NodeSmcs) SetSendName (s string) {
	ns.nodesend.Name = s;
}

// 设置发送的Status
func (ns *NodeSmcs) SetSendStatus (s uint8) {
	ns.nodesend.Status = s;
}

// 设置发送的运行日志，逐条添加
func (ns *NodeSmcs) SetSendRunLog (s string) {
	ns.nodesend.RunLog = append(ns.nodesend.RunLog, s);
}

// 设置发送的错误日志，逐条添加
func (ns *NodeSmcs) SetSendErrLog (s string) {
	ns.nodesend.ErrLog = append(ns.nodesend.ErrLog, s);
}

// TcpClient要求的TcpReturn方法
func (ns *NodeSmcs) TcpReturn (key, id string, data []byte) {
	var databody CenterSend;
	e := nst.BytesGobStruct(data, &databody);
	if e != nil {
		ns.Logerr(fmt.Errorf("smcs: [NodeSmcs]TcpReturn: ",e));
		return;
	}
	
	bsend := bridges.BridgeData{
		Id: ns.runtimeid,
		Operate: "SmcsReturn",
		Data: databody,
	};
	ns.BridgeSend("OUT_SMCS_DATA", bsend);	//将从中心结点发送的数据转发给输出桥，接收方需要有smcsReturn方法
}

// 做监控，每分钟将信息发送出去
func (ns *NodeSmcs) GoMonitor () {
	for {
		//fmt.Println("node send:", ns.nodesend);
		bytes, err := nst.StructGobBytes(ns.nodesend);
		if err != nil { 
			ns.ReturnLog().ErrLog(err) ; 
			time.Sleep(SLEEP_TIME * time.Second);
			continue; 
		}
		err = ns.tcpc.Send(bytes);
		if err != nil {
			ns.ReturnLog().ErrLog(err) ;
			time.Sleep(SLEEP_TIME * time.Second);
			continue;
		}
		ns.nodesend.RunLog = make([]string, 0);
		ns.nodesend.ErrLog = make([]string, 0);
		time.Sleep(SLEEP_TIME * time.Second);
	}
}

// Logerr 做日志
func (ns *NodeSmcs) Logerr (err interface{}) {
	if err == nil { return };
	if ns.logs != nil {
		ns.logs.ErrLog(err);
	} else {
		fmt.Println(err);
	}
}
