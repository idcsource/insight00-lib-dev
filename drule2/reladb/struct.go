// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package reladb

import (
	"github.com/idcsource/insight00-lib/drule2/operator"
	"github.com/idcsource/insight00-lib/drule2/trule"
	"github.com/idcsource/insight00-lib/roles"
)

// The tables control Role. id:TABLE_CONTROL_NAME
type TablesControl struct {
	roles.Role
}

// A table's main Role. id:TABLE_NAME_PREFIX + tablename
type TableMain struct {
	roles.Role
	TableName      string   // Table's name
	Prototype      string   // Role's prototype name
	IncrementCount uint64   // Auto increment's count
	IncrementField string   // Auto increment's field
	IndexField     []string // The Fields whitch need be index
}

// The auto increment Role. id:TABLE_NAME_PREFIX + tablename + TABLE_AUTOINCREMENT_NAME + count
type TableAutoIncrement struct {
	roles.Role
	Index map[string]string
}

type IndexGather []uint64 // The index id(increment) gather

// The index field Role. id:TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
type TableIndex struct {
	roles.Role
	FieldName string                 // The field name.
	FieldType FieldType              // The field's data type.
	Index     map[string]IndexGather // The index, [string] is index's value
}

// The RelaDB's service
type reladbService struct {
	dtype    DRule2Type         // Use trule or operator
	trule    *trule.TRule       // if use trule
	drule    *operator.Operator // if use drule
	areaname string             // the area name, like database name
}

// RelaDb
type RelaDB struct {
	service *reladbService // the reladb's service
}
