// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 分布式角色控制机。
// Distributed Roles Control Machine.
package drcm

import (
	"sync"
	
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/bridges"
)

// 这是“操作者”，也就是具体对控制器进行操作的方法。可以对配置中所有的主控制器进行镜像访问。
type Operator struct {
	// 为利用通讯桥而继承RolePlus
	rolesplus.RolePlus
	// 与主控制器的网络连接，因为支持无差别镜像，所以用了切片
	controller				[]*nst.TcpClient
	// 配置内容，见operator.cfg的例子
	config					*cpool.Block
	// 内部通讯桥
	inside_bridge			*bridges.Bridge
	
	// 循环缓存的大小
	loopcache_len			int
	// 循环缓存计数，也就是循环缓存下一个更新点在哪里
	loopcache_count			int
	// 循环缓存影射，map的string为角色的id，int则是缓存的位置
	loopcache_map			map[string]loopCacheMap
	// 循环缓存，string记录的是loopcache_map中的string
	loopcache				[]string			
	
	// 日志
	logs					*ilogs.Logs
	// 读写锁
	lock					*sync.RWMutex
}

// string为数据的名称，如果是关系，则类似与_children之类的名字，如果是数据则是数据的名字。
type loopCacheMap map[string]interface{}
