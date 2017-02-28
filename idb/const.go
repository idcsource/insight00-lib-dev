// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 数据库连接处理，根据配置文件进行。
//
// 提供的配置信息为*cpool.Section类型。
// 需要配置信息提供的配置项为：
// 		type		# 数据库类型，目前只支持使用postgresql
// 		server		# 数据库服务器的地址
// 		port		# 数据库服务器的端口号
// 		user		# 数据库访问用户名
// 		passwd		# 数据库访问用户的密码
// 		dbname		# 数据库名
// 目前只支持PostgreSQL连接。
package idb

const (
	//数据库的类型
	DATABASE_TYPE_POSTGRESQL = iota
	DATABASE_TYPE_MYSQL
	DATABASE_TYPE_SQLITE
)
