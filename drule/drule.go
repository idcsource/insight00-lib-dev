// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"reflect"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 创建一个分布式统治者
//
// 注意：如果想用drcm.ZrStorage那样的Own模式，则直接去用TRule。
func NewDRule(config *cpool.Block, logs *ilogs.Logs) (d *DRule, err error) {
	d = &DRule{
		config: config,
		logs:   logs,
		closed: false,
	}
	// 查找运行模式
	var mode string
	mode, err = config.GetConfig("main.mode")
	if err != nil {
		err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
		return
	}
	switch mode {
	case "master":
		d.dmode = DMODE_MASTER
		err = d.startForMaster()
	case "slave":
		d.dmode = DMODE_SLAVE
		err = d.startForSlave()
	default:
		err = fmt.Errorf("drule[DRule]NewDRule: The mode config must master or slave.")
		return
	}
	if err != nil {
		err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
	}
	return
}

// 关闭DRule
func (d *DRule) Close() (err error) {
	d.closed = true
	/* 是不是要检查一下事务是否都结束呢 */
	for i := 0; i < 5; i++ {
		if d.trule.TransactionCount() != 0 {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	d.listen.Close()
	d.closeSlavePool()
	return
}

// 本地的启动
func (d *DRule) startForOwn() (err error) {
	// 创建本地存储
	hardstore_config, err := d.config.GetSection("local")
	if err != nil {
		return err
	}
	local_store, err := hardstore.NewHardStore(hardstore_config)
	if err != nil {
		return err
	}
	// 创建事务统治者
	d.trule, err = NewTRule(local_store)
	// 查看自己的名字，没有就直接忽略
	d.selfname, _ = d.config.GetConfig("main.selfname")
	return
}

// 使用slave模式来启动，也就是启动一个tcp的监听（但先要启用本地存储）
func (d *DRule) startForSlave() (err error) {
	err = d.startForOwn()
	if err != nil {
		return
	}
	port, err := d.config.GetConfig("main.port")
	if err != nil {
		return err
	}
	d.listen, err = nst.NewTcpServer(d, port, d.logs)
	if err != nil {
		return err
	}
	d.code, err = d.config.GetConfig("main.code")
	if err != nil {
		return err
	}
	return
}

// 使用master模式来启动
func (d *DRule) startForMaster() (err error) {
	err = d.startForSlave()
	if err != nil {
		return
	}
	d.slaves = make(map[string][]*slaveIn)
	d.slavepool = make(map[string]*nst.TcpClient)
	d.slavecpool = make(map[string]*slaveIn)
	// 获取slave的配置名
	slaves, err := d.config.GetEnum("main.slave")
	if err != nil {
		return err
	}
	// 遍历所有的slave配置名
	for _, one := range slaves {
		// 获取每个slave的配置
		onecfg, err := d.config.GetSection(one)
		if err != nil {
			d.closeSlavePool()
			return err
		}
		// 获取这个slave可管理的角色首字母
		control_whos, err := onecfg.GetEnum("control")
		if err != nil {
			d.closeSlavePool()
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
			d.closeSlavePool()
			return err
		}
		// 获取连接地址
		addr, err := onecfg.GetConfig("address")
		if err != nil {
			d.closeSlavePool()
			return err
		}
		// 创建连接和连接池，放到池子里主要是为了到时候出错了关闭方便
		//z.slavepool = make(map[string]*nst.TcpClient)
		sconn, err := nst.NewTcpClient(addr, conn_num, d.logs)
		if err != nil {
			d.closeSlavePool()
			return err
		}
		d.slavepool[one] = sconn
		d.slavecpool[one] = &slaveIn{
			name:    one,
			code:    code,
			tcpconn: sconn,
		}
		// 遍历可管理角色首字母创建连接序列
		for _, onewho := range control_whos {
			// 序列里没有这个字母就建立一个
			if _, have := d.slaves[onewho]; have == false {
				d.slaves[onewho] = make([]*slaveIn, 0)
			}
			// 将这个字母的序列中加入这个slave的名字
			d.slaves[onewho] = append(d.slaves[onewho], d.slavecpool[one])
		}
	}
	return
}

// 关闭整个slavepool
func (d *DRule) closeSlavePool() {
	for _, conn := range d.slavepool {
		conn.Close()
	}
}

// 查看连接是哪个，id为角色的id，connmode来自CONN_IS_*
func (d *DRule) findConn(id string) (connmode uint8, conn []*slaveIn) {
	id = hardstore.GetRoleStoreName(id)
	// 如果模式为own，则直接返回本地
	if d.dmode == DMODE_OWN {
		connmode = CONN_IS_LOCAL
		return
	}

	// 找到第一个首字母。
	theChar := string(id[0])
	// slave池中有没有
	conn, find := d.slaves[theChar]
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
func (d *DRule) statusValueType(value interface{}) (types uint8) {
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

// 处理错误日志
func (d *DRule) logerr(err error) {
	if err == nil {
		return
	}
	err = fmt.Errorf("drule[DRule]: %v", err)
	if d.logs != nil {
		d.logs.ErrLog(err)
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (d *DRule) logrun(err error) {
	if err == nil {
		return
	}
	err = fmt.Errorf("drule[DRule]: %v", err)
	if d.logs != nil {
		d.logs.RunLog(err)
	} else {
		fmt.Println(err)
	}
}
