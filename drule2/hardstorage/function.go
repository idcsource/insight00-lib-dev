// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package hardstorage

import "regexp"

// 检查区域的名称是否合法，只接受英文数字等，长度则在40个字符以内
func CheckAreaName(name string) (allow bool) {
	check, _ := regexp.Compile("^[A-Za-z0-9_-]{1,40}$")
	allow = check.MatchString(name)
	return
}
