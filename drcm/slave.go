// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"sync"
	"fmt"
	
	"github.com/idcsource/Insight-0-0-lib/rolesio"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 新建SlaveController
func NewSlaveController (ro rolesio.RolesInOutManager, max int64, logs *ilogs.Logs) *SlaveController {
	sc := &SlaveController{
		readWrite: ro,
		rolesCache : make(map[string]roles.Roleer),
		deleteCache: make([]string,0),
		cacheMax : max,
		rolesCount : 0,
		cacheIsFull : make(chan bool),
		logs : logs,
		lock: new(sync.RWMutex),
		checkCacheNumOn: false,
	};
	if sc.cacheMax > 0 {
		go rb.runtimeStore();
	}
	return sc;
}

// 实现TCPServer的ConnExecer接口
func (s *SlaveController) ExecTCP (tcp *nst.TCP) {
	defer tcp.Close();
	for {
		oprate , err1 := tcp.GetStat();
		if err1 != nil {
			if fmt.Sprint(err1) == "EOF" {
				return;
			} else {
				s.logs.ErrLog("drcm [SlaveController]ExecTCP: ",err1) ;
				return; 
			}
		}
		databody, err2 := tcp.GetData();
		if err2 != nil {
			if fmt.Sprint(err1) == "EOF" {
				return;
			} else {
				s.logs.ErrLog("drcm [SlaveController]ExecTCP: ",err2) ;
				return; 
			}
		}
		s.oprateSeparate(tcp, oprate, databody);
	}
}

// 配置分离，将oprate给出的配置项交给不同的函数进行处理，而TCP连接将会一直传递下去
func (s *SlaveController) oprateSeparate (tcp *nst.TCP, oprate uint8, databody []byte) {
	
}

func (s *SlaveController) runtimeStore() {
	
}
