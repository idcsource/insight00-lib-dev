// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstore

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 解码一个角色，将二进制的角色存储进行解码
func DecodeRole(roleb, relab, verb []byte) (role roles.Roleer, err error) {
	role, err = nst.BytesGobStructForRoleer(roleb)
	if err != nil {
		return nil, fmt.Errorf("hardstore[HardStore]DecodeRole: %v", err)
	}
	var r_ralation RoleRelation
	err = nst.BytesGobStruct(relab, &r_ralation)
	if err != nil {
		return nil, fmt.Errorf("hardstore[HardStore]DecodeRole: %v", err)
	}
	var r_version RoleVersion
	err = nst.BytesGobStruct(verb, &r_version)
	if err != nil {
		return nil, fmt.Errorf("hardstore[HardStore]DecodeRole: %v", err)
	}

	role.SetFather(r_ralation.Father)
	role.SetChildren(r_ralation.Children)
	role.SetFriends(r_ralation.Friends)
	role.SetContexts(r_ralation.Contexts)
	role.SetVersion(r_version.Version)
	return role, nil
}

// 编码角色，将角色编码为两个部分的[]byte，一个是角色本身的数据roleb，一个是角色的关系relab
func EncodeRole(role roles.Roleer) (roleb, relab, verb []byte, err error) {
	roleb, err = nst.StructGobBytesForRoleer(role)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("hardstore[HardStore]EncodeRole: %v", err)
	}

	r_ralation := RoleRelation{
		Father:   role.GetFather(),
		Children: role.GetChildren(),
		Friends:  role.GetFriends(),
		Contexts: role.GetContexts(),
	}
	relab, err = nst.StructGobBytes(r_ralation)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("hardstore[HardStore]EncodeRole: %v", err)
	}
	r_version := RoleVersion{
		Version: role.Version(),
	}
	verb, err = nst.StructGobBytes(r_version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("hardstore[HardStore]EncodeRole: %v", err)
	}
	return
}

// 得到角色的保存名
func GetRoleStoreName(id string) (name string) {
	return random.GetSha1Sum(id)
}
