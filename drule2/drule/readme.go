// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// drule2的“分布式统治者”
//
// DRule需要提供一个*cpool.Block类型的配置信息，示例如下:
//	{drule}
//
//	[main]
//	# mode可以是master、slave
//	mode = master
//	# 自己的标识名称
//	selfname = Master0001
//	# slave配置的名字，用逗号分割，master需要
//	slave = s001,s002
//
//	[s001]
//	# 连接地址
//	address = 192.168.1.101:11111
//	# 连接数
//	conn_num = 5
//	# 是否加密，true或false
//	tls = true
//	# 用户名
//	username = username
// 	# 密码
//	password = password
//
//	[s002]
//	# 连接地址
//	address = 192.168.1.102:11111
//	# 连接数
//	conn_num = 5
//	# 是否加密，true或false
//	tls = false
//	# 用户名
//	username = username
// 	# 密码
//	password = password
//
// 分布式路由功能在运行时设置，不依靠配置文件
package drule
