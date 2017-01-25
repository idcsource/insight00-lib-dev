// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 实现了nst.ConnExecer的接口。可以接收Tcp Server所传回的连接。
type CenterSmcs struct {
	tonode      map[string]CenterSend // string为nodetype-name，将要发送给结点的信息
	getnode     map[string]NodeSend   // string为nodetype-name，从结点受到的信息
	config      *cpool.ConfigPool     // 结点的配置文件
	allconfnode []string              // 所有被配置文件管理的结点的名字
	logs        *ilogs.Logs
}

// 建立一个中心管理结点使用的配置蔓延
func NewCenterSmcs(config *cpool.ConfigPool, logs *ilogs.Logs) *CenterSmcs {
	return &CenterSmcs{
		tonode:      make(map[string]CenterSend),
		getnode:     make(map[string]NodeSend),
		allconfnode: make([]string, 0),
		config:      config,
		logs:        logs,
	}
}

// 设置某个结点的配置
func (cs *CenterSmcs) SetNodeStatus(typen, name string, opt CenterSend) {
	nodename := typen + "-" + name
	cs.tonode[nodename] = opt
}

// 获取设置了的结点的配置
func (cs *CenterSmcs) GetSetNodeStatus(node string) (ns CenterSend, err error) {
	var find bool
	ns, find = cs.tonode[node]
	if find == false {
		err = fmt.Errorf("smcs center: %v", node, " node not be found. ")
	}
	return
}

// 返回所有配置文件中的结点
func (cs *CenterSmcs) GetAllConfNode() []string {
	return cs.config.GetAllBlockName()
}

// 返回所有收到信息的结点名字
func (cs *CenterSmcs) BeGetNodeName() []string {
	re := make([]string, 0)
	for name, _ := range cs.getnode {
		re = append(re, name)
	}
	return re
}

// 返回一个结点发送来的信息
func (cs *CenterSmcs) GetNodeStatus(node string) (ns NodeSend, err error) {
	var find bool
	ns, find = cs.getnode[node]
	if find == false {
		err = fmt.Errorf("smcs center: %v", node, " node not be found. ")
	}
	return
}

// nst.ConnExecer接口的实现
func (cs *CenterSmcs) ExecTCP(conn_exec *nst.ConnExec) error {
	tcp := conn_exec.Tcp()
	var pullstruct NodeSend
	err1 := tcp.GetStruct(&pullstruct)
	if err1 != nil {
		cs.logerr(fmt.Errorf("smcs [CenterSmcs]ExecTCP: %v", err1))
		return err1
	}
	nodename := pullstruct.Type + "-" + pullstruct.Name
	cs.getnode[nodename] = pullstruct
	push, ok := cs.tonode[nodename]
	fmt.Println("center要发送：", nodename)
	if ok == false {
		newblock := cpool.NewBlock(nodename, "")
		blockencode := newblock.EncodeBlock()
		push = CenterSend{
			NextWorkSet:  WORK_SET_NO,
			SetStartTime: 0,
			NewConfig:    false,
			Config:       blockencode,
		}
	}
	err2 := tcp.SendStruct(push)
	fmt.Println("center发送完：", nodename)
	if err2 != nil {
		cs.logerr(fmt.Errorf("smcs [CenterSmcs]ExecTCP: ", err2))
		return nil
	}
	return nil
}

func (cs *CenterSmcs) logerr(err interface{}) {
	if err == nil {
		return
	}
	if cs.logs != nil {
		cs.logs.ErrLog(err)
	} else {
		fmt.Println(err)
	}
}
