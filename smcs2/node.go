// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs2

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个为节点使用的配置蔓延
func NewNodeSmcs(name, center string, tcp *nst.TcpClient, outoperate NodeOperator) (ns *NodeSmcs, err error) {
	logs, err := ilogs.NewLog("", "", name+"_Node_SMCS_Log", true)
	if err != nil {
		err = fmt.Errorf("smcs2[NodeSmcs]NewNodeSmcs: %v", err)
		return
	}
	ns = &NodeSmcs{
		name:       name,
		centername: center,
		tcpc:       tcp,
		runtimeid:  random.GetRand(40),
		operate:    NODE_OPERATE_FUNCTION,
		outoperate: outoperate,
		sleeptime:  60,
		logs:       logs,
	}
	return
}
