// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// 节点树基本数据类型
type NodeTree struct {
	Name       string               // 节点的名称
	Mark       string               // 用来做路由的，也就是未来显示在连接上的地址
	ConfigAll  *cpool.ConfigPool    // 全部的配置文件
	Config     *cpool.Block         // 节点配置文件
	IfChildren bool                 // 是否有下层
	Type       int                  // 类型，首页、普通页、入口Door
	FloorValue reflect.Value        // 控制器的反射值
	Children   map[string]*NodeTree // 下层的信息，map的键为Mark
	Log        *ilogs.Logs
}

// 路由器基本类型
type Router struct {
	NodeTree *NodeTree                 // 节点树
	NotFound reflect.Value             // 404路由
	RouterOk bool                      // 其实就是看是否已经设定了NodeTree的根节点
	Static   map[string]*regexp.Regexp // 静态路由
	Config   *cpool.ConfigPool         // 配置文件
	Log      *ilogs.Logs
}

// 创建路由器，在NewWeb()中调用，不需要手动调用
func NewRouter(conf *cpool.ConfigPool, log *ilogs.Logs) (router *Router) {
	router = &Router{
		NodeTree: &NodeTree{},
		NotFound: reflect.ValueOf(&NotFoundFloor{}),
		RouterOk: false,
		Static:   make(map[string]*regexp.Regexp),
		Config:   conf,
		Log:      log,
	}
	return
}

// 修改默认的404处理
func (r *Router) SetNotFound(f FloorInterface) {
	r.NotFound = reflect.ValueOf(f)
}

// 开始NodeTree，输入根节点的信息
func (r *Router) BeginNodeTree(f FloorInterface, cf string) (nt *NodeTree, err error) {
	var config *cpool.Block
	if len(cf) != 0 {
		config, err = r.Config.GetBlock(cf)
		if err != nil {
			r.Log.ErrLog(fmt.Errorf("webs: [Router]BeginNodeTree: %v", err))
			return
		}
	}
	r.RouterOk = true
	r.NodeTree = &NodeTree{
		Name:       "Index",
		Mark:       "",
		ConfigAll:  r.Config,
		Config:     config,
		IfChildren: false,
		Type:       NODETREE_IS_ROOT,
		FloorValue: reflect.ValueOf(f),
		Children:   make(map[string]*NodeTree),
	}
	nt = r.NodeTree
	return
}

// 添加静态路由
func (r *Router) AddStatic(url, path string) {
	url = pubfunc.PathMustBegin(url)
	url = "^" + url + "/(.*)"
	tu, err := regexp.Compile(url)
	if err != nil {
		return
	}
	r.Static[path] = tu
	return
}

// 获取静态路由对应的地址
func (r *Router) GetStatic(mark string) (path string, have bool) {
	have = false
	if len(r.Static) != 0 {
		for k, v := range r.Static {
			if v.MatchString(mark) {
				nameA := v.FindStringSubmatch(mark)
				if len(nameA) > 1 {
					name := nameA[1]
					path = pubfunc.LocalFile("") + pubfunc.DirMustEnd(k) + name
					have = true
				}
				return
			}
		}
	}
	return
}

// 找到现在需要去运行的Floor
func (r *Router) GetRunFloor(rt Runtime) (runfloor reflect.Value, rtn Runtime) {
	var np *NodeTree
	var nothing bool
	runfloor = r.NotFound
	np, rtn, nothing = r.NodeTree.GetRunFloor(rt)
	// 如果最终返回的是一个Door
	if np.Type == NODETREE_IS_DOOR {
		floorlistv := np.FloorValue.MethodByName("FloorList").Call(nil)
		floorlist := floorlistv[0].Interface().(FloorDoor)
		var fname string
		if len(rtn.NowRoutePath) > 0 {
			fname = rtn.NowRoutePath[0]
		} else {
			fname = "/"
		}
		dfloor, fok := floorlist[fname]
		if fok == false {
			runfloor = r.NotFound
		} else {
			runfloor = reflect.ValueOf(dfloor)
			if len(rtn.NowRoutePath) > 0 {
				rtn.NowRoutePath = rtn.NowRoutePath[1:]
			}
		}
	} else {
		if nothing == true {
			runfloor = r.NotFound
		} else {
			runfloor = np.FloorValue
		}
	}
	return
}

