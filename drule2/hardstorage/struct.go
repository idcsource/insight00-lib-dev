// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstorage

const (
	HARDSTORAGE_FILE_NAME_DATA     = "_data"
	HARDSTORAGE_FILE_NAME_RELATION = "_relation"
)

// 存储器类型
type HardStorage struct {
	local_path string // 本地路径
	path_deep  uint8  // 路径深度
}
