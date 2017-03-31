// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs2

import (
	"fmt"
	"reflect"
	"time"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个为节点使用的配置蔓延
func NewNodeSmcs(name, center string, tcp *nst.TcpClient, logs *ilogs.Logs) (ns *NodeSmcs, err error) {
	logsn, err := ilogs.NewLogForSmcs(name + "_Node_SMCS_Log")
	if err != nil {
		err = fmt.Errorf("smcs2[NodeSmcs]NewNodeSmcs: %v", err)
		return
	}
	ns = &NodeSmcs{
		name:       name,
		centername: center,
		tcpc:       tcp,
		runtimeid:  random.GetRand(40),
		//operate:    NODE_OPERATE_FUNCTION,
		//outoperate: reflect.ValueOf(outoperate),
		nodesend: &NodeSend{
			CenterName: center,
			Name:       name,
			Status:     NODE_STATUS_NO_CONFIG,
			WorkSet:    WORK_SET_NO,
			RunLog:     make([]string, 0),
			ErrLog:     make([]string, 0),
		},
		closeM:    make(chan bool),
		closeMt:   false,
		sleeptime: 60,
		logn:      logsn,
		logs:      logs,
	}
	return
}

// 注册内联操作
func (ns *NodeSmcs) RegOperate(outoperate NodeOperator) {
	ns.operate = NODE_OPERATE_FUNCTION
	ns.outoperate = reflect.ValueOf(outoperate)
}

// 关闭状态监控
func (ns *NodeSmcs) Close() {
	if ns.closeMt != true {
		ns.closeM <- true
	}
}

// 重启状态监控
func (ns *NodeSmcs) Start() {
	if ns.closeMt == true {
		go ns.goMonitor()
	}
}

// 内部的计时发送函数
func (ns *NodeSmcs) goMonitor() {
	ns.closeMt = false
	for {
		select {
		case <-ns.closeM:
			ns.closeMt = true
			return
		default:
			// 这里就是处理发送的
			err := ns.sendNodeSend()
			if err != nil {
				ns.logerr(err)
			}
			time.Sleep(time.Duration(ns.sleeptime) * time.Second)
		}
	}
}

// 发送NodeSend
func (ns *NodeSmcs) sendNodeSend() (err error) {
	ns.nodesend.RunLog = ns.logn.ReRunLog()
	ns.nodesend.ErrLog = ns.logn.ReErrLog()
	// 编码
	node_send_b, err := nst.StructGobBytes(ns.nodesend)
	if err != nil {
		return err
	}
	// 分配连接
	cprocess := ns.tcpc.OpenProgress()
	defer cprocess.Close()
	// 发送nodesend
	return_b, err := cprocess.SendAndReturn(node_send_b)
	if err != nil {
		return err
	}
	// 解码CenterSend
	center_send := CenterSend{}
	err = nst.BytesGobStruct(return_b, &center_send)
	if err != nil {
		return err
	}
	if center_send.Error != "" {
		return fmt.Errorf(center_send.Error)
	}
	// 发送给NodeOperator接口
	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(center_send)
	rerror := ns.outoperate.MethodByName("SmcsNodeOperator").Call(in)
	erra := rerror[0].Interface()
	if erra != nil {
		err = erra.(error)
	}
	return
}

// 更改等待间隔
func (ns *NodeSmcs) ChangeSleepTime(t int64) {
	ns.sleeptime = t
}

// 返回设置操作者
func (ns *NodeSmcs) Operator() (operator *NodeConfigOperator) {
	operator = &NodeConfigOperator{
		nodesend: ns.nodesend,
		logn:     ns.logn,
	}
	return
}

// Logerr 做日志
func (ns *NodeSmcs) logerr(err interface{}) {
	if err == nil {
		return
	}
	if ns.logs != nil {
		ns.logs.ErrLog(err)
	} else {
		fmt.Println(err)
	}
}

// 更改状态
func (n *NodeConfigOperator) ChangeStatus(status uint8) {
	n.nodesend.Status = status
}

// 更改工作设置
func (n *NodeConfigOperator) ChangeWorkSet(workset uint8) {
	n.nodesend.WorkSet = workset
}

// 追加错误日志
func (n *NodeConfigOperator) ErrLog(err ...interface{}) {
	n.logn.ErrLog(err...)
}

// 追加运行日志
func (n *NodeConfigOperator) RunLog(err ...interface{}) {
	n.logn.RunLog(err...)
}
