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
	"time"

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst2"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

func (o *OTransaction) operatorSend(process *nst2.CConnect, areaid, roleid string, oz OperateZone, operate OperatorType, data []byte) (receipt O_DRuleReceipt, err error) {
	thestat := O_OperatorSend{
		OperatorName:  o.selfname,
		OperateZone:   oz,
		Operate:       operate,
		TransactionId: o.transaction_id,
		InTransaction: true,
		RoleId:        roleid,
		AreaId:        areaid,
		User:          o.drule.username,
		Unid:          o.drule.unid,
		Data:          data,
	}
	statbyte, err := iendecode.StructGobBytes(thestat)
	if err != nil {
		fmt.Println("a", err)
		return
	}
	rdata, err := process.SendAndReturn(statbyte)
	if err != nil {
		fmt.Println("b", err)
		return
	}
	receipt = O_DRuleReceipt{}
	err = iendecode.BytesGobStruct(rdata, &receipt)
	fmt.Println("b", err)
	return
}

// 返回事务ID
func (o *OTransaction) TransactionId() (id string) {
	return o.transaction_id
}

// 执行事务
func (o *OTransaction) Commit() (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]Commit: This transaction has been deleted.")
		return
	}

	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]Commit: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_NORMAL, OPERATE_TRAN_COMMIT, nil)

	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]Commit: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]Commit: %v", drule_r.Error)
		return
	}

	// 发送信号
	s := tranService{
		unid:   o.transaction_id,
		askfor: TRANSACTION_ASKFOR_END,
	}
	o.service.tran_signal <- s
	// 自己删除
	o.bedelete = true

	return
}

// 回滚事务
func (o *OTransaction) Rollback() (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]Rollback: This transaction has been deleted.")
		return
	}

	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]Rollback: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_ZONE_NORMAL, OPERATE_TRAN_ROLLBACK, nil)

	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]Rollback: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]Rollback: %v", drule_r.Error)
		return
	}

	// 发送信号
	s := tranService{
		unid:   o.transaction_id,
		askfor: TRANSACTION_ASKFOR_END,
	}
	o.service.tran_signal <- s
	// 自己删除
	o.bedelete = true

	return
}

// 是否存在这个角色
func (o *OTransaction) ExistRole(area, id string) (have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	role := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	role_b, err := iendecode.StructGobBytes(role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: %v", err)
		return
	}
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_EXIST_ROLE, role_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: %v", drule_r.Error)
		return
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistRole: %v", err)
		return
	}
	have = role.IfHave
	return
}

// 读取角色
func (o *OTransaction) ReadRole(area, id string, role roles.Roleer) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_READ_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", drule_r.Error)
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	// 解码角色
	err = roles.DecodeMiddleToRole(rsend.RoleBody, role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	return
}

// 读取角色到中间格式
func (o *OTransaction) ReadRoleToMiddleData(area, id string) (md roles.RoleMiddleData, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   area,
		RoleID: id,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_ZONE_NORMAL, OPERATE_READ_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", drule_r.Error)
	}
	// 解码返回数据
	err = iendecode.BytesGobStruct(drule_r.Data, &rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadRole: %v", err)
		return
	}
	md = rsend.RoleBody
	return
}

// 存入角色
func (o *OTransaction) StoreRole(area string, role roles.Roleer) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 获取角色ID
	roleid := role.ReturnId()
	// 转码角色
	rolemid, err := roles.EncodeRoleToMiddle(role)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
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
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, roleid, OPERATE_ZONE_NORMAL, OPERATE_WRITE_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", drule_r.Error)
		return
	}
	return
}

// 存入角色
func (o *OTransaction) StoreRoleFromMiddleData(area string, rolemid roles.RoleMiddleData) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, roleid, OPERATE_ZONE_NORMAL, OPERATE_WRITE_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]StoreRole: %v", drule_r.Error)
		return
	}
	return
}

// 删除角色
func (o *OTransaction) DeleteRole(areaid, roleid string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 生成传输
	rsend := O_RoleSendAndReceive{
		Area:   areaid,
		RoleID: roleid,
	}
	rsend_b, err := iendecode.StructGobBytes(rsend)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteRole: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteRole: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_ROLE, rsend_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteRole: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteRole: %v", drule_r.Error)
		return
	}
	return
}

