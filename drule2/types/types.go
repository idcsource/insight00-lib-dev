// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package types

import (
	"encoding/gob"
	"time"

	"github.com/idcsource/insight00-lib/cpool"
	"github.com/idcsource/insight00-lib/roles"
)

func init() {
	// time
	gob.Register(time.Time{})

	// roles
	gob.Register(roles.RoleMiddleData{})

	// cpool
	gob.Register(cpool.PoolEncode{})
	gob.Register(cpool.BlockEncode{})
	gob.Register(cpool.SectionEncode{})
	gob.Register(cpool.ConfigEncode{})
}
