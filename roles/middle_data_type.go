// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package roles

import (
	"time"
)

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

// 角色的数据点类型（普通）
type RoleDataNormal struct {
	Time map[string]time.Time

	Byte map[string][]byte

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

// 角色的数据点类型（Map）
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

// 角色的数据点类型（切片）
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

// 角色的中期存储类型
type RoleMiddleData struct {
	Version   RoleVersion
	Relation  RoleRelation
	Normal    RoleDataNormal
	StringMap RoleDataStringMap
	Slice     RoleDataSlice
}
