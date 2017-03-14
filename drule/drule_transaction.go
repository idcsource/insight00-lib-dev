// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"reflect"

	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 查看连接是哪个，id为角色的id，connmode来自CONN_IS_*
func (d *DRuleTransaction) findConn(id string) (connmode uint8, conn []*slaveIn) {
	// 如果模式为own，则直接返回本地
	if d.connect.dmode == DMODE_OWN {
		connmode = CONN_IS_LOCAL
		return
	}

	// 找到第一个首字母。
	theChar := string(id[0])
	// slave池中有没有
	conn, find := d.connect.slaves[theChar]
	if find == false {
		// 如果在slave池里没有找到，那么就默认为本地存储
		connmode = CONN_IS_LOCAL
		return
	} else {
		connmode = CONN_IS_SLAVE
		return
	}
}

// 判断friend或context的状态的类型，types：1为int，2为float，3为complex
func (d *DRuleTransaction) statusValueType(value interface{}) (types uint8) {
	valuer := reflect.Indirect(reflect.ValueOf(value))
	vname := valuer.Type().String()
	switch vname {
	case "int64":
		return roles.STATUS_VALUE_TYPE_INT
	case "float64":
		return roles.STATUS_VALUE_TYPE_FLOAT
	case "complex128":
		return roles.STATUS_VALUE_TYPE_COMPLEX
	default:
		return roles.STATUS_VALUE_TYPE_NULL
	}
}
