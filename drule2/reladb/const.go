// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package reladb

const (
	TABLE_NAME_PREFIX            = "_Reladb_Table_" // 表名前缀，一个表的总表角色Id是 TABLE_NAME_PREFIX + name
	TABLE_CONTENT_PREFIX         = "_Content_"      // 一个内容角色的Id是 TABLE_NAME_PREFIX + name + TABLE_CONTENT_PREFIX + id
	TABLE_INDEX_NAME             = "_INDEX_"        // 内容索引的角色id是 TABLE_NAME_PREFIX + name + TABLE_INDEX_NAME + 数字（指明这个索引开始的数字是什么）
	TABLE_ONE_INDEX_COUNT uint64 = 1000             // 一个索引的角色管理多少个索引
)
