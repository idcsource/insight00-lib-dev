// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"fmt"
	"time"

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

func (o *OTransaction) operatorSend(process *nst.ProgressData, areaid, roleid string, operate OperatorType, data []byte) (receipt O_DRuleReceipt, err error) {
	thestat := O_OperatorSend{
		OperatorName:  o.selfname,
		Operate:       operate,
		TransactionId: o.transaction_id,
		InTransaction: true,
		RoleId:        roleid,
		AreaId:        areaid,
		Unid:          o.drule.unid,
		Data:          data,
	}
	statbyte, err := iendecode.StructGobBytes(thestat)
	if err != nil {
		return
	}
	rdata, err := process.SendAndReturn(statbyte)
	if err != nil {
		return
	}
	receipt = O_DRuleReceipt{}
	err = iendecode.BytesGobStruct(rdata, &receipt)
	return
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

	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_TRAN_COMMIT, nil)

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

	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, "", "", OPERATE_TRAN_ROLLBACK, nil)

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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_EXIST_ROLE, role_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, id, OPERATE_READ_ROLE, rsend_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, area, roleid, OPERATE_WRITE_ROLE, rsend_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_DEL_ROLE, rsend_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_SET_FATHER, of_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_GET_FATHER, of_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_GET_CHILDREN, oc_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_SET_CHILDREN, oc_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ADD_CHILD, rc_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	drule_r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_DEL_CHILD, rc_b)
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
