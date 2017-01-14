// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"fmt"
	"sync"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
)

// 创建一个锆存储，config的实例见源代码的zrstorage.cfg
func NewZrStorage(config *cpool.Block, logs *ilogs.Logs) (z *ZrStorage, err error) {
	z = &ZrStorage{
		RolePlus: rolesplus.RolePlus{
			Role: roles.Role{},
		},
		config: config,
		logs:   logs,
		lock:   new(sync.RWMutex),
	}
	z.New(random.Unid(1, "ZrStorage"))
	// 处理运行的模式
	mode, err := config.GetConfig("main.mode")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	switch mode {
	case "own":
		z.dmode = DMODE_OWN
		// 使用own模式来启动锆存储
		err = z.startUseOwn()
	case "master":
		z.dmode = DMODE_MASTER
		// 使用master模式来启动锆存储
		err = z.startUseMaster()
	case "slave":
		z.dmode = DMODE_SLAVE
		// 使用slave模式来启动锆存储
		err = z.startUseSlave()
	default:
		err = fmt.Errorf("drcm:NewZrStorage: the mode config must own, master or slave.")
		return
	}
	return
}

// 创建缓存
func (z *ZrStorage) buildCache() (err error) {
	z.cacheMax, err = z.config.TranInt64("local.cache_num")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	z.rolesCache = make(map[string]*oneRoleCache)
	z.deleteCache = make([]string, 0)
	z.cacheIsFull = make(chan bool)
	z.rolesCount = 0
	z.checkCacheNumOn = false
	return
}

// 使用own模式来启动锆存储
func (z *ZrStorage) startUseOwn() (err error) {
	// 创建本地存储
	z.local_store, err = hardstore.NewHardStore(z.config)
	if err != nil {
		return
	}
	// 创建缓存
	err = z.buildCache()
	return
}

// 使用slave模式来启动锆存储，也就是启动一个tcp的监听（但先要启用本地存储）
func (z *ZrStorage) startUseSlave() (err error) {
	err = z.startUseOwn()
	if err != nil {
		return
	}
	port, err := z.config.GetConfig("main.port")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	z.listen = nst.NewTcpServer(z, port, z.logs)
	z.code, err = z.config.GetConfig("main.code")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	return
}

// 使用master模式来启动锆存储，也就是连接所有的slave，当然也要启动监听和本地存储
func (z *ZrStorage) startUseMaster() (err error) {
	err = z.startUseSlave()
	if err != nil {
		return
	}
	z.slaves = make(map[string][]*slaveIn)
	// 获取slave的配置名
	slaves, err := z.config.GetEnum("main.slave")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	// 遍历所以的slave配置名
	for _, one := range slaves {
		// 获取每个slave的配置
		onecfg, err := z.config.GetSection(one)
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		// 获取这个slave可管理的角色首字母
		control_whos, err := onecfg.GetEnum("control")
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		// 获取连接数
		var conn_num int
		conn_num64, err := onecfg.TranInt64("conn_num")
		if err != nil {
			conn_num = 1
		} else {
			conn_num = int(conn_num64)
		}
		// 获取身份验证码
		code, err := onecfg.GetConfig("code")
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		// 获取连接地址
		addr, err := onecfg.GetConfig("address")
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		// 创建连接和连接池，放到池子里主要是为了到时候出错了关闭方便
		z.slavepool = make(map[string]*nst.TcpClient)
		sconn, err := nst.NewTcpClient(addr, conn_num, z.logs)
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		z.slavepool[one] = sconn
		// 遍历可管理角色首字母创建连接序列
		for _, onewho := range control_whos {
			// 序列里没有这个字母就建立一个
			if _, have := z.slaves[onewho]; have == false {
				z.slaves[onewho] = make([]*slaveIn, 0)
			}
			// 将这个字母的序列中加入这个slave的名字
			z.slaves[onewho] = append(z.slaves[onewho], &slaveIn{
				name:    one,
				code:    code,
				tcpconn: sconn,
			})
		}
	}
	return
}

// 关闭整个slavepool
func (z *ZrStorage) closeSlavePool() {
	for _, conn := range z.slavepool {
		conn.Close()
	}
}

// ExecTCP nst的ConnExecer接口
func (z *ZrStorage) ExecTCP(tcp *nst.TCP) (err error) {
	// 接收身份验证码
	// 回应是否可以传输数据
	// 接收指令
	// 回应可以接收数据
	// 转到相应方法
	return nil
}

// 检查缓存数，如果超出则执行运行时保存
func (z *ZrStorage) checkCacheNum() {
	if z.cacheMax > 0 && z.rolesCount >= z.cacheMax && z.checkCacheNumOn == false {
		z.cacheIsFull <- true
	}
}
