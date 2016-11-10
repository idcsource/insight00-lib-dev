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
	
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
)

// 常量区域

// 操作方式列表
const (
	// 什么操作都没有
	OPERATE_NOTHING					= iota
	// 强制保存
	OPREATE_TOSTORE
	
	// 创建一个新角色
	OPERATE_NEW_ROLE
	// 删除一个角色
	OPERATE_DEL_ROLE
	
	// 获取角色的一个值
	OPERATE_GET_DATA
	// 设置角色的一个值
	OPERATE_SET_DATA
	
	// 设置father
	OPERATE_SET_FATHER
	// 获取father
	OPERATE_GET_FATHER
	// 重置father
	OPERATE_RESET_FATHER
	
	// 设置children
	OPERATE_SET_CHILDREN
	// 重置children
	OPERATE_RESET_CHILDREN
	
	// 添加一个child
	OPERATE_ADD_CHILD
	// 删除一个child
	OPERATE_DEL_CHILD
	
	// 设置friends
	OPERATE_SET_FRIENDS
	// 重置friends
	OPERATE_RESET_FRIENDS
	
	// 添加一个friend
	OPERATE_ADD_FRIEND
	// 删除一个friend
	OPERATE_DEL_FRIEND
	// 修改一个friend
	OPERATE_CHANGE_FRIEND
	// 获取同样绑定值的friend
	OPERATE_SAME_BIND_FRIEND
	
	// 添加一个空的上下文组
	OPERATE_ADD_CONTEXT
	// 删除一个上下文组
	OPREATE_DEL_CONTEXT
	// 获取所有上下文的名称
	OPERATE_GET_CONTEXTS_NAME
	
	// 添加一个上文
	OPERATE_ADD_CONTEXT_UP
	// 删除上文
	OPREATE_DEL_CONTEXT_UP
	// 修改上文
	OPREATE_CHANGE_CONTEXT_UP
	// 返回同样绑定的上文
	OPREATE_SAME_BIND_CONTEXT_UP
	
	// 添加一个下文
	OPERATE_ADD_CONTEXT_DOWN
	// 删除下文
	OPREATE_DEL_CONTEXT_DOWN
	// 修改下文
	OPREATE_CHANGE_CONTEXT_DOWN
	// 返回同样绑定的下文
	OPREATE_SAME_BIND_CONTEXT_DOWN
	
	// 设置朋友的状态
	OPERATE_SET_FRIEND_STATUS
	// 获取朋友的状态
	OPERATE_GET_FRIEND_STATUS
	
	// 设置上下文的状态
	OPERATE_SET_CONTEXT_STATUS
	// 获取上下文的状态
	OPERATE_GET_CONTEXT_STATUS
)

// 分布式模式
const (
	// 没有分布式，只有自己
	DMODE_OWN					= iota
	// 在分布式里做master
	DMODE_MASTER
	// 在分布式里做slave
	DMODE_SLAVE
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


// 锆存储。
// 这是一个带缓存的存储体系，基本功能方面可以看作是hardstore与rcontrol的合并（虽然还有很大不同），而增强方面它支持分布式存储。
type ZrStorage struct {
	rolesplus.RolePlus
	
	/* 下面这部分是存储相关的 */
	
	// 配置信息
	config				*cpool.Block
	// 本地存储
	local_store			*hardstore.HardStore
	
	/* 下面这部分是缓存相关的 */
	
	// 角色缓存
	rolesCache				map[string]oneRoleCache
	// 最大缓存角色数
	cacheMax				int64
	// 缓存数量
	rolesCount				int64
	// 缓存满的触发
	cacheIsFull				chan bool
	// 删除缓存
	deleteCache				[]string
	// 检查缓存数量中
	checkCacheNumOn			bool
	
	/* 下面是分布式服务相关的 */
	
	// 分布式服务的模式，来自于常量DMODE_*
	dmode					uint8
	// 自身的身份码，做服务的时候使用
	code					string
	// 请求slave执行或返回数据的连接，string为slave对应的管理第一个值的首字母，而那个切片则是做镜像的
	slaves			map[string][]*slaveIn
	// 监听的实例
	listen			*nst.TcpServer
	
	// 日志
	logs					*ilogs.Logs
}

// 一个角色的缓存，提供了锁
type oneRoleCache struct {
	lock					*sync.RWMutex
	role					roles.Roleer
}

// 一台从机的信息
type slaveIn struct {
	name string;
	tcpconn *nst.TcpClient;
}

// 关系存储类型
type roleRelation struct {
	// 父角色
	Father string
	// 虚拟的子角色群，只保存键名
	Children []string
	// 朋友角色群
	Friends map[string]roles.Status
	// 上下文角色群
	Contexts map[string]roles.Context
}
