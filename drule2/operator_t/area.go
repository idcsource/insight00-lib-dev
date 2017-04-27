// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator_t

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
)

// area的命令执行器，从split[1]开始
func (o *OperatorT) execArea(split []string, l int) (r []string, err error) {
	errs := operator.NewDRuleError()
	// area list
	if split[1] == "list" {
		r, errs = o.operator.AreaList()
		err = errs.IsError()
		return
	}
	if l < 3 {
		err = fmt.Errorf("Command syntax error.")
		return
	}

	switch split[1] {
	case "add":
		// area add 'area name'
		errs = o.operator.AreaAdd(split[2])
		err = errs.IsError()
	case "delete":
		// area delete 'area name'
		errs = o.operator.AreaDel(split[2])
		err = errs.IsError()
	case "rename":
		// area rename 'old name' to 'new name'
		if l < 5 {
			err = fmt.Errorf("Command syntax error.")
		} else {
			errs = o.operator.AreaRename(split[2], split[4])
			err = errs.IsError()
		}
	default:
		err = fmt.Errorf("Command syntax error.")
	}
	return
}
