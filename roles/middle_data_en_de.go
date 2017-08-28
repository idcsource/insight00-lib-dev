// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package roles

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// 编码角色，将角色编码为中期存储格式
func EncodeRoleToMiddle(role Roleer) (mid *RoleMiddleData, err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]EncodeRoleToMiddle: %v", e)
		}
	}()

	mid = &RoleMiddleData{
		VersionChange:  true,
		DataChange:     true,
		RelationChange: true,
	}

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
	mid.Data = RoleData{
		Point: make(map[string]*RoleDataPoint),
	}

	role_v := reflect.ValueOf(role).Elem()
	role_t := role_v.Type()
	if role_t.String() == "roles.Role" || role_t.String() == "*roles.Role" {
		return
	}
	field_num := role_v.NumField()
	for i := 0; i < field_num; i++ {
		field_v := role_v.Field(i)
		field_t := role_t.Field(i)
		if field_t.Name == "Role" {
			continue
		}
		field_name := field_t.Name
		field_type := field_t.Type.Name()
		// field_type := field_t.Type.String()
		mid.Data.Point[field_name] = &RoleDataPoint{
			Type: field_type,
			Data: field_v.Interface(),
		}

		//		if in := typeWithIn(field_type); in == true {
		//			mid.Data.Point[field_name] = field_v.Interface()
		//		} else {
		//			mid.Data.Point[field_name], err = iendecode.StructGobBytes(field_v.Interface())
		//			if err != nil {
		//				err = fmt.Errorf("roles[RoleMiddleData]EncodeRoleToMiddle: %v", err)
		//				return
		//			}
		//		}
	}

	return
}

// 解码角色，从中间编码转为角色
func DecodeMiddleToRole(mid *RoleMiddleData, role Roleer) (err error) {
	// 拦截恐慌
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("roles[RoleMiddleData]DecodeMiddleToRole %v", e)
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
	if role_t.String() == "roles.Role" || role_t.String() == "*roles.Role" {
		return
	}
	field_num := role_v.NumField()
	for i := 0; i < field_num; i++ {
		field_v := role_v.Field(i)
		field_t := role_t.Field(i)
		if field_t.Name == "Role" {
			continue
		}
		field_name := field_t.Name
		// field_type := field_t.Type.String()
		fv := reflect.ValueOf(mid.Data.Point[field_name].Data)
		field_v.Set(fv)
		//		if _, find := mid.Data.Point[field_name]; find == true {
		//			if in := typeWithIn(field_type); in == true {
		//				fv := reflect.ValueOf(mid.Data.Point[field_name])
		//				field_v.Set(fv)
		//			} else {
		//				err = iendecode.BytesGobReflect(mid.Data.Point[field_name].([]byte), field_v)
		//			}
		//		}
	}

	return
}

func typeWithIn(name string) (in bool, status uint8) {
	name = strings.ToLower(name)
	types := []string{
		"bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128", "string", "byte", "time.time",
	}
	for _, the := range types {
		if name == the {
			return true, 1
		}
		if name == "[]"+the {
			return true, 2
		}
		if name == "*"+the {
			return true, 3
		}

		if name == "*[]"+the {
			return true, 4
		}
		if t, _ := regexp.MatchString(`map\[([^ ]+)\]`+the, name); t == true {
			return true, 5
		}
	}
	return false, 0
}
