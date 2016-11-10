// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this souDMODE_MASTERrce code is governed by GNU LGPL v3 license

package drcm

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 创建一个锆存储，config的实例见源代码的zrstorage.cfg
func NewZrStorage (config *cpool.Block, logs *ilogs.Logs) (z *ZrStorage, err error) {
	z = &ZrStorage{
		RolePlus : rolesplus.RolePlus{
			Role: roles.Role{},
		},
		config : config,
		logs : logs,
	};
	z.New(random.Unid(1,"ZrStorage"));
	// 处理运行的模式
	mode, err := config.GetConfig("main.mode");
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err);
		return;
	}
	switch mode {
		case "own":
			z.dmode = DMODE_OWN;
			// 使用own模式来启动锆存储
			err = z.startUseOwn();
		case "master":
			z.dmode = DMODE_MASTER;
			// 使用master模式来启动锆存储
			err = z.startUseMaster();
		case "slave":
			z.dmode = DMODE_SLAVE;
			// 使用slave模式来启动锆存储
			err = z.startUseSlave();
		default :
			err = fmt.Errorf("drcm:NewZrStorage: the mode config must own, master or slave.");
			return;
	}
	return;
}

// 创建缓存
func (z *ZrStorage) buildCache () (err error) {
	z.cacheMax, err = z.config.TranInt64("local.cache_num");
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err);
		return;
	}
	z.rolesCache = make(map[string]oneRoleCache);
	z.deleteCache = make([]string,0);
	z.cacheIsFull = make(chan bool);
	z.rolesCount = 0;
	z.checkCacheNumOn = false;
	return;
}

// 使用own模式来启动锆存储
func (z *ZrStorage) startUseOwn () (err error) {
	// 创建本地存储
	z.local_store, err = hardstore.NewHardStore(z.config);
	if err != nil {
		return;
	}
	// 创建缓存
	err = z.buildCache();
	return;
}

// 使用slave模式来启动锆存储，也就是启动一个tcp的监听（但先要启用本地存储）
func (z *ZrStorage) startUseSlave () (err error) {
	err = z.startUseOwn();
	if err != nil {
		return;
	}
	port, err := z.config.GetConfig("main.port");
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err);
		return;
	}
	z.listen = nst.NewTcpServer(z, port, z.logs);
	z.code, err = z.config.GetConfig("main.code");
	if err != nil {
		err = fmt.Errorf("drcm:NewZrStorage: %v", err);
		return;
	}
	return;
}

// 使用master模式来启动锆存储，也就是连接所有的slave，当然也要启动监听和本地存储
func (z *ZrStorage) startUseMaster () (err error) {
	err = z.startUseSlave();
	if err != nil {
		return;
	}
	return;
}