// 设置父角色
func (o *OTransaction) WriteFather(areaid, roleid, fatherid string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFather: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	of := O_RoleFatherChange{
		Area:   areaid,
		Id:     roleid,
		Father: fatherid,
	}
	of_b, err := iendecode.StructGobBytes(of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFather: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFather: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FATHER, of_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFather: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFather: %v", drule_r.Error)
		return
	}
	return
}

// 重置父角色
func (o *OTransaction) ResetFather(areaid, roleid string) (errs DRuleError) {
	return o.WriteFather(areaid, roleid, "")
}

// 读角色的父亲
func (o *OTransaction) ReadFather(areaid, roleid string) (fatherid string, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	of := O_RoleFatherChange{
		Area: areaid,
		Id:   roleid,
	}
	of_b, err := iendecode.StructGobBytes(of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: %v", err)
		return
	}
	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FATHER, of_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: %v", drule_r.Error)
		return
	}
	// 解码获得
	err = iendecode.BytesGobStruct(drule_r.Data, &of)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFather: %v", err)
		return
	}
	fatherid = of.Father
	return
}

// 读取所有子角色名
func (o *OTransaction) ReadChildren(areaid, roleid string) (children []string, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构造
	oc := O_RoleAndChildren{
		Area: areaid,
		Id:   roleid,
	}
	oc_b, err := iendecode.StructGobBytes(oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CHILDREN, oc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: %v", drule_r.Error)
		return
	}
	// 解码结果
	err = iendecode.BytesGobStruct(drule_r.Data, &oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadChildren: %v", err)
		return
	}
	children = oc.Children
	return
}

// 写入所有子角色名
func (o *OTransaction) WriteChildren(areaid, roleid string, children []string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChildren: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构造
	oc := O_RoleAndChildren{
		Area:     areaid,
		Id:       roleid,
		Children: children,
	}
	oc_b, err := iendecode.StructGobBytes(oc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChildren: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChildren: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CHILDREN, oc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChildren: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChildren: %v", drule_r.Error)
		return
	}
	return
}

// 重置子角色
func (o *OTransaction) ResetChildren(areaid, roleid string) (errs DRuleError) {
	children := make([]string, 0)
	return o.WriteChildren(areaid, roleid, children)
}

// 写一个子角色
func (o *OTransaction) WriteChild(areaid, roleid, childid string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChild: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChild: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChild: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_ADD_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChild: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteChild: %v", drule_r.Error)
		return
	}
	return
}

// 删除一个子角色
func (o *OTransaction) DeleteChild(areaid, roleid, childid string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteChild: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteChild: %v", err)
		return
	}

	// 开始传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteChild: %v", err)
		return
	}
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteChild: %v", err)
		return
	}
	errs.Code = drule_r.DataStat
	if drule_r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteChild: %v", drule_r.Error)
		return
	}
	return
}

// 是否拥有这个子角色
func (o *OTransaction) ExistChild(areaid, roleid, childid string) (have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndChild{
		Area:  areaid,
		Id:    roleid,
		Child: childid,
	}
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: %v", err)
		return
	}

	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_EXIST_CHILD, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: %v", r.Error)
		return
	}
	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistChild: %v", err)
		return
	}
	have = rc.Exist
	return
}

// 读取所有朋友关系
func (o *OTransaction) ReadFriends(areaid, roleid string) (firends map[string]roles.Status, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rf := O_RoleAndFriends{
		Area: areaid,
		Id:   roleid,
	}
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: %v", err)
		return
	}

	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FRIENDS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: %v", r.Error)
		return
	}
	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriends: %v", err)
		return
	}
	firends = rf.Friends
	return
}

// 写入所有朋友关系
func (o *OTransaction) WriteFriends(areaid, roleid string, friends map[string]roles.Status) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriends: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rf := O_RoleAndFriends{
		Area:    areaid,
		Id:      roleid,
		Friends: friends,
	}
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriends: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriends: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FRIENDS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriends: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriends: %v", r.Error)
		return
	}
	return
}

