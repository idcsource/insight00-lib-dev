// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"fmt"
	"reflect"

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 是否存在这个角色
func (o *Operator) ExistRole(area, id string) (have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	role := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	role_b, err := iendecode.StructGobBytes(role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistRole: %v", err)
		return
	}
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_EXIST_ROLE, role_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ExistRole: %v", drule_r.Error)
		return
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistRole: %v", err)
		return
	}
	have = role.IfHave
	return
}

// 读取角色
func (o *Operator) ReadRole(area, id string, role roles.Roleer) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_READ_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", drule_r.Error)
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	// 解码角色
	err = roles.DecodeMiddleToRole(rsend.RoleBody, role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	return
}

// 读角色到中间数据格式
func (o *Operator) ReadRoleToMiddleData(area, id string) (md roles.RoleMiddleData, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_READ_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", drule_r.Error)
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadRole: %v", err)
		return
	}
	md = rsend.RoleBody
	return
}

// 存入角色
func (o *Operator) StoreRole(area string, role roles.Roleer) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		errs.Code = DATA_RETURN_ERROR
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 获取角色ID
	roleid := role.ReturnId()
	// 转码角色
	rolemid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		errs.Code = DATA_RETURN_ERROR
		return
	}

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:     area,
		RoleID:   roleid,
		RoleBody: rolemid,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		errs.Code = DATA_RETURN_ERROR
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		errs.Code = DATA_RETURN_ERROR
		return
	}
	defer cprocess.Close()
	fmt.Println("this")
	drule_r, err := o.operatorSend(cprocess, area, roleid, OPERATE_ZONE_NORMAL, OPERATE_WRITE_ROLE, rsend_b)
	fmt.Println("this2")
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		errs.Code = DATA_RETURN_ERROR
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", drule_r.Error)
		return
	}
	return
}

// 保存角色
func (o *Operator) StoreRoleFromMiddleData(area string, rolemid roles.RoleMiddleData) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 获取角色ID
	roleid := rolemid.Version.Id

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:     area,
		RoleID:   roleid,
		RoleBody: rolemid,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, roleid, OPERATE_ZONE_NORMAL, OPERATE_WRITE_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]StoreRole: %v", drule_r.Error)
		return
	}
	return
}

// 删除角色
func (o *Operator) DeleteRole(areaid, roleid string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   areaid,
		RoleID: roleid,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DeleteRole: %v", drule_r.Error)
		return
	}
	return
}

// 设置父角色
func (o *Operator) WriteFather(areaid, roleid, fatherid string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	of := O_RoleFatherChange{
		Area:   areaid,
		Id:     roleid,
		Father: fatherid,
	}
	of_b, err := iendecode.StructGobBytes(of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFather: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFather: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FATHER, of_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFather: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteFather: %v", drule_r.Error)
		return
	}
	return
}

// 重置父角色
func (o *Operator) ResetFather(areaid, roleid string) (errs DRuleError) {
	return o.WriteFather(areaid, roleid, "")
}

// 读角色的父亲
func (o *Operator) ReadFather(areaid, roleid string) (fatherid string, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	of := O_RoleFatherChange{
		Area: areaid,
		Id:   roleid,
	}
	of_b, err := iendecode.StructGobBytes(of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFather: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFather: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FATHER, of_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFather: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadFather: %v", drule_r.Error)
		return
	}
	// 解码获得
	err = iendecode.BytesGobStruct(drule_r.Data, &of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFather: %v", err)
		return
	}
	fatherid = of.Father
	return
}

// 读取所有子角色名
func (o *Operator) ReadChildren(areaid, roleid string) (children []string, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	oc := O_RoleAndChildren{
		Area: areaid,
		Id:   roleid,
	}
	oc_b, err := iendecode.StructGobBytes(oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadChildren: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadChildren: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CHILDREN, oc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadChildren: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadChildren: %v", drule_r.Error)
		return
	}
	// 解码结果
	err = iendecode.BytesGobStruct(drule_r.Data, &oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadChildren: %v", err)
		return
	}
	children = oc.Children
	return
}

// 写入所有子角色名
func (o *Operator) WriteChildren(areaid, roleid string, children []string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	oc := O_RoleAndChildren{
		Area:     areaid,
		Id:       roleid,
		Children: children,
	}
	oc_b, err := iendecode.StructGobBytes(oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChildren: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChildren: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CHILDREN, oc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChildren: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteChildren: %v", drule_r.Error)
		return
	}
	return
}

