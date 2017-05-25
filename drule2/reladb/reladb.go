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
	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
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

// Drop table if the table exist.
func (rdb *RelaDB) DropTable(tablename string) (err error) {
	// check if have the table, if table not exist, just return.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		return
	}

	tableid := TABLE_NAME_PREFIX + tablename

	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tran, _ := db.Begin()
		// log the table Role
		err = tran.LockRole(rdb.service.areaname, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		// read the index fields name
		var indexfields []string
		err = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if err != nil {
			tran.Rollback()
			return
		}
		// read the max increment
		var max_count uint64
		err = tran.ReadData(rdb.service.areaname, tableid, "IncrementCount", &max_count)
		if err != nil {
			tran.Rollback()
			return
		}
		// travel the increment id and delete
		for i := uint64(0); i <= max_count; i++ {
			id := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(i, 10)
			existid := tran.ExistRole(rdb.service.areaname, id)
			if existid == true {
				err = tran.DeleteRole(rdb.service.areaname, id)
				if err != nil {
					tran.Rollback()
					return
				}
			}
		}
		// travel index fields and delete
		for i := range indexfields {
			id := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + indexfields[i]
			existid := tran.ExistRole(rdb.service.areaname, id)
			if existid == true {
				err = tran.DeleteRole(rdb.service.areaname, id)
				if err != nil {
					tran.Rollback()
					return
				}
			}
		}
		// delete table's main Role
		err = tran.DeleteRole(rdb.service.areaname, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		// delete the child in TablesControl
		err = tran.DeleteChild(rdb.service.areaname, TABLE_CONTROL_NAME, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		tran.Commit()
	} else {
		db := rdb.service.drule
		var errd operator.DRuleError
		tran, errd := db.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		// log the table Role
		errd = tran.LockRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// read the index fields name
		var indexfields []string
		errd = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// read the max increment
		var max_count uint64
		errd = tran.ReadData(rdb.service.areaname, tableid, "IncrementCount", &max_count)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// travel the increment id and delete
		for i := uint64(0); i <= max_count; i++ {
			id := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(i, 10)
			existid, errd := tran.ExistRole(rdb.service.areaname, id)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			if existid == true {
				errd = tran.DeleteRole(rdb.service.areaname, id)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
			}
		}
		// travel index fields and delete
		for i := range indexfields {
			id := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + indexfields[i]
			existid, errd := tran.ExistRole(rdb.service.areaname, id)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			if existid == true {
				errd = tran.DeleteRole(rdb.service.areaname, id)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
			}
		}
		// delete table's main Role
		errd = tran.DeleteRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// delete the child in TablesControl
		errd = tran.DeleteChild(rdb.service.areaname, TABLE_CONTROL_NAME, tableid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		tran.Commit()
	}

	return
}

// Drop table if the table exist. It's alias of the DropTable.
func (rdb *RelaDB) DeleteTable(tablename string) (err error) {
	return rdb.DropTable(tablename)
}

// List all tables' name.
func (rdb *RelaDB) TableList() (list []string, err error) {
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tran, _ := db.Begin()
		var list_id []string
		list_id, err = tran.ReadChildren(rdb.service.areaname, TABLE_CONTROL_NAME)
		if err != nil {
			tran.Rollback()
			return
		}
		list = make([]string, len(list_id))
		for i := range list_id {
			var list1 string
			err = tran.ReadData(rdb.service.areaname, list_id[i], "TableName", &list1)
			if err != nil {
				tran.Rollback()
				return
			}
			list[i] = list1
		}
		tran.Commit()
	} else {
		db := rdb.service.drule
		var errd operator.DRuleError
		tran, errd := db.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		var list_id []string
		list_id, errd = tran.ReadChildren(rdb.service.areaname, TABLE_CONTROL_NAME)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		list = make([]string, len(list_id))
		for i := range list_id {
			var list1 string
			errd = tran.ReadData(rdb.service.areaname, list_id[i], "TableName", &list1)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			list[i] = list1
		}
		tran.Commit()
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

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("reladb[RelaDB]Insert: %v", e)
		}
	}()

	valuer := reflect.Indirect(reflect.ValueOf(instance))
	tname := valuer.Type().String()

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
		// 获取类型
		var ptype string
		err = tran.ReadData(rdb.service.areaname, tableid, "Prototype", &ptype)
		if err != nil {
			tran.Rollback()
			return
		}
		if ptype != tname {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Insert: The table's type is  %v but you give .", ptype, tname)
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
		// 获取类型
		var ptype string
		errd = tran.ReadData(rdb.service.areaname, tableid, "Prototype", &ptype)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if ptype != tname {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Insert: The table's type is  %v but you give .", ptype, tname)
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

// Select one Role use the id
func (rdb *RelaDB) Select(tablename string, id uint64, therole roles.Roleer) (err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("reladb[RelaDB]Select: %v", e)
		}
	}()

	valuer := reflect.Indirect(reflect.ValueOf(therole))
	tname := valuer.Type().String()

	tableid := TABLE_NAME_PREFIX + tablename
	rolesid := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(id, 10)

	// If use TRule
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tran, _ := db.Begin()
		var tabletype string
		err = tran.ReadData(rdb.service.areaname, tableid, "Prototype", &tabletype)
		if err != nil {
			tran.Rollback()
			return
		}
		if tabletype != tname {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Select: The table's type is  %v but you give .", tabletype, tname)
			return
		}
		exist := db.ExistRole(rdb.service.areaname, rolesid)
		if exist == false {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Select: Can not find the id %v.", id)
			return
		}
		err = db.ReadRole(rdb.service.areaname, rolesid, therole)
		if err != nil {
			tran.Rollback()
			return
		}
		tran.Commit()
		return
	} else {
		db := rdb.service.drule
		errd := operator.NewDRuleError()
		tran, errd := db.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		var tabletype string
		errd = tran.ReadData(rdb.service.areaname, tableid, "Prototype", &tabletype)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if tabletype != tname {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Select: The table's type is  %v but you give .", tabletype, tname)
			return
		}
		exist, errd := db.ExistRole(rdb.service.areaname, rolesid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if exist == false {
			tran.Rollback()
			err = fmt.Errorf("reladb[RelaDB]Select: Can not find the id %v.", id)
			return
		}
		errd = db.ReadRole(rdb.service.areaname, rolesid, therole)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		tran.Commit()
		return
	}
	return
}

// Select the given fields' value.
func (rdb *RelaDB) SelectFields(tablename string, id uint64, fields ...interface{}) (err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}

	fieldslen := len(fields)
	if pubfunc.IsOdd(fieldslen) == true {
		err = fmt.Errorf("reladb[RelaDB]SelectFields: The fields parameter is wrong.")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("reladb[RelaDB]SelectFields: %v", e)
		}
	}()

	//tableid := TABLE_NAME_PREFIX + tablename
	rolesid := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(id, 10)

	// If use TRule
	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		tran, _ := db.Begin()
		for i := range fields {
			tname := reflect.TypeOf(fields[i]).String()
			// 0 or even number, is field name.
			if i == 0 || pubfunc.IsOdd(i) == false {
				if tname != "string" {
					tran.Rollback()
					err = fmt.Errorf("reladb[RelaDB]SelectFields: The fields parameter is wrong.")
					return
				}
				err = tran.ReadData(rdb.service.areaname, rolesid, fields[i].(string), fields[i+1])
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
		for i := range fields {
			tname := reflect.TypeOf(fields[i]).String()
			// 0 or even number, is field name.
			if i == 0 || pubfunc.IsOdd(i) == false {
				if tname != "string" {
					tran.Rollback()
					err = fmt.Errorf("reladb[RelaDB]SelectFields: The fields parameter is wrong.")
					return
				}
				errd = tran.ReadData(rdb.service.areaname, rolesid, fields[i].(string), fields[i+1])
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

// check if have the id's column
func (rdb *RelaDB) Exist(tablename string, id uint64) (exist bool, err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}

	realid := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(id, 10)

	if rdb.service.dtype == DRULE2_USE_TRULE {
		db := rdb.service.trule
		exist = db.ExistRole(rdb.service.areaname, realid)
		return
	} else {
		db := rdb.service.drule
		var errd operator.DRuleError
		exist, errd = db.ExistRole(rdb.service.areaname, realid)
		if errd.IsError() != nil {
			err = errd.IsError()
		}
		return
	}
}

// Find all equal the request parameter's id gather(collection).
func (rdb *RelaDB) FindOr(tablename string, request ...interface{}) (gather IndexGather, err error) {
	allgather, err := rdb.find(tablename, request)
	if err != nil {
		err = fmt.Errorf("reladb[RelaDB]FindOr: %v", err)
		return
	}
	gather = make([]uint64, 0)
	for i := range allgather {
		rdb.slicesCollection(&gather, &(allgather[i]))
	}
	return
}

// Find all equal the request parameter's id gather(intersection).
func (rdb *RelaDB) FindAnd(tablename string, request ...interface{}) (gather IndexGather, err error) {
	allgather, err := rdb.find(tablename, request)
	if err != nil {
		err = fmt.Errorf("reladb[RelaDB]FindAnd: %v", err)
		return
	}
	fmt.Println(allgather)
	for i := 0; i < len(allgather)-1; i++ {
		if i == 0 {
			gather = rdb.slicesIntersection(&(allgather[i]), &(allgather[i+1]))
			i++
		} else {
			gather2 := rdb.slicesIntersection(&gather, &(allgather[i]))
			gather = gather2
		}
		//rdb.slicesIntersection(&(allgather[i]), &(allgather[i+1]))
	}
	//gather = allgather[len(allgather)-1]
	return
}

// find all equal the request parameter's id gather.
func (rdb *RelaDB) find(tablename string, request []interface{}) (gather []IndexGather, err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}

	fieldslen := len(request)
	if pubfunc.IsOdd(fieldslen) == true {
		err = fmt.Errorf("The fields parameter is wrong.")
		return
	}

	gather = make([]IndexGather, fieldslen/2)
	j := 0
	for i := range request {
		// range the request and opreate all even number
		if i == 0 || pubfunc.IsOdd(i) == false {
			indexid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + request[i].(string)
			if rdb.service.dtype == DRULE2_USE_TRULE {
				have := rdb.service.trule.ExistRole(rdb.service.areaname, indexid)
				if have == false {
					err = fmt.Errorf("The field %v not be index.", request[i].(string))
					return
				}
			} else {
				have, errd := rdb.service.drule.ExistRole(rdb.service.areaname, indexid)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
				if have == false {
					err = fmt.Errorf("The field %v not be index.", request[i].(string))
					return
				}
			}
			var vb []byte
			vb, err = iendecode.StructGobBytes(request[i+1])
			if err != nil {
				return
			}
			vbsha1 := random.GetSha1SumBytes(vb)
			var allindex map[string]IndexGather
			if rdb.service.dtype == DRULE2_USE_TRULE {
				err = rdb.service.trule.ReadData(rdb.service.areaname, indexid, "Index", &allindex)
				if err != nil {
					return
				}
			} else {
				errd := rdb.service.drule.ReadData(rdb.service.areaname, indexid, "Index", &allindex)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
			}
			one, find := allindex[vbsha1]
			if find == true {
				gather[j] = one
			} else {
				gather[j] = make([]uint64, 0)
			}
			j++
		}
	}
	return
}

// Two slices' collection, and the gather1 is the result.
func (rdb *RelaDB) slicesCollection(gather1, gather2 *IndexGather) {
	for gi := range *gather2 {
		have := false
		for gj := range *gather1 {
			if (*gather1)[gj] == (*gather2)[gi] {
				have = true
				break
			}
		}
		if have == false {
			*gather1 = append(*gather1, (*gather2)[gi])
		}
	}
	return
}

// Two slices' intersection, and the gather2 is the result.
func (rdb *RelaDB) slicesIntersection(gather1, gather2 *IndexGather) (gather IndexGather) {
	//var gather IndexGather
	gather = make([]uint64, 0)
	for gi := range *gather2 {
		for gj := range *gather1 {
			if (*gather1)[gj] == (*gather2)[gi] {
				gather = append(gather, (*gather2)[gi])
			}
		}
	}
	//gather2 = &gather
	return
}

// Delete the column if it exist
func (rdb *RelaDB) Delete(tablename string, id uint64) (err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("reladb[RelaDb]Delete: The table %v not exist.", tablename)
		return
	}

	tableid := TABLE_NAME_PREFIX + tablename
	rolesid := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(id, 10)

	// If use TRule
	if rdb.service.dtype == DRULE2_USE_TRULE {
		tran, _ := rdb.service.trule.Begin()
		err = tran.LockRole(rdb.service.areaname, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		exist := tran.ExistRole(rdb.service.areaname, rolesid)
		if exist == false {
			tran.Rollback()
			return
		}
		// get the Role's middle data
		var mid roles.RoleMiddleData
		mid, err = tran.ReadRoleMiddleData(rdb.service.areaname, rolesid)
		if err != nil {
			tran.Rollback()
			return
		}
		// get the index from table main Role.
		var indexfields []string
		err = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if err != nil {
			tran.Rollback()
			return
		}
		// operate the index field
		for _, onefield := range indexfields {
			var thea string
			thea = random.GetSha1SumBytes(mid.Data.Point[onefield])
			indexid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + onefield
			// get the index
			var indexc map[string]IndexGather
			err = tran.ReadData(rdb.service.areaname, indexid, "Index", &indexc)
			if err != nil {
				tran.Rollback()
				return
			}
			// find if have the index value
			gather, find := indexc[thea]
			if find == true {
				var count int
				for i, v := range gather {
					if v == id {
						count = i
						break
					}
				}
				gather = append(gather[:count], gather[count+1:]...)
				indexc[thea] = gather
			}
			// restore the gather
			err = tran.WriteData(rdb.service.areaname, indexid, "Index", indexc)
			if err != nil {
				tran.Rollback()
				return
			}
		}
		// delete the column
		err = tran.DeleteRole(rdb.service.areaname, rolesid)
		if err != nil {
			tran.Rollback()
			return
		}
		tran.Commit()
	} else {
		var errd operator.DRuleError
		tran, errd := rdb.service.drule.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		errd = tran.LockRole(rdb.service.areaname, tableid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		exist, errd := tran.ExistRole(rdb.service.areaname, rolesid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if exist == false {
			tran.Rollback()
			return
		}
		// get the Role's middle data
		var mid roles.RoleMiddleData
		mid, errd = tran.ReadRoleToMiddleData(rdb.service.areaname, rolesid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// get the index from table main Role.
		var indexfields []string
		errd = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		// operate the index field
		for _, onefield := range indexfields {
			var thea string
			thea = random.GetSha1SumBytes(mid.Data.Point[onefield])
			indexid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + onefield
			// get the index
			var indexc map[string]IndexGather
			errd = tran.ReadData(rdb.service.areaname, indexid, "Index", &indexc)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			// find if have the index value
			gather, find := indexc[thea]
			if find == true {
				var count int
				for i, v := range gather {
					if v == id {
						count = i
						break
					}
				}
				gather = append(gather[:count], gather[count+1:]...)
				indexc[thea] = gather
			}
			// restore the gather
			errd = tran.WriteData(rdb.service.areaname, indexid, "Index", indexc)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
		}
		// delete the column
		errd = tran.DeleteRole(rdb.service.areaname, rolesid)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		tran.Commit()
	}

	return
}

// Update fields data.
func (rdb *RelaDB) UpdateFields(tablename string, id uint64, parameter ...interface{}) (err error) {
	// check the table if exist.
	var have bool
	have, err = rdb.TableExist(tablename)
	if err != nil {
		return
	}
	if have == false {
		err = fmt.Errorf("The table %v not exist.", tablename)
		return
	}
	// check the fields parameter's number.
	fieldslen := len(parameter)
	if pubfunc.IsOdd(fieldslen) == true {
		err = fmt.Errorf("reladb[RelaDB]UpdateFields: The parameter is wrong.")
		return
	}
	// recover panic
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("reladb[RelaDB]UpdateFields: %v", e)
		}
	}()
	tableid := TABLE_NAME_PREFIX + tablename
	rolesid := TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + strconv.FormatUint(id, 10)
	if rdb.service.dtype == DRULE2_USE_TRULE {
		tran, _ := rdb.service.trule.Begin()
		// lock the table Role.
		err = tran.LockRole(rdb.service.areaname, tableid)
		if err != nil {
			tran.Rollback()
			return
		}
		// check if the id exist
		have := tran.ExistRole(rdb.service.areaname, rolesid)
		if have == false {
			err = fmt.Errorf("The id %v not exist.", id)
			return
		}
		// get all the index fields name.
		var indexfields []string
		err = tran.ReadData(rdb.service.areaname, tableid, "IndexField", &indexfields)
		if err != nil {
			tran.Rollback()
			return
		}
		// get the Role's middle data.
		var mid roles.RoleMiddleData
		mid, err = tran.ReadRoleMiddleData(rdb.service.areaname, rolesid)
		if err != nil {
			tran.Rollback()
			return
		}
		// for range parameter
		for i := range parameter {
			tname := reflect.TypeOf(parameter[i]).String()
			if i == 0 || pubfunc.IsOdd(i) == false {
				if tname != "string" {
					tran.Rollback()
					err = fmt.Errorf("reladb[RelaDB]UpdateFields: The fields parameter is wrong.")
					return
				}
				fieldname := parameter[i].(string)
				// check if the column Role have the field
				old_valb, have := mid.Data.Point[fieldname]
				if have == false {
					tran.Rollback()
					err = fmt.Errorf("reladb[RelaDB]UpdateFields: The fields parameter is wrong.")
					return
				}
				var new_valb []byte
				new_valb, err = iendecode.StructGobBytes(parameter[i+1])
				if err != nil {
					tran.Rollback()
					return
				}
				// check if the field be index
				if rdb.stringInSlice(indexfields, fieldname) == true {
					// change the index
					err = rdb.changeOneFieldIndexTRule(tran, tablename, id, fieldname, old_valb, new_valb)
					if err != nil {
						tran.Rollback()
						return
					}
				}
				// change the field's value
				mid.Data.Point[fieldname] = new_valb
			}
		}
		// restore the column Role.
		err = tran.StoreRoleFromMiddleData(rdb.service.areaname, mid)
		if err != nil {
			tran.Rollback()
			return
		}
		tran.Commit()
	} else {

	}
	return
}

// check if the string in the slice.
func (rdb *RelaDB) stringInSlice(slice []string, one string) (yes bool) {
	yes = false
	for _, s := range slice {
		if s == one {
			yes = true
			return
		}
	}
	return
}

// change one field's index
func (rdb *RelaDB) changeOneFieldIndexTRule(tran *trule.Transaction, tablename string, id uint64, fieldname string, oldb, newb []byte) (err error) {
	indexid := TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
	olda := random.GetSha1SumBytes(oldb)
	newa := random.GetSha1SumBytes(newb)
	// read the index
	var indexc map[string]IndexGather
	err = tran.ReadData(rdb.service.areaname, indexid, "Index", &indexc)
	if err != nil {
		return
	}
	// delete index from old value
	oldi, have := indexc[olda]
	if have == true {
		var count int
		for i, v := range oldi {
			if v == id {
				count = i
				break
			}
		}
		oldi = append(oldi[:count], oldi[count+1:]...)
		indexc[olda] = oldi
	}
	// add index from new value
	newi, have := indexc[newa]
	if have == true {
		newi = append(newi, id)
		indexc[newa] = newi
	} else {
		newi = []uint64{id}
		indexc[newa] = newi
	}
	// restore the index
	err = tran.WriteData(rdb.service.areaname, indexid, "Index", indexc)
	if err != nil {
		return
	}
	return
}
