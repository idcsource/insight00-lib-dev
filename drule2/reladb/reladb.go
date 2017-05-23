// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package reladb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// Use TRule to create new relational database.
func NewRelaDBWithTRule(areaname string, instance *trule.TRule) (rdb *RelaDB, err error) {
	if instance == nil {
		err = fmt.Errorf("The TRule can not nil.")
		return
	}
	if exist := instance.AreaExist(areaname); exist == false {
		err = fmt.Errorf("The area %v not exist.", areaname)
		return
	}
	if have := instance.ExistRole(areaname, TABLE_CONTROL_NAME); have == false {
		tc := &TablesControl{}
		tc.New(TABLE_CONTROL_NAME)
		err = instance.StoreRole(areaname, tc)
		if err != nil {
			return
		}
	}
	service := &reladbService{
		dtype:    DRULE2_USE_TRULE,
		trule:    instance,
		areaname: areaname,
	}
	rdb = &RelaDB{
		service: service,
	}
	return
}

// Use DRule (in fact is drule2/operator) to create new relational database.
func NewRelaDBWithDRule(areaname string, instance *operator.Operator) (rdb *RelaDB, err error) {
	if instance == nil {
		err = fmt.Errorf("The DRule can not nil.")
		return
	}
	exist, errd := instance.AreaExist(areaname)
	if errd.IsError() != nil {
		err = errd.IsError()
		return
	}
	if exist == false {
		err = fmt.Errorf("The area %v not exist.", areaname)
		return
	}
	have, errd := instance.ExistRole(areaname, TABLE_CONTROL_NAME)
	if errd.IsError() != nil {
		err = errd.IsError()
		return
	}
	if have == false {
		tc := &TablesControl{}
		tc.New(TABLE_CONTROL_NAME)
		errd = instance.StoreRole(areaname, tc)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
	}
	service := &reladbService{
		dtype:    DRULE2_USE_DRULE,
		drule:    instance,
		areaname: areaname,
	}
	rdb = &RelaDB{
		service: service,
	}
	return
}

// Create new table.
// The fields is what fields need to index.
func (rdb *RelaDB) NewTable(tablename string, prototype roles.Roleer, fields ...string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("reladb[RelaDB]NewTable: %v", e)
		}
	}()
	tableid := TABLE_NAME_PREFIX + tablename
	// 查看用什么连接的
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tablehave := db.ExistRole(rdb.service.areaname, tableid)
		if tablehave == true {
			err = fmt.Errorf("The Table %v already exist.", tablename)
			return
		}
		valuer := reflect.Indirect(reflect.ValueOf(prototype))
		vname := valuer.Type().String()
		// 查看索引的字段（扫一遍，如果panci了正好就弹出错误了）
		fieldstype := make(map[string]string)
		for _, field := range fields {
			fv := valuer.FieldByName(field)
			fvt := fv.Type().String()
			fieldstype[field] = fvt
		}
		// 建立主表角色
		tablemain := &TableMain{
			TableName:      tablename,
			Prototype:      vname,
			IncrementCount: 0,
			IndexField:     fields,
		}
		tablemain.New(tableid)
		// 保存主表角色
		err = db.StoreRole(rdb.service.areaname, tablemain)
		if err != nil {
			return
		}

		// 建立索引表
		for fieldname, _ := range fieldstype {
			fieldid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
			indexrole := &TableIndex{
				FieldName: fieldname,
			}
			indexrole.New(fieldid)
			indexrole.Index = make(map[string]IndexGather)
			// 保存角色
			err = db.StoreRole(rdb.service.areaname, indexrole)
			if err != nil {
				return
			}
		}
		err = db.WriteChild(rdb.service.areaname, TABLE_CONTROL_NAME, tableid)
		if err != nil {
			return
		}
	} else {
		db := rdb.service.drule
		errd := operator.NewDRuleError()
		tablehave, errd := db.ExistRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		if tablehave == true {
			err = fmt.Errorf("The Table %v already exist.", tablename)
			return
		}
		valuer := reflect.Indirect(reflect.ValueOf(prototype))
		vname := valuer.Type().String()
		// 查看索引的字段（扫一遍，如果panci了正好就弹出错误了）
		fieldstype := make(map[string]string)
		for _, field := range fields {
			fv := valuer.FieldByName(field)
			fvt := fv.Type().String()
			if fvt != "string" && fvt != "int64" && fvt != "float64" && fvt != "bool" && fvt != "time.Time" {
				err = fmt.Errorf("The field %v type %v not support to index.", field, fvt)
				return
			}
			fieldstype[field] = fvt
		}
		// 建立主表角色
		tablemain := &TableMain{
			TableName:      tablename,
			Prototype:      vname,
			IncrementCount: 0,
			IndexField:     fields,
		}
		tablemain.New(tableid)
		// 保存主表角色
		errd = db.StoreRole(rdb.service.areaname, tablemain)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}

		// 建立索引表
		for fieldname, _ := range fieldstype {
			fieldid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
			indexrole := &TableIndex{
				FieldName: fieldname,
			}
			indexrole.New(fieldid)
			indexrole.Index = make(map[string]IndexGather)
			// 保存角色
			errd = db.StoreRole(rdb.service.areaname, indexrole)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
		}
		errd = db.WriteChild(rdb.service.areaname, TABLE_CONTROL_NAME, tableid)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
	}

	return
}

