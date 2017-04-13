// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstore

import (
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// 存储器类型
type HardStore struct {
	*rolesio.NilReadWrite
	config        *cpool.Section
	local_path    string
	path_deep     int64
	version_name  string
	relation_name string
	body_name     string
	data_name     string
}

type RoleVersion struct {
	Version int
	Id      string
}

type RoleRelation struct {
	Father   string                  // 父角色（拓扑结构层面）
	Children []string                // 虚拟的子角色群，只保存键名
	Friends  map[string]roles.Status // 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	Contexts map[string]roles.Context
}

type RoleDataNormal struct {
	Time map[string]time.Time

	String map[string]string

	Bool map[string]bool

	Uint8  map[string]uint8
	Uint   map[string]uint
	Uint64 map[string]uint64

	Int8  map[string]int8
	Int   map[string]int
	Int64 map[string]int64

	Float32 map[string]float32
	Float64 map[string]float64

	Complex64  map[string]complex64
	Complex128 map[string]complex128
}

type RoleDataStringMap struct {
	String map[string]map[string]string

	Bool map[string]map[string]bool

	Uint8  map[string]map[string]uint8
	Uint   map[string]map[string]uint
	Uint64 map[string]map[string]uint64

	Int8  map[string]map[string]int8
	Int   map[string]map[string]int
	Int64 map[string]map[string]int64

	Float32 map[string]map[string]float32
	Float64 map[string]map[string]float64

	Complex64  map[string]map[string]complex64
	Complex128 map[string]map[string]complex128
}

type RoleDataSlice struct {
	String map[string][]string

	Bool map[string][]bool

	Uint8  map[string][]uint8
	Uint   map[string][]uint
	Uint64 map[string][]uint64

	Int8  map[string][]int8
	Int   map[string][]int
	Int64 map[string][]int64

	Float32 map[string][]float32
	Float64 map[string][]float64

	Complex64  map[string][]complex64
	Complex128 map[string][]complex128
}

type RoleMiddleData struct {
	Version   RoleVersion
	Relation  RoleRelation
	Normal    RoleDataNormal
	StringMap RoleDataStringMap
	Slice     RoleDataSlice
}
