// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// spots : The 2nd generation Roles.
package spots

const (
	// Father be changed
	FATHER_CHANGED = iota
	// Children be changed
	CHILDREN_CHANGED
	// friends be changed
	FRIENDS_CHANGED
	// self be changed, all top three
	SELF_CHANGED
	// data be changed
	DATA_CHANGED
	// context be changed
	CONTEXT_CHANGED
)

type ContextUpDown uint8

const (
	// 上下文上游
	CONTEXT_UP ContextUpDown = iota
	// 上下文下游
	CONTEXT_DOWN
)

type StatusValueType uint8

const (
	// 状态位的值类型：null
	STATUS_VALUE_TYPE_NULL StatusValueType = iota
	// 状态位的值类型：int64
	STATUS_VALUE_TYPE_INT
	// 状态位的值类型：float64
	STATUS_VALUE_TYPE_FLOAT
	// 状态位的值类型：complex128
	STATUS_VALUE_TYPE_COMPLEX
	// 状态位的值类型：string
	STATUS_VALUE_TYPE_STRING
)
