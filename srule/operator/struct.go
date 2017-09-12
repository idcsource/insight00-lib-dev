// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator

import (
	"sync"
	"time"

	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/nst2"
)

type tranService struct {
	unid   string            // 事务ID
	askfor TransactionAskFor // 请求操作
}

type operatorService struct {
	tran_signal chan tranService // 事务信号
}

// 这叫做“操作机”，是用来远程连接DRule的。
type Operator struct {
	selfname         string                   // 自己的名字
	drule            *druleInfo               // 服务器端
	login            bool                     // 是否在登陆状态，如果不是则是false
	service          *operatorService         // 操作者服务
	transaction      map[string]*OTransaction // 事务列表，time是事务的活跃时间
	transaction_lock *sync.RWMutex            // transaction的锁
	logs             *ilogs.Logs              // 日志
	runstatus        uint8                    // 运行状态，OPERATOR_RUN_*
	closeing_signal  chan bool                // 正在停止信号
	closed_signal    chan bool                // 已经停止信号
	tran_wait        *sync.WaitGroup          // 事务等待计数
}

// 操作机事务
type OTransaction struct {
	selfname       string           // 自己的名字
	transaction_id string           // 事务id
	drule          *druleInfo       // 服务器端
	service        *operatorService // 操作者服务
	logs           *ilogs.Logs      // 日志
	bedelete       bool             // 如果为true则被删除
	activetime     time.Time        // 活跃日期
}

// 一台服务器的信息
type druleInfo struct {
	name        string       // 机器名称
	username    string       // 用户名
	password    string       // 密码
	unid        string       // 登录唯一码
	active_time time.Time    // 活跃日期
	tcpconn     *nst2.Client // tcp连接
}
