// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"sync"
	"time"

	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/spots"
	"github.com/idcsource/insight00-lib/srule/hardstorage"
)

// 单个角色缓存
type spotCache struct {
	// 角色缓存处理锁，由spotCacheOp来操作
	op_lock *sync.RWMutex
	// 角色所在分区，roleCacheOp管理
	area string
	// 角色的id，roleCacheOp管理
	id string
	// 角色是否真正存在，roleCacheOp管理
	exist bool
	// 当前角色，roleCacheOp管理
	spot *spots.Spots
	// 事务中操作的锁，由Transaction操作
	spot_lock *sync.RWMutex
	// 是否为写模式，roleCacheOp管理
	forwrite bool
	// 被删除，TRAN_ROLE_BE_DELETE_*，roleCacheOp管理，由Transaction操作
	be_delete uint8
	// 被修改，roleCacheOp管理，由Transaction操作
	be_change bool

	// 占用的事务id
	tran_id string
	// 被事务的占用开始时间
	tran_time time.Time
	// 请求排队队列
	wait_line []*cacheAskSpot
	// 排队锁
	wait_line_lock *sync.RWMutex
}

// 事务的等待Spot排队
type cacheAskSpot struct {
	// 操作类型，SPOT_CACHE_ASK_*的部分项目
	optype uint8
	// 事务ID
	tran_id string
	// 是为了写吗
	forwrite bool
	// 返回句柄
	approved chan bool
	// 请求的时间
	ask_time time.Time
}

// 角色信号的返回
type spotCacheReturn struct {
	status uint8      // 状态，SPOT_CACHE_RETURN_HANDLE_*
	exist  bool       // 是否存在
	err    error      // 错误
	spot   *spotCache // 获得这个Spot
}

// Spot的处理信号
type spotCacheSig struct {
	ask      uint8                 // 请求什么 SPOT_CACHE_ASK_*
	area     string                // 角色的区域
	id       string                // 角色的id
	tranid   string                // 事务id
	forwrite bool                  // 是否为了写操作
	ask_time time.Time             // 请求的时间
	re       chan *spotCacheReturn // 角色信号的返回
}

// Spot缓存处理机
type spotCacheOp struct {
	local_store *hardstorage.HardStorage         // 本地存储
	signal      chan *spotCacheSig               // Spot的请求
	cache       map[string]map[string]*spotCache // 缓存池
	clean_count int                              // 清理计数器
	log         *ilogs.Logs                      // the Log
	closed      bool                             // 如果关闭就是true
	closesig    chan bool                        // 关闭的信号
}

// 事务的请求处理信号
type transactionSig struct {
	ask uint8                   // 请求TRANSACTION_ASK_*
	id  string                  // transaction的id
	re  chan *transactionReturn // 返回值的channel
}

// 事务请求处理的返回值
type transactionReturn struct {
	status uint8        // 状态，TRAN_RETURN_HANDLE_*
	err    error        // 错误
	tran   *Transaction // 事务
}

// 事务
type Transaction struct {
	// 事务id
	id string
	// 事务内角色缓存
	spot_cache map[string]map[string]*spotCache
	// 事务内角色缓存名称
	spot_cache_name map[string]map[string]bool
	// 事务内角色缓存锁
	spot_cache_lock *sync.RWMutex
	// 缓存处理机的信号
	spot_cache_sig chan *spotCacheSig
	// 事务处理机的信号
	tran_sig chan *transactionSig
	// 事务的活动日期
	tran_time time.Time
	// 被删除标记
	be_delete bool
}

// 事务的处理机
type transactionOp struct {
	signal             chan *transactionSig    // 事务的处理信号
	transaction        map[string]*Transaction // 事务池
	spotCache          chan *spotCacheSig      // 缓存处理机的信号
	max_transaction    int                     // 最大允许事务数
	count_transaction  int                     // 当前事务数
	tran_timeout       int64                   // 事务超时时间，单位秒
	tran_timeout_check int64                   // 事务超时监测时间，单位秒
	closed             uint8                   // 真正关闭是2,配置成关闭是1,正常运行是0
	closesig           chan bool               // 关闭的信号
}

// 事务统治者
type TRule struct {
	/* 下面是基础部分 */

	// 本地存储
	local_store *hardstorage.HardStorage
	// 日志
	log *ilogs.Logs

	/* 下面是事务相关部分 */

	spot_cache_op *spotCacheOp
	// 角色缓存的信号
	spot_cache_sig chan *spotCacheSig

	transcation_op *transactionOp
	// 事务的信号
	transaction_signal chan *transactionSig

	// 正在暂停信号
	pausing_signal chan bool
	// 已经暂停信号
	paused_signal chan bool
	// 工作状态，来自TRULE_RUN_*
	work_status uint8
	// 事务等待计数
	tran_wait *sync.WaitGroup
}