// 重置子角色
func (o *Operator) ResetChildren(areaid, roleid string) (errs DRuleError) {
	children := make([]string, 0)
	return o.WriteChildren(areaid, roleid, children)
}

// 写一个子角色
func (o *Operator) WriteChild(areaid, roleid, childid string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChild: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChild: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_ADD_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteChild: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteChild: %v", drule_r.Error)
		return
	}
	return
}

// 删除一个子角色
func (o *Operator) DeleteChild(areaid, roleid, childid string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteChild: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteChild: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteChild: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DeleteChild: %v", drule_r.Error)
		return
	}
	return
}

// 是否拥有这个子角色
func (o *Operator) ExistChild(areaid, roleid, childid string) (have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistChild: %v", err)
		return
	}

	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistChild: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_EXIST_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistChild: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ExistChild: %v", r.Error)
		return
	}
	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistChild: %v", err)
		return
	}
	have = rc.Exist
	return
}

// 读取所有朋友关系
func (o *Operator) ReadFriends(areaid, roleid string) (firends map[string]roles.Status, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rf := O_RoleAndFriends{
		Area: areaid,
		Id:   roleid,
	}
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriends: %v", err)
		return
	}

	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriends: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FRIENDS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriends: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriends: %v", r.Error)
		return
	}
	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriends: %v", err)
		return
	}
	firends = rf.Friends
	return
}

// 写入所有朋友关系
func (o *Operator) WriteFriends(areaid, roleid string, friends map[string]roles.Status) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rf := O_RoleAndFriends{
		Area:    areaid,
		Id:      roleid,
		Friends: friends,
	}
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriends: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriends: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FRIENDS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriends: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriends: %v", r.Error)
		return
	}
	return
}

// 重置所有朋友关系
func (o *Operator) ResetFriends(areaid, roleid string) (errs DRuleError) {
	friends := make(map[string]roles.Status)
	return o.WriteFriends(areaid, roleid, friends)
}

// 设置朋友状态
func (o *Operator) WriteFriendStatus(areaid, roleid, friendid string, bindbit int, value interface{}) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	rf := O_RoleAndFriend{
		Area:   areaid,
		Id:     roleid,
		Friend: friendid,
		Bit:    bindbit,
	}

	// 拦截反射错误
	defer func() {
		if e := recover(); e != nil {
			errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: %v", e)
		}
	}()
	value_v := reflect.Indirect(reflect.ValueOf(value))
	vname := value_v.Type().String()
	switch vname {
	case "int":
		rf.Int = value_v.Int()
		rf.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		rf.Int = value_v.Int()
		rf.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		rf.Float = value_v.Float()
		rf.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		rf.Float = value_v.Float()
		rf.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		rf.Complex = value_v.Complex()
		rf.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		rf.Complex = value_v.Complex()
		rf.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "string":
		rf.String = value_v.String()
		rf.Single = roles.STATUS_VALUE_TYPE_STRING
	default:
		errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: The value's type must int64, float64, complex128 or string.")
		return
	}
	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FRIEND_STATUS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteFriendStatus: %v", r.Error)
		return
	}
	return
}

// 写入一个朋友
func (o *Operator) WriteFriend(areaid, roleid, friendid string, bind int64) (errs DRuleError) {
	return o.WriteFriendStatus(areaid, roleid, friendid, 0, bind)
}

// 读取朋友的状态
func (o *Operator) ReadFriendStatus(areaid, roleid, friendid string, bindbit int, value interface{}) (have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	rf := O_RoleAndFriend{
		Area:   areaid,
		Id:     roleid,
		Friend: friendid,
		Bit:    bindbit,
	}

	// 拦截反射错误
	defer func() {
		if e := recover(); e != nil {
			errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", e)
		}
	}()
	value_v := reflect.Indirect(reflect.ValueOf(value))
	vname := value_v.Type().String()
	switch vname {
	case "int":
		rf.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		rf.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		rf.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		rf.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		rf.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		rf.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "string":
		rf.Single = roles.STATUS_VALUE_TYPE_STRING
	default:
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: The value's type must int64, float64, complex128 or string.")
		return
	}

	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FRIEND_STATUS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", r.Error)
		return
	}

	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadFriendStatus: %v", err)
		return
	}

	// 装入数据
	switch rf.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		value_v.SetInt(rf.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		value_v.SetFloat(rf.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		value_v.SetComplex(rf.Complex)
	case roles.STATUS_VALUE_TYPE_STRING:
		value_v.SetString(rf.String)
	default:
		err = fmt.Errorf("operator[Operator]ReadFriendStatus: The value's type not int64, float64, complex128 or string.")
	}
	have = rf.Exist
	return
}

