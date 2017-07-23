// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"github.com/idcsource/insight00-lib/ilogs"
)

// 这叫做“操作机”，是用来远程连接DRule的。
type Operator struct {
	// 自己的名字
	selfname string
	// 是否在事务里
	inTransaction bool
	// 事务的ID
	transactionId string
	// 服务器端，slaveIn中的name是服务器的地址和端口字符串
	slaves []*slaveIn
	// 日志
	logs *ilogs.Logs
}
