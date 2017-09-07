// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

const (
	TRANSACTION_SIGNAL_CHANNEL_LEN = 2   // 事务信号channel的大小
	TRANSACTION_CLEAN_TIME_OUT     = 600 // 事务的超时时间，超时这段时间后就会被清理，单位秒

	SPOT_CACHE_SIGNAL_CHANNEL_LEN = 2   // spot缓存信号channel的大小
	SPOT_CACHE_CLEAN_CYCLE        = 200 // spot缓存的清理周期，每几次有效的角色请求之后进行
	SPOT_CACHE_CLEAN_TIME_OUT     = 600 // 超时的spot缓存时间，超时这段时间后就会被清理，单位秒
)

const (
	// 运行状态
	TRULE_RUN_NO       = iota // 未指定
	TRULE_RUN_RUNNING         // 正在运行
	TRULE_RUN_PAUSEING        // 正在暂停
	TRULE_RUN_PAUSED          // 已经暂停
)

const (
	// 事务占用角色的锁模式
	TRAN_LOCK_MODE_NO = iota
	// 读
	TRAN_LOCK_MODE_READ
	// 写
	TRAN_LOCK_MODE_WRITE
)

const (
	// 事务超时监测时间，单位秒
	TRAN_TIME_OUT_CHECK = 30
	// 事务超时时间，单位为秒
	TRAN_TIME_OUT = 120
	// 最多事务数
	TRAN_MAX_COUNT = 1000
)

const (
	// 事务的提交请求
	TRANSACTION_ASK_NO = iota
	// 请求建立
	TRANSACTION_ASK_BEGIN
	// 请求执行
	TRANSACTION_ASK_COMMIT
	// 请求回滚
	TRANSACTION_ASK_ROLLBACK
	// 请求消灭
	TRANSACTION_ASK_DELETE
	// 请求清理
	TRANSACTION_ASK_CLEAN
)

const (
	// spot缓存的请求
	SPOT_CACHE_ASK_NO = iota
	// 请求获取
	SPOT_CACHE_ASK_GET
	// 请求写入
	SPOT_CACHE_ASK_WRITE
	// 请求完成
	SPOT_CACHE_ASK_STORE
	// 请求重置
	SPOT_CACHE_ASK_RESET
	// 请求删除
	SPOT_CACHE_ASK_DELETE
	// 请求释放
	SPOT_CACHE_ASK_RELEASE
	// 请求清理
	SPOT_CACHE_ASK_CLEAN
)

const (
	// spot请求的返回
	SPOT_CACHE_RETURN_HANDLE_NO = iota
	// 返回OK
	SPOT_CACHE_RETURN_HANDLE_OK
	// 返回错误
	SPOT_CACHE_RETURN_HANDLE_ERROR
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
	// spot被删除——没有
	TRAN_SPOT_BE_DELETE_NO = iota
	// spot被标记删除
	TRAN_SPOT_BE_DELETE_YES
	// spot被真正删除掉了
	TRAN_SPOT_BE_DELETE_COMMIT
)
