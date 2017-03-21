// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"github.com/idcsource/Insight-0-0-lib/rolesio"
)

// DRule的接口。TRule和Operator都符合。
//
// 同时DRuler接口完全符合rolesio.RolesInOutManager接口，但其中的ToStore()并不做实现
type DRuler interface {
	rolesio.RolesInOutManager

	Begin() (tran *Transaction)
	Transaction() (tran *Transaction)
	Prepare(roleids ...string) (tran *Transaction, err error)
}
