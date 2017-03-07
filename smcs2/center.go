// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package smcs2

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 新建一个中央蔓延点，这里的name也将作为配置节点的名称前缀
func NewCenterSmcs(name string, store rolesio.RolesInOutManager) (center *CenterSmcs, err error) {
	center = &CenterSmcs{
		name:  name,
		node:  make(map[string]sendAndReceive),
		store: store,
	}
	root_id := name + "_" + ROLE_ROOT
	// 查看是否存在这个root
	root, err := center.store.ReadRole(root_id)
	if err != nil {
		newroot := &roles.Role{}
		newroot.New(root_id)
		err = center.store.StoreRole(newroot)
		if err != nil {
			return nil, err
		}
		center.root = newroot
	} else {
		center.root = root
	}
	return center, nil
}

// 添加一个节点，types见ROLE_TYPE_*，group为所在分组
func (c *CenterSmcs) AddNode(nodename string, types uint8, groupname string) (err error) {
	var group_id string
	if groupname == "" {
		// 根的名字
		group_id = c.name + "_" + ROLE_ROOT
	} else {
		group_id = c.name + "_" + groupname
	}
	node_id := c.name + "_" + nodename
	have, err := c.store.ExistChild(group_id, node_id)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	if have == true {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: The node name %v already exist in group.", nodename)
	}
	node := &NodeConfig{}
	node.New(node_id)
	node.Name = nodename
	node.RoleType = types
	node.ConfigStatus = CONFIG_NO
	node.SetFather(group_id)
	err = c.store.StoreRole(node)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	err = c.store.WriteChild(group_id, node_id)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	return
}

// 设置节点的配置信息
func (c *CenterSmcs) SetNodeConfig(nodename string, config *cpool.ConfigPool) (err error) {
	node_id := c.name + "_" + nodename
	poolEncode := config.EncodePool()
	err = c.store.WriteData(node_id, "Config", poolEncode)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	err = c.store.WriteData(node_id, "ConfigStatus", uint8(CONFIG_NOT_READY))
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	return
}

// 设置节点的配置信息（编码模式）
func (c *CenterSmcs) SetNodeConfigEncode(nodename string, config *cpool.PoolEncode) (err error) {
	node_id := c.name + "_" + nodename
	err = c.store.WriteData(node_id, "Config", config)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigCode: %v", err)
	}
	err = c.store.WriteData(node_id, "ConfigStatusCode", uint8(CONFIG_NOT_READY))
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfigCode: %v", err)
	}
	return
}

// 获取节点的配置信息
func (c *CenterSmcs) GetNodeConfig(nodename string) (config *cpool.ConfigPool, err error) {
	node_id := c.name + "_" + nodename
	poolEncode := cpool.PoolEncode{}
	err = c.store.ReadData(node_id, "Config", &poolEncode)
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
	err = c.store.ReadData(node_id, "Config", &config)
	if err != nil {
		return config, fmt.Errorf("smcs[CenterSmcs]GetNodeConfigCode: %v", err)
	}
	return
}

// 设置的运行时保存
func (c *CenterSmcs) ToStore() (err error) {
	return c.store.ToStore()
}

// nst的TcpServer接口实现
func (c *CenterSmcs) ExecTCP(ce *nst.ConnExec) (err error) {
	return
}
