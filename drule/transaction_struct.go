// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"sync"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/hardstore"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// 事务统治者
type TRule struct {
	/* 下面是基础部分 */

	// 配置信息
	config *cpool.Block
	// 本地存储
	local_store *hardstore.HardStore
	// 日志
	log *ilogs.Logs

	/* 下面是事务相关部分 */

	// 事务服务
	tran_service *tranService
	// 事务列表，string为事务的unid
	transaction map[string]*Transaction
	// 最大允许事务数
	max_transaction int
	// 当前事务数
	count_transaction int
	// 事务超时时间，单位秒
	tran_timeout int64
	// 事务列表锁
	tran_lock *sync.RWMutex
	// 事务的信号
	tran_commit_signal chan *tranCommitSignal
}

// 事务
type Transaction struct {
	// 事务id
	unid string
	// 事务内缓存
	tran_cache map[string]*roleCache
	// 事务内缓存锁
	lock *sync.RWMutex
	// 事务服务
	tran_service *tranService
	// 事务的开始时间
	tran_time time.Time
	// 事务的信号，其实是发送给ZrStorage的
	tran_commit_signal chan *tranCommitSignal
	// 被删除标记
	be_delete bool
}

// 事务服务
type tranService struct {
	// 本地存储
	local_store *hardstore.HardStore
	// 全事务角色缓存
	role_cache map[string]*roleCache
	// 全事务角色缓存锁
	lock *sync.RWMutex
}

// 角色缓存
type roleCache struct {
	// 当前角色
	role *roles.RoleMiddleData
	// 角色的本尊
	role_store roles.RoleMiddleData
	// 被删除
	be_delete uint8
	// 占用的事务id
	tran_id string
	// 被事务的占用开始时间
	tran_time time.Time
	// 请求排队队列
	wait_line []*tranAskGetRole
	// 排队锁
	lock *sync.RWMutex
}

// 角色的存储类型，也就是角色编码后的内容（使用hardstore中的编码方案）
type roleStore struct {
	body []byte
	rela []byte
	vers []byte
}

// 事务的请求角色排队
type tranAskGetRole struct {
	// 事务ID
	tran_id string
	// 返回句柄
	approved chan bool
	// 请求的时间
	ask_time time.Time
}

// 事务执行的处理信号
type tranCommitSignal struct {
	// 事务ID
	tran_id string
	// 请求内容，TRAN_COMMIT_ASK_*
	ask uint8
	// 返回句柄
	return_handle chan tranReturnHandle
}

// 事务的返回句柄
type tranReturnHandle struct {
	// 状态，TRAN_RETURN_HANDLE_*
	Status uint8
	// 错误
	Error error
}
