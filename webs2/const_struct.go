// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Web服务器的第二个版本，一个HTTP服务器的实现
package webs2

import (
	"net/http"
	"reflect"
	"regexp"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/idb"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

const (
	// 默认最大并发
	MAX_ROUTINE_RATIO = 500
)

const (
	//节点的类型
	NODE_IS_NO     = iota // 没有定义
	NODE_IS_ROOT          // 根结点
	NODE_IS_DOOR          // 入口Door
	NODE_IS_NOMAL         // 普通
	NODE_IS_STATIC        // 静态
	NODE_IS_EMPTY         // 空
)

// Web的数据结构
type Web struct {
	local       string                 // 本地路径
	static      string                 // 静态资源路径
	config      *cpool.Section         // 自身的配置文件
	database    *idb.DB                // 主数据库连接
	multiDB     map[string]*idb.DB     // 扩展多数据库准备
	ext         map[string]interface{} // Extension扩展数据（功能）
	router      *Router                // 路由器
	log         *ilogs.Logs            // 运行日志
	max_routine chan bool              // 最大并发
}

// 路由器基本类型
type Router struct {
	node_tree    *NodeTree                 // 节点树
	not_found    reflect.Value             // 404路由
	router_ok    bool                      // 其实就是看是否已经设定了NodeTree的根节点
	static_route map[string]*regexp.Regexp // 静态路由
}

// 节点树基本数据类型
type NodeTree struct {
	name        string               // 节点的名称
	mark        string               // 用来做路由的，也就是未来显示在连接上的地址
	config      *cpool.Block         // 节点配置文件
	if_children bool                 // 是否有下层
	node_type   int                  // 类型，首页、普通页、入口Door，NODE_IS_*
	floor       reflect.Value        // 控制器的反射值
	children    map[string]*NodeTree // 下层的信息，map的键为Mark
}

// 运行时数据结构
type Runtime struct {
	AllRoutePath string            //整个的RoutePath，也就是除域名外的完整路径
	NowRoutePath []string          //AllRoutePath经过层级路由之后剩余的部分
	RealNode     string            //当前节点的树名，如/node1/node2，如果没有使用节点则此处为空
	MyConfig     *cpool.Block      //当前节点的配置文件，从ConfigTree中获取，如当前节点没有配置文件，则去寻找父节点，直到载入站点的配置文件
	UrlRequest   map[string]string //Url请求的整理，风格为:id=1/:type=notype
	Log          *ilogs.Logs       // 日志
}

// FloorInterface 此为控制器接口的定义
type FloorInterface interface {
	InitHTTP(w http.ResponseWriter, r *http.Request, b *Web, rt Runtime)
	ExecHTTP()
}

//控制器原型的数据类型
type Floor struct {
	W  http.ResponseWriter
	R  *http.Request
	Rt Runtime
	B  *Web
}

// FloorDoor的接口和数据类型
type FloorDoor map[string]FloorInterface

type FloorDoorInterface interface {
	FloorList() FloorDoor
}
