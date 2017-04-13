// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstore

import (
	"fmt"
	"reflect"
	"time"

	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 编码角色，将角色编码为中期存储格式
func EncodeRoleToMiddle(role roles.Roleer) (mid RoleMiddleData, err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]EncodeRoleToMiddle: %v", e)
		}
	}()

	mid = RoleMiddleData{}

	// 这里是生成relation
	mid.Relation = RoleRelation{
		Father:   role.GetFather(),
		Children: role.GetChildren(),
		Friends:  role.GetFriends(),
		Contexts: role.GetContexts(),
	}
	// 这里是生成Version
	mid.Version = RoleVersion{
		Version: role.Version(),
		Id:      role.ReturnId(),
	}

	// 这里是开始准备生成数据
	mid.Normal, mid.Slice, mid.StringMap = initMidData()
	// 要开始好多好多的反射呀！吓人了！
	role_v := reflect.ValueOf(role).Elem()
	role_t := role_v.Type()
	field_num := role_v.NumField()
	for i := 0; i < field_num; i++ {
		field_v := role_v.Field(i)
		field_t := role_t.Field(i)
		if field_t.Name == "Role" {
			continue
		}
		field_name := field_t.Name
		// 开始遍历
		switch field_t.Type.String() {
		case "time.Time":
			mid.Normal.Time[field_name] = field_v.Interface().(time.Time)
		case "string":
			mid.Normal.String[field_name] = field_v.Interface().(string)
		case "bool":
			mid.Normal.Bool[field_name] = field_v.Interface().(bool)
		case "uint8":
			mid.Normal.Uint8[field_name] = field_v.Interface().(uint8)
		case "uint":
			mid.Normal.Uint[field_name] = field_v.Interface().(uint)
		case "uint64":
			mid.Normal.Uint64[field_name] = field_v.Interface().(uint64)
		case "int8":
			mid.Normal.Int8[field_name] = field_v.Interface().(int8)
		case "int":
			mid.Normal.Int[field_name] = field_v.Interface().(int)
		case "int64":
			mid.Normal.Int64[field_name] = field_v.Interface().(int64)
		case "float32":
			mid.Normal.Float32[field_name] = field_v.Interface().(float32)
		case "float64":
			mid.Normal.Float64[field_name] = field_v.Interface().(float64)
		case "complex64":
			mid.Normal.Complex64[field_name] = field_v.Interface().(complex64)
		case "complex128":
			mid.Normal.Complex128[field_name] = field_v.Interface().(complex128)

		case "[]string":
			mid.Slice.String[field_name] = field_v.Interface().([]string)
		case "[]bool":
			mid.Slice.Bool[field_name] = field_v.Interface().([]bool)
		case "[]uint8":
			mid.Slice.Uint8[field_name] = field_v.Interface().([]uint8)
		case "[]uint":
			mid.Slice.Uint[field_name] = field_v.Interface().([]uint)
		case "[]uint64":
			mid.Slice.Uint64[field_name] = field_v.Interface().([]uint64)
		case "[]int8":
			mid.Slice.Int8[field_name] = field_v.Interface().([]int8)
		case "[]int":
			mid.Slice.Int[field_name] = field_v.Interface().([]int)
		case "[]int64":
			mid.Slice.Int64[field_name] = field_v.Interface().([]int64)
		case "[]float32":
			mid.Slice.Float32[field_name] = field_v.Interface().([]float32)
		case "[]float64":
			mid.Slice.Float64[field_name] = field_v.Interface().([]float64)
		case "[]complex64":
			mid.Slice.Complex64[field_name] = field_v.Interface().([]complex64)
		case "[]complex128":
			mid.Slice.Complex128[field_name] = field_v.Interface().([]complex128)

		case "map[string]string":
			mid.StringMap.String[field_name] = field_v.Interface().(map[string]string)
		case "map[string]bool":
			mid.StringMap.Bool[field_name] = field_v.Interface().(map[string]bool)
		case "map[string]uint8":
			mid.StringMap.Uint8[field_name] = field_v.Interface().(map[string]uint8)
		case "map[string]uint":
			mid.StringMap.Uint[field_name] = field_v.Interface().(map[string]uint)
		case "map[string]uint64":
			mid.StringMap.Uint64[field_name] = field_v.Interface().(map[string]uint64)
		case "map[string][]int8":
			mid.StringMap.Int8[field_name] = field_v.Interface().(map[string]int8)
		case "map[string]int":
			mid.StringMap.Int[field_name] = field_v.Interface().(map[string]int)
		case "map[string]int64":
			mid.StringMap.Int64[field_name] = field_v.Interface().(map[string]int64)
		case "map[string]float32":
			mid.StringMap.Float32[field_name] = field_v.Interface().(map[string]float32)
		case "map[string]float64":
			mid.StringMap.Float64[field_name] = field_v.Interface().(map[string]float64)
		case "map[string]complex64":
			mid.StringMap.Complex64[field_name] = field_v.Interface().(map[string]complex64)
		case "map[string]complex128":
			mid.StringMap.Complex128[field_name] = field_v.Interface().(map[string]complex128)

		default:
			err = fmt.Errorf("hardstore[RoleMiddleData]EncodeRoleToMiddle: Unsupported data type: %v", field_t.Type.String())
			return
		}
	}
	return
}

