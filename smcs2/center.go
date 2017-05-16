// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs2

import (
	"fmt"
	"strings"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 新建一个中央蔓延点，这里的name也将作为配置节点的名称前缀
func NewCenterSmcs(name, area string, store *trule.TRule) (center *CenterSmcs, err error) {
	if store.AreaExist(area) == false {
		err = fmt.Errorf("Please check local store set.")
		return
	}
	center = &CenterSmcs{
		name:  name,
		node:  make(map[string]*sendAndReceive),
		store: store,
		area:  area,
	}
	root_id := name + "_" + ROLE_ROOT
	center.root_id = root_id
	// 查看是否存在这个root
	have := center.store.ExistRole(area, root_id)
	if have == false {
		newroot := &NodeConfig{
			Disname:  "Root Point",
			Code:     "Root Point",
			Name:     ROLE_ROOT,
			RoleType: ROLE_TYPE_GROUP,
		}
		newroot.New(root_id)
		err = center.store.StoreRole(area, newroot)
		if err != nil {
			return nil, err
		}
		center.root = newroot
	} else {
		newroot := &NodeConfig{}
		err = center.store.ReadRole(area, root_id, newroot)
		center.root = newroot
	}
	return center, nil
}

// 添加一个节点，types见ROLE_TYPE_*，group为所在分组
func (c *CenterSmcs) AddNode(nodename, disname, code string, types uint8, groupname string) (err error) {
	var group_id string
	if groupname == "" {
		// 根的名字
		group_id = c.root_id
	} else {
		group_id = c.name + "_" + groupname
	}
	node_id := c.name + "_" + nodename
	tran, _ := c.store.Begin()
	haverole := tran.ExistRole(c.area, node_id)
	if haverole == true {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]AddNode: The Node " + nodename + " is exist.")
	}
	have, err := tran.ExistChild(c.area, group_id, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	if have == true {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]AddNode: The node name %v already exist in group.", nodename)
	}
	node := &NodeConfig{}
	node.New(node_id)
	node.Name = nodename
	node.Code = code
	node.Disname = disname
	node.RoleType = types
	node.ConfigStatus = CONFIG_STATUS_NO
	node.SetFather(group_id)
	err = tran.StoreRole(c.area, node)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	err = tran.WriteChild(c.area, group_id, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	tran.Commit()
	return
}

// 删除一个节点，如果这个节点有子节点则不允许删除
func (c *CenterSmcs) DelNode(nodename string) (err error) {
	node_id := c.name + "_" + nodename
	tran, _ := c.store.Begin()
	// 获取是否有子节点
	children, err := tran.ReadChildren(c.area, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]DelNode: %v", err)
	}
	if len(children) != 0 {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]DelNode: The Node %v has child node.", nodename)
	}
	// 获取father
	father, err := tran.ReadFather(c.area, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]DelNode: %v", err)
	}
	err = tran.DeleteChild(c.area, father, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]DelNode: %v", err)
	}
	err = tran.DeleteRole(c.area, node_id)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]DelNode: %v", err)
	}
	tran.Commit()
	return
}

// 设置节点的配置信息
func (c *CenterSmcs) SetNodeConfig(nodename string, config *cpool.ConfigPool) (err error) {
	node_id := c.name + "_" + nodename
	poolEncode := config.EncodePool()
	tran, _ := c.store.Begin()
	err = tran.WriteData(c.area, node_id, "Config", poolEncode)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	err = tran.WriteData(c.area, node_id, "ConfigStatus", CONFIG_STATUS_NOT_READY)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	err = tran.WriteData(c.area, node_id, "NewConfig", true)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	tran.Commit()
	return
}

// 设置节点的配置信息（编码模式）
func (c *CenterSmcs) SetNodeConfigEncode(nodename string, config *cpool.PoolEncode) (err error) {
	node_id := c.name + "_" + nodename
	tran, _ := c.store.Begin()
	err = tran.WriteData(c.area, node_id, "Config", config)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigCode: %v", err)
	}
	err = tran.WriteData(c.area, node_id, "ConfigStatusCode", CONFIG_STATUS_NOT_READY)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigCode: %v", err)
	}
	err = tran.WriteData(c.area, node_id, "NewConfig", true)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigCode: %v", err)
	}
	tran.Commit()
	return
}

// 获取节点的配置信息
func (c *CenterSmcs) GetNodeConfig(nodename string) (config *cpool.ConfigPool, err error) {
	node_id := c.name + "_" + nodename
	poolEncode := cpool.PoolEncode{}
	err = c.store.ReadData(c.area, node_id, "Config", &poolEncode)
	if err != nil {
		return nil, fmt.Errorf("smcs[CenterSmcs]GetNodeConfig: %v", err)
	}
	config = cpool.NewConfigPoolNoFile()
	config.DecodePool(poolEncode)
	return
}

