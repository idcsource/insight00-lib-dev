// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs2

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst2"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个为节点使用的配置蔓延
func NewNodeSmcs(center string, tcp *nst2.Client) (ns *NodeSmcs, err error) {
	ns = &NodeSmcs{
		centername: center,
		tcpc:       tcp,
		runtimeid:  random.GetRand(40),
		sleeptime:  60,
	}
	return
}

// 发送NodeSend
func (ns *NodeSmcs) SendNodeSend(node_send NodeSend) (center_send CenterSend, err error) {
	// 编码
	node_send.CenterName = ns.centername
	node_send_b, err := iendecode.StructGobBytes(node_send)
	if err != nil {
		return
	}
	// 分配连接
	cprocess, err := ns.tcpc.OpenProgress()
	if err != nil {
		err = fmt.Errorf("smcs2[NodeSmcs]SendNodeSend: %v", err)
		return
	}
	defer cprocess.Close()
	// 发送nodesend
	return_b, err := cprocess.SendAndReturn(node_send_b)
	if err != nil {
		return
	}
	// 解码CenterSend
	center_send = CenterSend{}
	err = iendecode.BytesGobStruct(return_b, &center_send)
	if err != nil {
		return
	}
	if center_send.Error != "" {
		err = fmt.Errorf(center_send.Error)
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
