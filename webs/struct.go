// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// HTTP服务器实现。
//
// TODO: 此部分的文档还十分不完整，待慢慢完善
package webs

import (
	"github.com/idcsource/Insight-0-0-lib/idb"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

const (
	// 默认最大并发
	MAX_ROUTINE_RATIO			= 500
)

const (	
	//节点树节点的类型
	NODETREE_IS_ROOT			= iota
	NODETREE_IS_DOOR
	NODETREE_IS_NOMAL
	NODETREE_IS_STATIC
	NODETREE_IS_EMPTY
)

// Web的数据结构
type Web struct {
	Local				string						//本地路径
	Config				*cpool.ConfigPool			//主配置文件
	SelfConfig			*cpool.Block				//站点自身的配置文件
	Database			*idb.DB						//主数据库连接
	MultiDB				map[string]*idb.DB			//扩展多数据库准备
	Ext					map[string]interface{}		//Extension扩展数据（功能）
	Router				*Router						//路由器
	Log					*ilogs.Logs					//运行日志
	MaxRoutine			chan bool					//最大并发
}

// 运行时数据结构
type Runtime struct {
	AllRoutePath		string						//整个的RoutePath，也就是除域名外的完整路径
	NowRoutePath		[]string					//AllRoutePath经过层级路由之后剩余的部分
	RealNode			string						//当前节点的树名，如/node1/node2，如果没有使用节点则此处为空
	MyConfig			*cpool.Block  				//当前节点的配置文件，从ConfigTree中获取，如当前节点没有配置文件，则去寻找父节点，直到载入站点的配置文件
	UrlRequest			map[string]string			//Url请求的整理，风格为:id=1/:type=notype
}