// 删除一个朋友关系
func (o *Operator) DeleteFriend(areaid, roleid, friendid string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	rf := O_RoleAndFriend{
		Area:   areaid,
		Id:     roleid,
		Friend: friendid,
	}

	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_FRIEND, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", r.Error)
		return
	}
	return
}

// 创建一个空上下文
func (o *Operator) CreateContext(areaid, roleid, contextname string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]CreateContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]CreateContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_ADD_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DeleteFriend: %v", r.Error)
		return
	}
	return
}

// 查看是否有这个上下文
func (o *Operator) ExistContext(areaid, roleid, contextname string) (have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_EXIST_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ExistContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ExistContext: %v", r.Error)
		return
	}
	have = rc.Exist
	return
}

// 清除一个上下文
func (o *Operator) DropContext(areaid, roleid, contextname string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DropContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DropContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DROP_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DropContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DropContext: %v", r.Error)
		return
	}
	return
}

// 返回某个上下文的全部信息
func (o *Operator) ReadContext(areaid, roleid, contextname string) (context roles.Context, have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext_Data{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_READ_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadContext: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContext: %v", err)
		return
	}
	context = rc.ContextBody
	have = rc.Exist
	return
}

// 删除一个上下文绑定
func (o *Operator) DeleteContextBind(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext{
		Area:     areaid,
		Id:       roleid,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindrole,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteContextBind: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteContextBind: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_CONTEXT_BIND, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]DeleteContextBind: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]DeleteContextBind: %v", r.Error)
		return
	}
	return
}

// 获取同一个上下文的同一个绑定值的所有角色id
func (o *Operator) ReadContextSameBind(areaid, roleid, contextname string, upordown roles.ContextUpDown, bind int64) (rolesid []string, have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext_Data{
		Area:     areaid,
		Id:       roleid,
		Context:  contextname,
		UpOrDown: upordown,
		Bit:      0,
		Int:      bind,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextSameBind: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextSameBind: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_CONTEXT_SAME_BIND, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextSameBind: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextSameBind: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextSameBind: %v", err)
		return
	}
	rolesid = rc.Gather
	have = rc.Exist
	return
}

// 返回所有上下文组的名称
func (o *Operator) ReadContextsName(areaid, roleid string) (names []string, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContext_Data{
		Area: areaid,
		Id:   roleid,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextsName: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextsName: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXTS_NAME, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextsName: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextsName: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextsName: %v", err)
		return
	}
	names = rc.Gather
	return
}

// 设定上下文的状态
func (o *Operator) WriteContextStatus(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string, bindbit int, value interface{}) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	rc := O_RoleAndContext_Data{
		Area:     areaid,
		Id:       roleid,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindrole,
		Bit:      bindbit,
	}

	// 拦截反射错误
	defer func() {
		if e := recover(); e != nil {
			errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: %v", e)
		}
	}()
	value_v := reflect.Indirect(reflect.ValueOf(value))
	vname := value_v.Type().String()
	switch vname {
	case "int":
		rc.Int = value_v.Int()
		rc.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		rc.Int = value_v.Int()
		rc.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		rc.Float = value_v.Float()
		rc.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		rc.Float = value_v.Float()
		rc.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		rc.Complex = value_v.Complex()
		rc.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		rc.Complex = value_v.Complex()
		rc.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "string":
		rc.String = value_v.String()
		rc.Single = roles.STATUS_VALUE_TYPE_STRING
	default:
		errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: The value's type must int64, float64, complex128 or string.")
		return
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CONTEXT_STATUS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteContextStatus: %v", r.Error)
		return
	}
	return
}

