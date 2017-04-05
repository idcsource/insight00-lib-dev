// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs2

import (
	"reflect"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 创建根节点，内部调用使用，不需要直接新建
func AddRootNode(f FloorInterface, config *cpool.Block, log *ilogs.Logs) (root *NodeTree) {
	root = &NodeTree{
		name:        "Index",
		mark:        "",
		config:      config,
		if_children: false,
		node_type:   NODE_IS_ROOT,
		children:    make(map[string]*NodeTree),
		floor:       reflect.ValueOf(f),
		log:         log,
	}
	return
}

// 增加一个普通的节点
func (nt *NodeTree) AddNode(name, mark string, f FloorInterface, config *cpool.Block) (child *NodeTree) {
	nt.if_children = true
	child = &NodeTree{
		name:        name,
		mark:        mark,
		config:      config,
		if_children: false,
		node_type:   NODE_IS_NOMAL,
		children:    make(map[string]*NodeTree),
		floor:       reflect.ValueOf(f),
		log:         nt.log,
	}
	nt.children[mark] = child
	return
}

// 增加一个节点Door
func (nt *NodeTree) AddDoor(name, mark string, f FloorDoorInterface, config *cpool.Block) (child *NodeTree) {
	nt.if_children = true
	child = &NodeTree{
		name:        name,
		mark:        mark,
		config:      config,
		if_children: false,
		node_type:   NODE_IS_DOOR,
		floor:       reflect.ValueOf(f),
		children:    make(map[string]*NodeTree),
		log:         nt.log,
	}
	nt.children[mark] = child
	return
}

// 增加一个处理静态文件的节点（静态文件节点只能做最后一个节点，所以不再提供节点树的返回）
func (nt *NodeTree) AddStatic(mark, path string) {
	nt.if_children = true
	child := &NodeTree{
		name:        mark,
		mark:        mark,
		if_children: false,
		node_type:   NODE_IS_STATIC,
		floor: reflect.ValueOf(&StaticFileFloor{
			path: path,
		}),
	}
	nt.children[mark] = child
}

// 增加一个空节点
func (nt *NodeTree) AddEmpty(name, mark string) (child *NodeTree) {
	nt.if_children = true
	child = &NodeTree{
		name:        name,
		mark:        mark,
		if_children: false,
		node_type:   NODE_IS_EMPTY,
		floor:       reflect.ValueOf(&EmptyFloor{}),
		children:    make(map[string]*NodeTree),
		log:         nt.log,
	}
	nt.children[mark] = child
	return
}
