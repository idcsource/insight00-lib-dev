// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package roles

import (
	"github.com/idcsource/insight00-lib/iendecode"
)

// 角色的版本，包含版本号和Id
type RoleVersion struct {
	Version int
	Id      string
}

func (r *RoleVersion) EncodeBinary() (b []byte, lens int64, err error) {
	version_b := iendecode.Uint64ToBytes(uint64(r.Version))
	id_b := []byte(r.Id)
	//b = append(version_b, id_b...)
	lens = int64(8 + len(id_b))
	b = make([]byte, lens)
	copy(b, version_b)
	copy(b[8:], id_b)
	return
}

func (r *RoleVersion) DecodeBinary(b []byte) (err error) {
	version_b := b[0:8]
	r.Version = int(iendecode.BytesToUint64(version_b))
	id_b := b[8:]
	r.Id = string(id_b)
	return
}

// 角色的关系，包含了父、子、朋友、上下文
type RoleRelation struct {
	Father   string             // 父角色（拓扑结构层面）
	Children []string           // 虚拟的子角色群，只保存键名
	Friends  map[string]Status  // 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	Contexts map[string]Context // 上下文
}

type RoleData struct {
	Point map[string]*RoleDataPoint
}

type RoleDataPoint struct {
	Type string
	Data interface{}
}

// 角色的中期存储类型
type RoleMiddleData struct {
	Version        RoleVersion
	VersionChange  bool
	Relation       RoleRelation
	RelationChange bool
	Data           RoleData
	DataChange     bool
}
