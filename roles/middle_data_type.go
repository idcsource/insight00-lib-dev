// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package roles

// 角色的版本，包含版本号和Id
type RoleVersion struct {
	Version int
	Id      string
}

// 角色的关系，包含了父、子、朋友、上下文
type RoleRelation struct {
	Father   string             // 父角色（拓扑结构层面）
	Children []string           // 虚拟的子角色群，只保存键名
	Friends  map[string]Status  // 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	Contexts map[string]Context // 上下文
}

type RoleDataPoint struct {
	Point map[string]interface{}
}

// 角色的中期存储类型
type RoleMiddleData struct {
	Version        RoleVersion
	VersionChange  bool
	Relation       RoleRelation
	RelationChange bool
	Data           RoleDataPoint
	DataChange     bool
}
