// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package rolesplus

import (
	"github.com/idcsource/insight00-lib/bridges"
	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/nst"
	"github.com/idcsource/insight00-lib/roles"
)

// RolesPlus的接口
type RolePluser interface {
	roles.Roleer
	ReturnBridges() map[string]*bridges.BridgeBind
	ReturnLog() *ilogs.Logs
	SetLog(logs *ilogs.Logs)
	BridgeBind(name string, br *bridges.Bridge)
	BridgeSend(name string, data bridges.BridgeData)
	ErrLog(err interface{})
	RunLog(err interface{})
	ReRunLog() []string
	ReErrLog() []string
	ExecTCP(ce *nst.ConnExec) error
}
