// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spots

type DataBodyer interface {
	//Set(name string, data interface{}) error
	//Get(name string, data interface{}) error
	DecodeBbody(b map[string][]byte) error
	EncodeBbody() (b map[string][]byte, err error)
}
