// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// D.R.C.M.为Distributed Roles Control Machine（分布式角色控制机）的缩写。
//
// D.R.C.M.提供了针对角色接口（roles.Roleer）的分布式跨存储跨服务器的储存方式。
//
// ZrStorage
//
// ZrStorage又为“锆存储”，它实现了rolesio.RolesInOutManager接口的全部功能，可以对角色内关系和数据进行完全控制。
//
// 在底层，ZrStorage使用HardStore进行本地存储。其自身则内建缓存机制与读写锁机制，并增加ToStore()方法提供自由的运行时保存功能。
//
// ZrStorage包括三种运行模式：own、master、slave。own为本地模式，只是一个带缓存的本地存储。master和slave模式则构成了“主-从”分布式存储。如果单独使用slave模式则可以构建普通的角色存储服务器。
// "主-从"模式下，支持根据角色名的第一个字母进行路由，且支持镜像，配置方法见配置文件。
//
// ZrStorage需要提供一个*cpool.Block类型的配置信息，具体实力可参见源代码中drcm包的zrstorage.cfg文件。
//	{zr_storage}
//
//	[main]
//	# mode可以是own、master、slave
//	mode = master
//	# 作为服务监听的端口，master和slave必须
//	port = 9999
//	# 自己的身份验证码，访问者需要提供这串代码来获得操作权限，master和slave必须
//	code = the000Master
//	# slave配置的名字，用逗号分割，master必须
//	slave = s001,s002
//
//	# local是本地的存储设置，包含缓存和底层HardStore的设置
//	[local]
//	# 角色的内存缓存大小
//	cache_num = 1000
//	# 本地的存储位置
//	path = /pathname/
//	# 本地存储的路径层深
//	path_deep = 2
//
//	[s001]
//	# 路由方案，角色名的首字母的列表，设置表明这些字母开头的角色将交由这台slave处理
//	control = a,b,c,d,e,f,1,2,3
//	# master与slave的连接数
//	conn_num = 5
//	# slave的身份验证码
//	code = slave_code
//	# slave的地址和端口
//	address = 192.168.1.101:11111
//
//	[s002]
//	# 与s001重复的字母将规整为镜像，执行读操作的时候随机选取，进行写操作的时候同时执行
//	control = 0,1,2,3,4,5,6,7,8,9
//	conn_num = 5
//	code = codecodecodecode
//	address = 192.168.1.102:11111
//
// Operator
//
// Operator为ZrStorage的“操作机”。它同样实现了rolesio.RolesInOutManager接口，但其作用仅是用来操作远程ZrStorage的master或slave，并支持镜像管理。
//
// Operator在新建时需要提供一台Zrstorage的相关信息，包括地址、连接数、身份验证码。之后如果需要进行镜像管理则可以通过AddServer()方法添加。
// Operator对所有镜像同等对待，没有路由规则，读操作的时候随即选取，写操作的时候同时执行。
package drcm

import (
	"fmt"
	"sync"

	"github.com/idcsource/insight00-lib/cpool"
	"github.com/idcsource/insight00-lib/hardstore"
	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/nst"
)

// 创建一个锆存储，config的实例见源代码的zrstorage.cfg
func NewZrStorage(config *cpool.Block, logs *ilogs.Logs) (z *ZrStorage, err error) {
	z = &ZrStorage{
		config: config,
		logs:   logs,
		lock:   new(sync.RWMutex),
	}
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
	hardstore_config, err := z.config.GetSection("local")
	if err != nil {
		return err
	}
	z.local_store, err = hardstore.NewHardStore(hardstore_config)
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
	z.listen, err = nst.NewTcpServer(z, port, z.logs)
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err)
		return
	}
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
