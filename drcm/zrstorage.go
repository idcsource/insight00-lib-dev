// Copyright 2016-2017
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
	if z.cacheMax <= 0 {
		return
	}
	z.rolesCache = make(map[string]*oneRoleCache)
	z.deleteCache = make([]string, 0)
	z.cacheIsFull = make(chan bool)
	z.rolesCount = 0
	z.checkCacheNumOn = false
	go z.runtimeStore()
	return
}

// 运行时自动存储
func (z *ZrStorage) runtimeStore() {
	for {
		full := <-z.cacheIsFull
		if full == true {
			// 保存
			z.toCacheStore()
		}
	}
}

func (z *ZrStorage) deferCheckCacheNumOn() {
	z.checkCacheNumOn = false
}

// 缓存的保存
func (z *ZrStorage) toCacheStore() (err error) {
	z.lock.Lock()
	defer z.lock.Unlock()

	z.checkCacheNumOn = true
	defer z.deferCheckCacheNumOn()

	for _, rolec := range z.rolesCache {
		err = z.local_store.StoreRole(rolec.role)
		if err != nil {
			z.logerr(err)
		}
	}

	// 重置缓存
	z.rolesCache = make(map[string]*oneRoleCache)
	z.deleteCache = make([]string, 0)
	z.rolesCount = 0
	var errstring string
	for _, onec := range z.slavecpool {
		// 分配连接
		cprocess := onec.tcpconn.OpenProgress()
		defer cprocess.Close()
		// 发送前导
		slave_receipt, err := SendPrefixStat(cprocess, onec.code, OPERATE_TOSTORE)
		if err != nil {
			z.logerr(err)
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
		if slave_receipt.DataStat != DATA_ALL_OK {
			z.logerr(slave_receipt.Error)
			errstring += fmt.Sprint(onec.name, ": ", err, " | ")
		}
	}
	if len(errstring) != 0 {
		return fmt.Errorf(errstring)
	}
	return nil
}

// 缓存的运行时保存
func (z *ZrStorage) ToStore() (err error) {
	err = z.toCacheStore()
	if err != nil {
		return err
	}

	// 向所有的slave发送
	if z.dmode != DMODE_OWN {
		for _, onec := range z.slavecpool {
			// 分配连接
			cprocess := onec.tcpconn.OpenProgress()
			defer cprocess.Close()
			// 发送前导OPERATE_TOSTORE
			slave_receipt, err := z.sendPrefixStat(cprocess, onec.code, OPERATE_TOSTORE)
			if err != nil {
				z.logerr(err)
				return err
			}
			if slave_receipt.DataStat != DATA_ALL_OK {
				z.logerr(slave_receipt.Error)
				return fmt.Errorf(slave_receipt.Error)
			}
		}
	}
	return nil
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
	z.slavepool = make(map[string]*nst.TcpClient)
	z.slavecpool = make(map[string]*slaveIn)

	// 获取slave的配置名
	slaves, err := z.config.GetEnum("main.slave")
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
	// 遍历所有的slave配置名
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
		//z.slavepool = make(map[string]*nst.TcpClient)
		sconn, err := nst.NewTcpClient(addr, conn_num, z.logs)
		if err != nil {
			err = fmt.Errorf("drcm:NewZrStorage: %v", err)
			z.closeSlavePool()
			return err
		}
		z.slavepool[one] = sconn
		z.slavecpool[one] = &slaveIn{
			name:    one,
			code:    code,
			tcpconn: sconn,
		}
		// 遍历可管理角色首字母创建连接序列
		for _, onewho := range control_whos {
			// 序列里没有这个字母就建立一个
			if _, have := z.slaves[onewho]; have == false {
				z.slaves[onewho] = make([]*slaveIn, 0)
			}
			// 将这个字母的序列中加入这个slave的名字
			z.slaves[onewho] = append(z.slaves[onewho], z.slavecpool[one])
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

// 检查缓存数，如果超出则执行运行时保存
func (z *ZrStorage) checkCacheNum() {
	if z.cacheMax > 0 && z.rolesCount >= z.cacheMax && z.checkCacheNumOn == false {
		z.cacheIsFull <- true
	}
}

// 处理错误日志
func (z *ZrStorage) logerr(err interface{}) {
	if err == nil {
		return
	}
	if z.logs != nil {
		z.logs.ErrLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (z *ZrStorage) logrun(err interface{}) {
	if err == nil {
		return
	}
	if z.logs != nil {
		z.logs.RunLog(fmt.Errorf("drcm[ZrStorage]: %v", err))
	} else {
		fmt.Println(err)
	}
}
