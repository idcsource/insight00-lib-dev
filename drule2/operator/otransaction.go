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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_EXIST_CHILD, rc_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_GET_FRIENDS, rf_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_SET_FRIENDS, rf_b)
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
	default:
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: The value's type must int64, float64 or complex128.")
		return
	}
	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]WriteFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_SET_FRIEND_STATUS, rf_b)
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
func (o *OTransaction) ReadFriendStatus(areaid, roleid, friendid string, bindbit int, value interface{}) (errs DRuleError) {
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
	default:
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: The value's type must int64, float64 or complex128.")
		return
	}

	// 编码
	rf_b, err := iendecode.StructGobBytes(rf)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: %v", err)
		return
	}
	// 传输
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_GET_FRIEND_STATUS, rf_b)
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
	default:
		err = fmt.Errorf("operator[OTransaction]ReadFriendStatus: The value's type not int64, float64 or complex128.")
	}
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_DEL_FRIEND, rf_b)
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_ADD_CONTEXT, rc_b)
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

// 查看是否有这个上下文
func (o *OTransaction) ExistContext(areaid, roleid, contextname string) (errs DRuleError) {
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
	cprocess := o.drule.tcpconn.OpenProgress()
	defer cprocess.Close()
	r, err := o.operatorSend(cprocess, areaid, roleid, OPERATE_EXIST_CONTEXT, rc_b)
	if err != nil {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", err)
		return
	}
	errs.Code = r.DataStat
	if r.DataStat != DATA_ALL_OK {
		errs.Err = fmt.Errorf("operator[OTransaction]ExistContext: %v", r.Error)
		return
	}
	return
}
