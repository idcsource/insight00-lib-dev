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
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 新建一个中央蔓延点
func NewCenterSmcs(name string, store rolesio.RolesInOutManager) (center *CenterSmcs, err error) {
	center = &CenterSmcs{
		name:  name,
		node:  make(map[string]sendAndReceive),
		store: store,
	}
	root := &roles.Role{}
	root.New(random.GetSha1Sum(ROLE_PREFIX + ROLE_ROOT))
	center.root = root
	err = center.store.StoreRole(root)
	return
}

// 添加一个节点，types见ROLE_TYPE_*，group为所在分组
func (c *CenterSmcs) AddNode(name string, types uint8, group string) (err error) {
	group_id := random.GetSha1Sum(random.GetSha1Sum(ROLE_PREFIX + group))
	id := random.GetSha1Sum(ROLE_PREFIX + name)
	have, err := c.store.ExistChild(group_id, id)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	if have == true {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: The node name %v already exist in group.", name)
	}
	node := &NodeConfig{}
	node.New(id)
	node.Name = name
	node.RoleType = types
	node.ConfigStatus = CONFIG_NO
	node.SetFather(group_id)
	err = c.store.StoreRole(node)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	err = c.store.WriteChild(group_id, id)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]AddNode: %v", err)
	}
	return
}

// 设置节点的配置文件信息
func (c *CenterSmcs) SetNodeConfig(name string, config *cpool.ConfigPool) (err error) {
	id := random.GetSha1Sum(ROLE_PREFIX + name)
	poolEncode := config.EncodePool()
	err = c.store.WriteData(id, "Config", poolEncode)
	if err != nil {
		return fmt.Errorf("smcs[CenterSmcs]SetNodeConfig: %v", err)
	}
	return
}

// 获取节点的配置文件信息
func (c *CenterSmcs) GetNodeConfig(name string) (config *cpool.ConfigPool, err error) {
	id := random.GetSha1Sum(ROLE_PREFIX + name)
	poolEncode := cpool.PoolEncode{}
	err = c.store.ReadData(id, "Config", &poolEncode)
	if err != nil {
		return nil, fmt.Errorf("smcs[CenterSmcs]GetNodeConfig: %v", err)
	}
	config = cpool.NewConfigPoolNoFile()
	config.DecodePool(poolEncode)
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