// 重置所有朋友关系
func (o *OTransaction) ResetFriends(areaid, roleid string) (errs DRuleError) {
	friends := make(map[string]roles.Status)
	return o.WriteFriends(areaid, roleid, friends)
}

// 设置朋友状态
func (o *OTransaction) WriteFriendStatus(areaid, roleid, friendid string, bindbit int, value interface{}) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
			errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", e)
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
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: The value's type must int64, float64, complex128 or string.")
		return
	}
	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_FRIEND_STATUS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", r.Error)
		return
	}
	return
}

// 写入一个朋友
func (o *OTransaction) WriteFriend(areaid, roleid, friendid string, bind int64) (errs DRuleError) {
	return o.WriteFriendStatus(areaid, roleid, friendid, 0, bind)
}

// 读取朋友的状态
func (o *OTransaction) ReadFriendStatus(areaid, roleid, friendid string, bindbit int, value interface{}) (have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
			errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", e)
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
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: The value's type must int64, float64, complex128 or string.")
		return
	}

	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_FRIEND_STATUS, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", r.Error)
		return
	}

	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", err)
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
		err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: The value's type not int64, float64, complex128 or string.")
	}
	have = rf.Exist
	return
}

// 删除一个朋友关系
func (o *OTransaction) DeleteFriend(areaid, roleid, friendid string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteFriend: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构造
	rf := O_RoleAndFriend{
		Area:   areaid,
		Id:     roleid,
		Friend: friendid,
	}

	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteFriend: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteFriend: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_FRIEND, rf_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteFriend: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteFriend: %v", r.Error)
		return
	}
	return
}

// 创建一个空上下文
func (o *OTransaction) CreateContext(areaid, roleid, contextname string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]CreateContext: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]CreateContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]CreateContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_ADD_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]CreateContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]CreateContext: %v", r.Error)
		return
	}
	return
}

// 查看是否有这个上下文
func (o *OTransaction) ExistContext(areaid, roleid, contextname string) (have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_EXIST_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", r.Error)
		return
	}
	have = rc.Exist
	return
}

// 清除一个上下文
func (o *OTransaction) DropContext(areaid, roleid, contextname string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]DropContext: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContext{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DropContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DropContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DROP_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DropContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]DropContext: %v", r.Error)
		return
	}
	return
}

// 返回某个上下文的全部信息
func (o *OTransaction) ReadContext(areaid, roleid, contextname string) (context roles.Context, have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContext_Data{
		Area:    areaid,
		Id:      roleid,
		Context: contextname,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_READ_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContext: %v", err)
		return
	}
	context = rc.ContextBody
	have = rc.Exist
	return
}

// 删除一个上下文绑定
func (o *OTransaction) DeleteContextBind(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteContextBind: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteContextBind: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteContextBind: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_DEL_CONTEXT_BIND, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteContextBind: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]DeleteContextBind: %v", r.Error)
		return
	}
	return
}

// 获取同一个上下文的同一个绑定值的所有角色id
func (o *OTransaction) ReadContextSameBind(areaid, roleid, contextname string, upordown roles.ContextUpDown, bind int64) (rolesid []string, have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_CONTEXT_SAME_BIND, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextSameBind: %v", err)
		return
	}
	rolesid = rc.Gather
	have = rc.Exist
	return
}

// 返回所有上下文组的名称
func (o *OTransaction) ReadContextsName(areaid, roleid string) (names []string, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContext_Data{
		Area: areaid,
		Id:   roleid,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXTS_NAME, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextsName: %v", err)
		return
	}
	names = rc.Gather
	return
}

// 设定上下文的状态
func (o *OTransaction) WriteContextStatus(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string, bindbit int, value interface{}) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
			errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: %v", e)
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
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: The value's type must int64, float64, complex128 or string.")
		return
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CONTEXT_STATUS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContextStatus: %v", r.Error)
		return
	}
	return
}