// Check if the table exist.
func (rdb *RelaDB) TableExist(tablename string) (exist bool, err error) {
	tableid := TABLE_NAME_PREFIX + tablename
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		exist = db.ExistRole(rdb.service.areaname, tableid)
		return
	} else {
		db := rdb.service.drule
		errd := operator.NewDRuleError()
		exist, errd = db.ExistRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
	}
	return
}

// Insert one Role to table. (Notice: the Role's id will be change.)
func (rdb *RelaDB) Insert(tablename string, instance roles.Roleer) (err error) {
	// 查看有无这个表
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}
	tableid := TABLE_NAME_PREFIX + tablename
	// 查看用什么连接的
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tran, _ := db.Begin()
		// 锁定表角色
		err = tran.LockRole(rdb.service.areaname, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		// 获取自增
		var ai uint64
		err = tran.ReadData(rdb.service.areaname, tableid, "IncrementCount", &ai)
		if err != nil {
			tran.Rollback()
			return
		}
		ai = ai + 1
		// 写入自增
		err = tran.WriteData(rdb.service.areaname, tableid, "IncrementCount", ai)
		if err != nil {
			tran.Rollback()
			return
		}
		// 获取索引字段
		var indexfields []string
		err = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if err != nil {
			tran.Rollback()
			return
		}
		// 编码角色
		var mid roles.RoleMiddleData
		mid, err = roles.EncodeRoleToMiddle(instance)
		if err != nil {
			tran.Rollback()
			return
		}
		// 修改角色id
		id := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(ai, 10)
		mid.Version.Id = id
		// 保存
		err = tran.StoreRoleFromMiddleData(rdb.service.areaname, mid)
		if err != nil {
			tran.Rollback()
			return
		}
		// 查看索引并保存
		for _, field := range indexfields {
			value, find := mid.Data.Point[field]
			if find == true {
				// 获取索引的索引信息
				var index map[string]IndexGather
				err = tran.ReadData(rdb.service.areaname, TABLE_NAME_PREFIX+tablename+TABLE_INDEX_PREFIX+field, "Index", &index)
				if err != nil {
					tran.Rollback()
					return
				}
				valuestring := random.GetSha1SumBytes(value)
				gather, have := index[valuestring]
				// 没有这个值就新建
				if have == false {
					index[valuestring] = []uint64{ai}
				} else {
					gather = append(gather, ai)
					index[valuestring] = gather
				}
				// 再存进去
				err = tran.WriteData(rdb.service.areaname, TABLE_NAME_PREFIX+tablename+TABLE_INDEX_PREFIX+field, "Index", &index)
				if err != nil {
					tran.Rollback()
					return
				}
			}
		}
		tran.Commit()
	} else {
		db := rdb.service.drule
		errd := operator.NewDRuleError()
		tran, errd := db.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		// 锁定表角色
		errd = tran.LockRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// 获取自增
		var ai uint64
		errd = tran.ReadData(rdb.service.areaname, tableid, "IncrementCount", &ai)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		ai = ai + 1
		// 写入自增
		errd = tran.WriteData(rdb.service.areaname, tableid, "IncrementCount", ai)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// 获取索引字段
		var indexfields []string
		errd = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// 编码角色
		var mid roles.RoleMiddleData
		mid, err = roles.EncodeRoleToMiddle(instance)
		if err != nil {
			tran.Rollback()
			return
		}
		// 修改角色id
		id := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(ai, 10)
		mid.Version.Id = id
		// 保存
		errd = tran.StoreRoleFromMiddleData(rdb.service.areaname, mid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// 查看索引并保存
		for _, field := range indexfields {
			value, find := mid.Data.Point[field]
			if find == true {
				// 获取索引的索引信息
				var index map[string]IndexGather
				errd = tran.ReadData(rdb.service.areaname, TABLE_NAME_PREFIX+tablename+TABLE_INDEX_PREFIX+field, "Index", &index)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
				valuestring := random.GetSha1SumBytes(value)
				gather, have := index[valuestring]
				// 没有这个值就新建
				if have == false {
					index[valuestring] = []uint64{ai}
				} else {
					gather = append(gather, ai)
					index[valuestring] = gather
				}
				// 再存进去
				errd = tran.WriteData(rdb.service.areaname, TABLE_NAME_PREFIX+tablename+TABLE_INDEX_PREFIX+field, "Index", &index)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
			}
		}
		tran.Commit()
	}
	return
}

// Return the max count.
func (rdb *RelaDB) Count(tablename string) (count uint64, err error) {
	// 查看有无这个表
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}
	tableid := TABLE_NAME_PREFIX + tablename
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		err = db.ReadData(rdb.service.areaname, tableid, "IncrementCount", &count)
		if err != nil {
			return
		}
	} else {
		db := rdb.service.drule
		errd := db.ReadData(rdb.service.areaname, tableid, "IncrementCount", &count)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
	}
	return
}
