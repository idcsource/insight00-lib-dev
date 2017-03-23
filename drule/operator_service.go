// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 读取一个角色
func (o *Operator) ReadRole(id string) (role roles.Roleer, err error) {
	// 去执行
	role_sr := Net_RoleSendAndReceive{}
	role_sr.RoleID = id
	// 转码
	role_sr_b, err := nst.StructGobBytes(role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadRole: %v", err)
		return
	}
	_, err = o.sendReadAndDecodeData(id, OPERATE_READ_ROLE, role_sr_b, &role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]ReadRole: %v", err)
		return
	}
	// 还原角色
	role, err = hardstore.DecodeRole(role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer)
	return
}

// 存储一个角色
func (o *Operator) StoreRole(role roles.Roleer) (err error) {
	// 获取角色ID
	roleid := role.ReturnId()
	// 编码角色
	role_sr := Net_RoleSendAndReceive{}
	role_sr.RoleBody, role_sr.RoleRela, role_sr.RoleVer, err = hardstore.EncodeRole(role)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	role_sr.RoleID = roleid
	// 编码传输的代码
	role_sr_b, err := nst.StructGobBytes(role_sr)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	err = o.sendWriteToServer(roleid, OPERATE_WRITE_ROLE, role_sr_b)
	if err != nil {
		err = fmt.Errorf("drule[Operator]StoreRole: %v", err)
		return
	}
	return
}
