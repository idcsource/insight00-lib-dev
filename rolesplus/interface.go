// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package rolesplus

import (
	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// RolesPlus的接口
type RolePluser interface {
	roles.Roleer
	ReturnBridges () map[string]*bridges.BridgeBind
	ReturnLog () *ilogs.Logs
	SetLog (logs *ilogs.Logs)
	BridgeBind (name string, br *bridges.Bridge)
	BridgeSend (name string, data bridges.BridgeData)
	ErrLog (err interface{})
	RunLog (err interface{})
	ReRunLog () []string
	ReErrLog () []string
	ExecTCP (tcp *nst.TCP)
}