// 解码角色，从中间编码转为角色
func DecodeMiddleToRole(mid RoleMiddleData, role roles.Roleer) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("hardstore[RoleMiddleData]DecodeMiddleToRole %v", e)
		}
	}()

	// 处理常规的那些东西
	role.New(mid.Version.Id)
	role.SetVersion(mid.Version.Version)
	role.SetFather(mid.Relation.Father)
	role.SetChildren(mid.Relation.Children)
	role.SetFriends(mid.Relation.Friends)
	role.SetContexts(mid.Relation.Contexts)

	// 处理数据
	role_v := reflect.ValueOf(role).Elem()
	role_t := role_v.Type()
	field_num := role_v.NumField()
	for i := 0; i < field_num; i++ {
		field_v := role_v.Field(i)
		field_t := role_t.Field(i)
		if field_t.Name == "Role" {
			continue
		}
		field_name := field_t.Name
		// 开始遍历
		var value interface{}
		var find bool
		switch field_t.Type.String() {
		case "time.Time":
			value, find = mid.Normal.Time[field_name]
		case "string":
			value, find = mid.Normal.String[field_name]
		case "bool":
			value, find = mid.Normal.Bool[field_name]
		case "uint8":
			value, find = mid.Normal.Uint8[field_name]
		case "uint":
			value, find = mid.Normal.Uint[field_name]
		case "uint64":
			value, find = mid.Normal.Uint64[field_name]
		case "int8":
			value, find = mid.Normal.Int8[field_name]
		case "int":
			value, find = mid.Normal.Int[field_name]
		case "int64":
			value, find = mid.Normal.Int64[field_name]
		case "float32":
			value, find = mid.Normal.Float32[field_name]
		case "float64":
			value, find = mid.Normal.Float64[field_name]
		case "complex64":
			value, find = mid.Normal.Complex64[field_name]
		case "complex128":
			value, find = mid.Normal.Complex128[field_name]

		case "[]string":
			value, find = mid.Slice.String[field_name]
		case "[]bool":
			value, find = mid.Slice.Bool[field_name]
		case "[]uint8":
			value, find = mid.Slice.Uint8[field_name]
		case "[]uint":
			value, find = mid.Slice.Uint[field_name]
		case "[]uint64":
			value, find = mid.Slice.Uint64[field_name]
		case "[]int8":
			value, find = mid.Slice.Int8[field_name]
		case "[]int":
			value, find = mid.Slice.Int[field_name]
		case "[]int64":
			value, find = mid.Slice.Int64[field_name]
		case "[]float32":
			value, find = mid.Slice.Float32[field_name]
		case "[]float64":
			value, find = mid.Slice.Float64[field_name]
		case "[]complex64":
			value, find = mid.Slice.Complex64[field_name]
		case "[]complex128":
			value, find = mid.Slice.Complex128[field_name]

		case "map[string]string":
			value, find = mid.StringMap.String[field_name]
		case "map[string]bool":
			value, find = mid.StringMap.Bool[field_name]
		case "map[string]uint8":
			value, find = mid.StringMap.Uint8[field_name]
		case "map[string]uint":
			value, find = mid.StringMap.Uint[field_name]
		case "map[string]uint64":
			value, find = mid.StringMap.Uint64[field_name]
		case "map[string][]int8":
			value, find = mid.StringMap.Int8[field_name]
		case "map[string]int":
			value, find = mid.StringMap.Int[field_name]
		case "map[string]int64":
			value, find = mid.StringMap.Int64[field_name]
		case "map[string]float32":
			value, find = mid.StringMap.Float32[field_name]
		case "map[string]float64":
			value, find = mid.StringMap.Float64[field_name]
		case "map[string]complex64":
			value, find = mid.StringMap.Complex64[field_name]
		case "map[string]complex128":
			value, find = mid.StringMap.Complex128[field_name]
		}
		if find == false {
			continue
		}
		value_v := reflect.ValueOf(value)
		field_v.Set(value_v)
	}
	return
}

func initMidData() (n RoleDataNormal, s RoleDataSlice, sm RoleDataStringMap) {
	n = RoleDataNormal{
		Time: make(map[string]time.Time),

		String: make(map[string]string),

		Bool: make(map[string]bool),

		Uint8:  make(map[string]uint8),
		Uint:   make(map[string]uint),
		Uint64: make(map[string]uint64),

		Int8:  make(map[string]int8),
		Int:   make(map[string]int),
		Int64: make(map[string]int64),

		Float32: make(map[string]float32),
		Float64: make(map[string]float64),

		Complex64:  make(map[string]complex64),
		Complex128: make(map[string]complex128),
	}

	s = RoleDataSlice{
		String: make(map[string][]string, 0),

		Bool: make(map[string][]bool, 0),

		Uint8:  make(map[string][]uint8, 0),
		Uint:   make(map[string][]uint, 0),
		Uint64: make(map[string][]uint64, 0),

		Int8:  make(map[string][]int8, 0),
		Int:   make(map[string][]int, 0),
		Int64: make(map[string][]int64, 0),

		Float32: make(map[string][]float32, 0),
		Float64: make(map[string][]float64, 0),

		Complex64:  make(map[string][]complex64, 0),
		Complex128: make(map[string][]complex128, 0),
	}

	sm = RoleDataStringMap{
		String: make(map[string]map[string]string),

		Bool: make(map[string]map[string]bool),

		Uint8:  make(map[string]map[string]uint8),
		Uint:   make(map[string]map[string]uint),
		Uint64: make(map[string]map[string]uint64),

		Int8:  make(map[string]map[string]int8),
		Int:   make(map[string]map[string]int),
		Int64: make(map[string]map[string]int64),

		Float32: make(map[string]map[string]float32),
		Float64: make(map[string]map[string]float64),

		Complex64:  make(map[string]map[string]complex64),
		Complex128: make(map[string]map[string]complex128),
	}

	return
}
