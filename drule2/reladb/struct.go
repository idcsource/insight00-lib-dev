// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package reladb

import (
	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 表管理，目前来说就是一个用child记录表名的功能 id:TABLE_CONTROL_NAME
type TablesControl struct {
	roles.Role
}

// 一个表的总角色 id:TABLE_NAME_PREFIX + tablename
type TableMain struct {
	roles.Role
	TableName      string   // 表名
	Prototype      string   // 角色原型的名称，使用反射得到
	IncrementCount uint64   // 增量计数（增量到哪里了）
	IndexField     []string // 需要索引的字符串
}

// 自增字段的管理 id:TABLE_NAME_PREFIX + tablename + TABLE_AUTOINCREMENT_NAME + 数字（指明这个索引开始的数字是什么）
type TableAutoIncrement struct {
	roles.Role
	Index map[string]string
}

type IndexGather []uint64 // 索引的合集

// 管理索引的角色 id:TABLE_NAME_PREFIX + tablename + TABLE_INDEX_PREFIX + fieldname
// 目前只提供对string、int64、float64、bool、time.Time类型的索引
type TableIndex struct {
	roles.Role
	FieldName string                 // 索引字段的名字
	FieldType FieldType              // 字符串类型
	Index     map[string]IndexGather // 索引，[string]为索引的内容值
}

// 关系型数据服务
type reladbService struct {
	dtype    DRule2Type         // 使用trule还是operator连接远程服务器
	trule    *trule.TRule       // 如果使用trule
	drule    *operator.Operator // 如果使用drule
	areaname string             // 涉及到的区域名称（在这里区域就类似于数据库了）
}

// 关系型数据
type RelaDB struct {
	service *reladbService // 服务
}
