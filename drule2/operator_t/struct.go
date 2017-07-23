// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// drule2的远程控制者控制台终端
//
// 具体命令如下：
//	area list
//	area add 'area_name'
//	area delete 'area_name'
//	area rename 'old_name' to 'new_name'
package operator_t

import (
	"regexp"

	"github.com/idcsource/insight00-lib/drule2/operator"
)

type OperatorT struct {
	operator *operator.Operator
	regexp   map[string]*regexp.Regexp
}
