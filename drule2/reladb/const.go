// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package reladb

const (
	TABLE_CONTROL_NAME                   = "_Reladb_Table_Control" // 表管理的角色id
	TABLE_NAME_PREFIX                    = "_Reladb_Table_"        // 表名前缀，一个表的总表角色Id是 TABLE_NAME_PREFIX + tablename
	TABLE_AUTOINCREMENT_NAME             = "_AUTOINCREMENT_"       // 内容自增字段的角色id是 TABLE_NAME_PREFIX + tablename + TABLE_AUTOINCREMENT_NAME + 数字（指明这个索引开始的数字是什么）
	TABLE_ONE_AUTOINCREMENT_COUNT uint64 = 1000                    // 一个自增的角色管理多少个自增
	TABLE_INDEX_PREFIX                   = "_INDEX_"               // 索引角色的id是 TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
	TABLE_COLUMN_PREFIX                  = "_COLUMN_"              // 每个保存的条目角色id是  TABLE_NAME_PREFIX + tablename + TABLE_COLUMN_PREFIX + 数字(自增数字)
)

type DRule2Type uint8 // 使用的drule2的类型
const (
	DRULE2_USE_NO    DRule2Type = iota // 空
	DRULE2_USE_TRULE                   // 使用drule2的trule
	DRULE2_USE_DRULE                   // 使用drule2的drule（operator）
)

type FieldType uint8 // 字段类型
const (
	FIELD_TYPE_NO     FieldType = iota // 空
	FIELD_TYPE_STRING                  // 字符串
	FIELD_TYPE_INT                     // 数字
	FIELD_TYPE_FLOAT                   // 浮点
	FIELD_TYPE_BOOL                    // 布尔
	FIELD_TYPE_TIME                    // 时间
)