// 增加一个普通的节点
func (nt *NodeTree) AddNode(name, mark string, f FloorInterface, cf string) (child *NodeTree, err error) {
	var config *cpool.Block
	if len(cf) != 0 {
		config, err = nt.ConfigAll.GetBlock(cf)
		if err != nil {
			nt.Log.ErrLog(fmt.Errorf("webs: [NodeTree]AddNode: %v", err))
			return
		}
	}
	nt.IfChildren = true
	child = &NodeTree{
		Name:       name,
		Mark:       mark,
		Config:     config,
		IfChildren: false,
		Type:       NODETREE_IS_NOMAL,
		FloorValue: reflect.ValueOf(f),
		Children:   make(map[string]*NodeTree),
	}
	nt.Children[mark] = child
	return
}

// 增加一个节点Door
func (nt *NodeTree) AddUnit(name, mark string, f FloorDoorInterface, cf string) (child *NodeTree, err error) {
	var config *cpool.Block
	if len(cf) != 0 {
		config, err = nt.ConfigAll.GetBlock(cf)
		if err != nil {
			nt.Log.ErrLog(fmt.Errorf("webs: [NodeTree]AddUnit: %v", err))
			return
		}
	}
	nt.IfChildren = true
	child = &NodeTree{
		Name:       name,
		Mark:       mark,
		Config:     config,
		IfChildren: false,
		Type:       NODETREE_IS_DOOR,
		FloorValue: reflect.ValueOf(f),
		Children:   make(map[string]*NodeTree),
	}
	nt.Children[mark] = child
	return
}

// 增加一个处理静态文件的节点
func (nt *NodeTree) AddStatic(mark, path string) {
	nt.IfChildren = true
	config := cpool.NewBlock("StaticFiles", "")
	config.SetConfig("main.path", path, "")
	child := &NodeTree{
		Name:       mark,
		Mark:       mark,
		IfChildren: false,
		Config:     config,
		Type:       NODETREE_IS_STATIC,
		FloorValue: reflect.ValueOf(&StaticFileFloor{}),
	}
	nt.Children[mark] = child
}

// 增加一个空节点
func (nt *NodeTree) AddEmpty(name, mark string) (child *NodeTree) {
	nt.IfChildren = true
	child = &NodeTree{
		Name:       name,
		Mark:       mark,
		IfChildren: false,
		Type:       NODETREE_IS_EMPTY,
		FloorValue: reflect.ValueOf(&EmptyFloor{}),
		Children:   make(map[string]*NodeTree),
	}
	nt.Children[mark] = child
	return
}

// （此方法不需要手动调用）找到现在需要去运行的Floor，如果找到了则返回的NodeTree为找到的NodeTree，如果找不到，则返回的是能找到的最后一个的NodeTree
func (nt *NodeTree) GetRunFloor(rt Runtime) (nnt *NodeTree, rtn Runtime, nothing bool) {
	nothing = false
	var ntok bool
	nnt, ntok = nt.Children[rt.NowRoutePath[0]]
	if ntok == false {
		// 如果找不到，就启动nothing，最终也就是走404
		nothing = true
		rtn = rt
		nnt = nt
		//rtn.ConfigTree[nnt.Mark] = nnt.Config;
		rtn.MyConfig = nnt.Config
	} else {
		rt.NowRoutePath = rt.NowRoutePath[1:]      //先将总的请求减去一个
		rt.RealNode = rt.RealNode + "/" + nnt.Mark //将RealNode修改掉
		nothing = false
		if nnt.IfChildren == false || len(rt.NowRoutePath) == 0 {
			//如果新的节点没有子节点存在，或者NowRoutePath已经为空，则不再继续下面的操作
			rtn = rt
			//rtn.ConfigTree[nnt.Mark] = nnt.Config;
			rtn.MyConfig = nnt.Config
			return
		} else {
			nnt, rtn, nothing = nnt.GetRunFloor(rt)
			return
		}
	}
	return
}
