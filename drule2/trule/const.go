// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package trule

const (
	// 事务占用角色的锁模式
	TRAN_LOCK_MODE_NO = iota
	// 读
	TRAN_LOCK_MODE_READ
	// 写
	TRAN_LOCK_MODE_WRITE
)

const (
	// 事务超时时间，单位为秒，监测的周期为30秒一次
	TRAN_TIME_OUT = 60
	// 最多事务数
	TRAN_MAX_COUNT = 1000
)

const (
	// 事务的提交请求
	TRAN_COMMIT_ASK_NO = iota
	// 请求执行
	TRAN_COMMIT_ASK_COMMIT
	// 请求回滚
	TRAN_COMMIT_ASK_ROLLBACK
)

const (
	// 事务提交或回滚后得到的从TRule的返回状态
	TRAN_RETURN_HANDLE_NO = iota
	// 返回OK
	TRAN_RETURN_HANDLE_OK
	// 返回错误
	TRAN_RETURN_HANDLE_ERROR
)

const (
	// 角色被删除——没有
	TRAN_ROLE_BE_DELETE_NO = iota
	// 角色被标记删除
	TRAN_ROLE_BE_DELETE_YES
	// 角色被真正删除掉了
	TRAN_ROLE_BE_DELETE_COMMIT
)