// 获取节点的配置信息（编码模式）
func (c *CenterSmcs) GetNodeConfigEncode(nodename string) (config cpool.PoolEncode, err error) {
	node_id := c.name + "_" + nodename
	config = cpool.PoolEncode{}
	err = c.store.ReadData(c.area, node_id, "Config", &config)
	if err != nil {
		return config, fmt.Errorf("smcs[CenterSmcs]GetNodeConfigCode: %v", err)
	}
	return
}

// 设置节点下一个工作状态，workset为WORK_SET_*
func (c *CenterSmcs) SetNodeWorkSet(nodename string, workset WorkSet) (err error) {
	node_id := c.name + "_" + nodename
	err = c.store.WriteData(c.area, node_id, "NextWorkSet", workset)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeWorkSet: %v", err)
	}
	return
}

// 获取节点的下一个工作状态
func (c *CenterSmcs) GetNodeWorkSet(nodename string) (workset WorkSet, err error) {
	node_id := c.name + "_" + nodename
	err = c.store.ReadData(c.area, node_id, "NextWorkSet", &workset)
	if err != nil {
		return 0, fmt.Errorf("smcs[CenterSmcs]GetNodeWorkSet: %v", err)
	}
	return
}

// 设置节点的配置状态,status为CONFIG_*，也就是如果被设置成CONFIG_ALL_READY，则系统将在下次发送更新的配置。这个设置一定要是所有配置信息以及工作状态都调整好之后再执行。
func (c *CenterSmcs) SetNodeConfigStatus(nodename string, status ConfigStatus) (err error) {
	node_id := c.name + "_" + nodename
	err = c.store.WriteData(c.area, node_id, "ConfigStatus", status)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigStatus: %v", err)
	}
	return
}

// 获取节点的配置状态
func (c *CenterSmcs) GetNodeConfigStatus(nodename string) (status ConfigStatus, err error) {
	node_id := c.name + "_" + nodename
	err = c.store.ReadData(c.area, node_id, "ConfigStatus", &status)
	if err != nil {
		return 0, fmt.Errorf("smcs[CenterSmcs]GetNodeConfigStatus: %v", err)
	}
	return
}

// 获取节点所有的错误日志
func (c *CenterSmcs) GetNodeErrLog(nodename string) (errlog []string, err error) {
	node_id := c.name + "_" + nodename
	err = c.store.ReadData(c.area, node_id, "ErrLog", &errlog)
	if err != nil {
		err = fmt.Errorf("smcs[CenterSmcs]GetNodeErrLog: %v", err)
	}
	return
}

// 获取节点所有的运行日志
func (c *CenterSmcs) GetNodeRunLog(nodename string) (runlog []string, err error) {
	node_id := c.name + "_" + nodename
	err = c.store.ReadData(c.area, node_id, "RunLog", &runlog)
	if err != nil {
		err = fmt.Errorf("smcs[CenterSmcs]GetNodeRunLog: %v", err)
	}
	return
}

// 清空节点的所有错误日志
func (c *CenterSmcs) EmptyNodeErrLog(nodename string) (err error) {
	node_id := c.name + "_" + nodename
	errlog := make([]string, 0)
	err = c.store.WriteData(c.area, node_id, "ErrLog", errlog)
	if err != nil {
		err = fmt.Errorf("smcs[CenterSmcs]EmptyNodeErrLog %v", err)
	}
	return
}

// 清空节点的所有运行日志
func (c *CenterSmcs) EmptyNodeRunLog(nodename string) (err error) {
	node_id := c.name + "_" + nodename
	runlog := make([]string, 0)
	err = c.store.WriteData(c.area, node_id, "RunLog", runlog)
	if err != nil {
		err = fmt.Errorf("smcs[CenterSmcs]EmptyNodeRunLog: %v", err)
	}
	return
}

// 获取节点状态，返回的是上次记录的时间到目前的间隔
func (c *CenterSmcs) GetNodeRunStatus(node_id string) (leave int64, workstatus WorkSet, err error) {
	node, find := c.node[node_id]
	if find == false {
		err = fmt.Errorf("smcs2[CenterSmcs]GetNodeRunStatus: Can not find the node.")
		return
	}
	leave = time.Now().Unix() - node.nodetime.Unix()
	workstatus = node.nodeSend.WorkSet
	return
}

// 返回节点树
func (c *CenterSmcs) GetNodeTree() (nodetree NodeTree, err error) {
	nodetree, err = c.getNodeTree(c.root_id)
	if err != nil {
		err = fmt.Errorf("smcs[CenterSmcs]GetNodeTree: %v", err)
	}
	return
}