// 获取一个上下文的状态
func (o *OTransaction) ReadContextStatus(areaid, roleid, contextname string, upordown roles.ContextUpDown, bindrole string, bindbit int, value interface{}) (have bool, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
			errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", e)
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
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: The value's type must int64, float64, complex128 or string.")
		return
	}

	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXT_STATUS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", r.Error)
		return
	}

	// 解码返回
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContextStatus: %v", err)
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
		err = fmt.Errorf("operator[OTransaction]ReadContextStatus: The value's type not int64, float64, complex128 or string.")
	}
	have = rc.Exist
	return
}

// 设定完全上下文
func (o *OTransaction) WriteContexts(areaid, roleid string, contexts map[string]roles.Context) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContexts: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContexts{
		Area:     areaid,
		Id:       roleid,
		Contexts: contexts,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContexts: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContexts: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_CONTEXTS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContexts: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteContexts: %v", r.Error)
		return
	}
	return
}

// 获取完全上下文
func (o *OTransaction) ReadContexts(areaid, roleid string) (contexts map[string]roles.Context, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rc := O_RoleAndContexts{
		Area: areaid,
		Id:   roleid,
	}
	// 编码
	rc_b, err := iendecode.StructGobBytes(rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_CONTEXTS, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: %v", r.Error)
		return
	}
	// 解码
	err = iendecode.BytesGobStruct(r.Data, &rc)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadContexts: %v", err)
		return
	}
	contexts = rc.Contexts
	return
}

// 重置上下文
func (o *OTransaction) ResetContexts(areaid, roleid string) (errs DRuleError) {
	contexts := make(map[string]roles.Context)
	return o.WriteContexts(areaid, roleid, contexts)
}

// 写入数据
func (o *OTransaction) WriteData(areaid, roleid, name string, data interface{}) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 数据编码
	data_b, err := iendecode.StructGobBytes(data)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: %v", err)
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
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteData: %v", r.Error)
		return
	}
	return
}

// 写入数据
func (o *OTransaction) WriteDataFromByte(areaid, roleid, name string, data_b []byte) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteDataFromByte: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

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
		errs.Err = fmt.Errorf("operator[OTransaction]WriteDataFromByte: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteDataFromByte: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_SET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteDataFromByte: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteDataFromByte: %v", r.Error)
		return
	}
	return
}

// 读取数据
func (o *OTransaction) ReadData(areaid, roleid, name string, data interface{}) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", err)
		return
	}
	// 解码数据
	err = iendecode.BytesGobStruct(rd.Data, data)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadData: %v", err)
		return
	}
	return
}

// 读取数据
func (o *OTransaction) ReadDataToByte(areaid, roleid, name string) (data_b []byte, errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rd := O_RoleData_Data{
		Area: areaid,
		Id:   roleid,
		Name: name,
	}
	// 编码
	rd_b, err := iendecode.StructGobBytes(rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ZONE_NORMAL, OPERATE_GET_DATA, rd_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: %v", r.Error)
		return
	}
	// 解码返回值
	err = iendecode.BytesGobStruct(r.Data, &rd)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadDataToByte: %v", err)
		return
	}
	data_b = rd.Data
	return
}

// 准备锁定角色
func (o *OTransaction) LockRole(areaid string, roleids ...string) (errs DRuleError) {
	var err error
	errs = NewDRuleError()
	// 查看是否删除
	if o.bedelete == true {
		errs.Err = fmt.Errorf("operator[OTransaction]LockRole: This transaction has been deleted.")
		return
	}
	// 事务续期
	o.activetime = time.Now()

	// 构建
	rt := O_Transaction{
		TransactionId: o.transaction_id,
		Area:          areaid,
		PrepareIDs:    roleids,
	}
	// 编码
	rt_b, err := iendecode.StructGobBytes(rt)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]LockRole: %v", err)
		return
	}
	// 传输
	cprocess, err := o.drule.tcpconn.OpenProgress()
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]LockRole: %v", err)
		return
	}
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, "", OPERATE_ZONE_NORMAL, OPERATE_TRAN_PREPARE, rt_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]LockRole: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]LockRole: %v", r.Error)
		return
	}
	return
}