// 获取一个上下文的状态
func (o *Operator) ReadContextStatus(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string, bindbit int, value interface{}) (have bool, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构造
	rc := O_RoleAndContext_Data{
		Area:     areaid,
		Id:       roleid,
		Context:  contextname,
		UpOrDown: upordown,
		BindRole: bindrole,
		Bit:      bindbit,
	}

	// 拦截反射错误
	defer func() {
		if e := recover(); e != nil {
			errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", e)
		}
	}()
	value_v := reflect.Indirect(reflect.ValueOf(value))
	vname := value_v.Type().String()
	switch vname {
	case "int":
		rc.Single = roles.STATUS_VALUE_TYPE_INT
	case "int64":
		rc.Single = roles.STATUS_VALUE_TYPE_INT
	case "float":
		rc.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "float64":
		rc.Single = roles.STATUS_VALUE_TYPE_FLOAT
	case "complex64":
		rc.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "complex128":
		rc.Single = roles.STATUS_VALUE_TYPE_COMPLEX
	case "string":
		rc.Single = roles.STATUS_VALUE_TYPE_STRING
	default:
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: The value's type must int64, float64, complex128 or string.")
		return
	}

	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXT_STATUS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", r.Error)
		return
	}

	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContextStatus: %v", err)
		return
	}

	// 装入数据
	switch rc.Single {
	case roles.STATUS_VALUE_TYPE_INT:
		value_v.SetInt(rc.Int)
	case roles.STATUS_VALUE_TYPE_FLOAT:
		value_v.SetFloat(rc.Float)
	case roles.STATUS_VALUE_TYPE_COMPLEX:
		value_v.SetComplex(rc.Complex)
	case roles.STATUS_VALUE_TYPE_STRING:
		value_v.SetString(rc.String)
	default:
		err = fmt.Errorf("operator[Operator]ReadContextStatus: The value's type not int64, float64, complex128 or string.")
	}
	have = rc.Exist
	return
}

// 设定完全上下文
func (o *Operator) WriteContexts(areaid, roleid string, contexts map[string]roles.Context) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContexts{
		Area:     areaid,
		Id:       roleid,
		Contexts: contexts,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContexts: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContexts: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CONTEXTS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteContexts: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteContexts: %v", r.Error)
		return
	}
	return
}

// 获取完全上下文
func (o *Operator) ReadContexts(areaid, roleid string) (contexts map[string]roles.Context, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rc := O_RoleAndContexts{
		Area: areaid,
		Id:   roleid,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContexts: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContexts: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXTS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContexts: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadContexts: %v", r.Error)
		return
	}
	// 解码
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadContexts: %v", err)
		return
	}
	contexts = rc.Contexts
	return
}

// 重置上下文
func (o *Operator) ResetContexts(areaid, roleid string) (errs DRuleError) {
	contexts := make(map[string]roles.Context)
	return o.WriteContexts(areaid, roleid, contexts)
}

// 写入数据
func (o *Operator) WriteData(areaid, roleid, name string, data interface{}) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 数据编码
	data_b, err := iendecode.StructGobBytes(data)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteData: %v", err)
		return
	}
	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
		Data: data_b,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteData: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteData: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteData: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteData: %v", r.Error)
		return
	}
	return
}

// 写入数据
func (o *Operator) WriteDataFromByte(areaid, roleid, name string, data_b []byte) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
		Data: data_b,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteDataFromByte: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteDataFromByte: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]WriteDataFromByte: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]WriteDataFromByte: %v", r.Error)
		return
	}
	return
}

// 读取数据
func (o *Operator) ReadData(areaid, roleid, name string, data interface{}) (errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", err)
		return
	}
	// 解码数据
	err = iendecode.BytesGobStruct(rd.Data, data)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadData: %v", err)
		return
	}
	return
}

// 读取数据
func (o *Operator) ReadDataToByte(areaid, roleid, name string) (data_b []byte, errs DRuleError) {
	errs = NewDRuleError()
	if o.runstatus != OPERATOR_RUN_RUNNING {
		errs.Err = fmt.Errorf("The Operator is colosed.")
		return
	}

	errs = o.checkLogin()
	if errs.IsError() != nil {
		return
	}

	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadDataToByte: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadDataToByte: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadDataToByte: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[Operator]ReadDataToByte: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[Operator]ReadDataToByte: %v", err)
		return
	}
	data_b = rd.Data
	return
}