// 返回节点树的内部使用
func (c *CenterSmcs) getNodeTree(node_id string) (nodetree NodeTree, err error) {
	children, err := c.store.ReadChildren(c.area, node_id)
	if err != nil {
		return
	}
	var name string
	err = c.store.ReadData(c.area, node_id, "Name", &name)
	if err != nil {
		return
	}
	var disname string
	err = c.store.ReadData(c.area, node_id, "Disname", &disname)
	if err != nil {
		return
	}
	var roletype uint8
	err = c.store.ReadData(c.area, node_id, "RoleType", &roletype)
	if err != nil {
		return
	}
	lifetime, workstatus, err := c.GetNodeRunStatus(node_id)
	var alive bool
	if err != nil {
		alive = false
		err = nil
	} else if lifetime <= 60 {
		alive = true
	} else {
		alive = false
	}
	nodetree = NodeTree{
		Name:     name,
		Disname:  disname,
		Id:       node_id,
		RoleType: roletype,
		Alive:    alive,
		Working:  workstatus,
		Tree:     make(map[string]NodeTree),
	}

	errall := make([]string, 0)
	for _, child := range children {
		nodetree.Tree[child], err = c.getNodeTree(child)
		if err != nil {
			errall = append(errall, err.Error())
		}
	}
	if len(errall) != 0 {
		errstr := strings.Join(errall, " | ")
		err = fmt.Errorf(errstr)
	}
	return
}

// nst的TcpServer接口实现
//
// 首先接收一段NodeSend
func (c *CenterSmcs) ExecTCP(ce *nst.ConnExec) (err error) {
	node_send_b, err := ce.GetData()
	if err != nil {
		return
	}
	node_send := NodeSend{}
	err = nst.BytesGobStruct(node_send_b, &node_send)
	if err != nil {
		return
	}
	// 查看是不是发给这个节点的
	if node_send.CenterName != c.name {
		c.sendError(ce, "The CenterName is wrong.")
		return
	}
	// 开始找寻有没有这个节点
	node_id := c.name + "_" + node_send.Name
	have := c.store.ExistRole(c.area, node_id)
	if have == false {
		c.sendError(ce, "Can't found the Node set: "+node_send.Name)
		return
	}
	var code string
	err = c.store.ReadData(c.area, node_id, "Code", &code)
	if err != nil {
		c.sendError(ce, err.Error())
		return
	}
	if code != node_send.Code {
		c.sendError(ce, "Node code wrong.")
		return
	}
	// 在c.node里找到，找不到就新建
	_, find := c.node[node_id]
	if find == false {
		c.node[node_id] = &sendAndReceive{
			cansend: false,
		}
	}
	c.node[node_id].nodeSend = node_send
	c.node[node_id].nodetime = time.Now()
	// 开启事务
	tran, _ := c.store.Begin()
	// 更新错误日志和运行日志
	log := make([]string, 0)
	if len(node_send.RunLog) != 0 {
		err = tran.ReadData(c.area, node_id, "RunLog", &log)
		if err != nil {
			tran.Rollback()
			c.sendError(ce, "Update node log error.")
			return
		}
		log = append(log, node_send.RunLog...)
		err = tran.WriteData(c.area, node_id, "RunLog", log)
		if err != nil {
			tran.Rollback()
			c.sendError(ce, "Update node log error.")
			return
		}
	}
	if len(node_send.ErrLog) != 0 {
		err = tran.ReadData(c.area, node_id, "ErrLog", &log)
		if err != nil {
			tran.Rollback()
			c.sendError(ce, "Update node log error.")
			return
		}
		log = append(log, node_send.ErrLog...)
		err = tran.WriteData(c.area, node_id, "ErrLog", log)
		if err != nil {
			tran.Rollback()
			c.sendError(ce, "Update node log error.")
			return
		}
	}
	// 执行事务
	tran.Commit()
	// 开启事务
	tran, _ = c.store.Begin()
	// 构建发送机制
	center_send := CenterSend{}
	err = tran.ReadData(c.area, node_id, "NextWorkSet", &center_send.NextWorkSet)
	if err != nil {
		tran.Rollback()
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	err = tran.ReadData(c.area, node_id, "ConfigStatus", &center_send.ConfigStatus)
	if err != nil {
		tran.Rollback()
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	err = tran.ReadData(c.area, node_id, "Config", &center_send.Config)
	if err != nil {
		tran.Rollback()
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	err = tran.ReadData(c.area, node_id, "NewConfig", &center_send.NewConfig)
	if err != nil {
		tran.Rollback()
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	err = tran.WriteData(c.area, node_id, "NewConfig", false)
	if err != nil {
		tran.Rollback()
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	// 执行事务
	tran.Commit()
	// 编码发送
	center_send_b, err := nst.StructGobBytes(center_send)
	if err != nil {
		c.sendError(ce, "Build CenterSend error.")
		return
	}
	err = ce.SendData(center_send_b)
	return
}

func (c *CenterSmcs) sendError(ce *nst.ConnExec, err string) {
	// 构造发送出去的结构体
	center_send := CenterSend{}
	center_send.Error = err
	//编码
	center_send_b, _ := nst.StructGobBytes(center_send)
	// 发送错误
	ce.SendData(center_send_b)
	return
}
