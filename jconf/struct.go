// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

// The configure use JSON

package jconf

type JconfDtype uint8

const (
	JSON_CONF_DATA_TYPE_STRING JconfDtype = iota
	JSON_CONF_DATA_TYPE_INT
	JSON_CONF_DATA_TYPE_FLOAT
	JSON_CONF_DATA_TYPE_BOOL
	JSON_CONF_DATA_TYPE_ENUM
	JSON_CONF_DATA_TYPE_NODE
)

type JsonConf struct {
	dtype  JconfDtype
	String string // the value, string
	Int    int64
	Float  float64
	Bool   bool
	enum   []string // the enum
	Node   map[string]*JsonConf
}
