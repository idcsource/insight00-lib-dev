// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// The RelaDB is a simulation of relational database.
//
// The lower layer, you can use drule2/trule or drule2/operator.
// It does not support SQL, only provides some simple functions. Include create table, index fields, insert, select, update, delete...
package reladb

const (
	TABLE_CONTROL_NAME                   = "_Reladb_Table_Control" // The table control role's id
	TABLE_NAME_PREFIX                    = "_Reladb_Table_"        // The table id's prefix (One table's main Role's id is: TABLE_NAME_PREFIX + tablename)
	TABLE_AUTOINCREMENT_NAME             = "_AUTOINCREMENT_"       // The auto increment's name (The Role's id is :TABLE_NAME_PREFIX + tablename + TABLE_AUTOINCREMENT_NAME + count)
	TABLE_ONE_AUTOINCREMENT_COUNT uint64 = 1000                    // One auto increment's Role can manange how many count.
	TABLE_INDEX_PREFIX                   = "_INDEX_"               // The index field Role id's prefix (The index role's id is: TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname)
	TABLE_COLUMN_PREFIX                  = "_COLUMN_"              // Column Role id's prefix (The column's id is: TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + count)
)

type DRule2Type uint8 // Use DRule2's what
const (
	DRULE2_USE_NO    DRule2Type = iota // none
	DRULE2_USE_TRULE                   // use drule2/trule
	DRULE2_USE_DRULE                   // use drule2/operator
)

type FieldType uint8 // Field's data type
const (
	FIELD_TYPE_NO     FieldType = iota // none
	FIELD_TYPE_STRING                  // string
	FIELD_TYPE_INT                     // int64
	FIELD_TYPE_FLOAT                   // float64
	FIELD_TYPE_BOOL                    // bool
	FIELD_TYPE_TIME                    // time
)
