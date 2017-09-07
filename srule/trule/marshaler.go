// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"github.com/idcsource/insight00-lib/spots"
)

type BbodyMarshaler interface {
	spots.DataBodyer
	BbodyMarshel(name string, data interface{}) (b []byte, err error)
	BbodyUnmarshaler(name string, b []byte, data interface{}) (err error)
}
