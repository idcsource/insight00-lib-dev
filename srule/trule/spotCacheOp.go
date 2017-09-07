// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package trule

import (
	"fmt"
	"time"

	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/srule/hardstorage"
)

// 初始化spotCacheOp
func initSpotCacheOp(local_store *hardstorage.HardStorage, log *ilogs.Logs) (sco *spotCacheOp) {
	sco = &spotCacheOp{
		local_store: local_store,
		signal:      make(chan *spotCacheSig, SPOT_CACHE_SIGNAL_CHANNEL_LEN),
		cache:       make(map[string]map[string]*spotCache),
		clean_count: 0,
		log:         log,
		closed:      true,
		closesig:    make(chan bool),
	}
	return
}

func (sco *spotCacheOp) ReturnSig() (sig chan *spotCacheSig) {
	return sco.signal
}

// Start
func (sco *spotCacheOp) Start() {
	sco.closed = false
	go sco.listen()
}

// Stop
func (sco *spotCacheOp) Stop() {
	sco.closesig <- true
	sco.consCleanSig()
	sco.closed = true
}

// listen the spot cache operate signal
func (sco *spotCacheOp) listen() {
	for {
		if sco.closed == true {
			return
		}
		select {
		case signal := <-sco.signal:
			sco.doSignal(signal)
		case closesig := <-sco.closesig:
			if closesig == true {
				return
			}
		}
	}
}

// 构建清理的信号
func (sco *spotCacheOp) consCleanSig() {
	signal := &spotCacheSig{
		ask: SPOT_CACHE_ASK_CLEAN,
		re:  make(chan *spotCacheReturn),
	}
	sco.signal <- signal
	re := <-signal.re
	if re.status == SPOT_CACHE_RETURN_HANDLE_ERROR {
		sco.log.ErrLog(re.err)
	}
	sco.clean_count = 0
}

// do signal
func (sco *spotCacheOp) doSignal(signal *spotCacheSig) {
	switch signal.ask {
	case SPOT_CACHE_ASK_GET:
		// ask get a spot
		sco.askGetSpot(signal, false)
	case SPOT_CACHE_ASK_WRITE:
		// ask write a spot，两者的区别只在对角色不存在的错误处理上
		sco.askGetSpot(signal, true)
	case SPOT_CACHE_ASK_STORE:
		// ask to store the spot's change, and to next wait, for transaction commit
		sco.askStoreSpot(signal)
	case SPOT_CACHE_ASK_RESET:
		// ask to reset the spot, and to next wait, for transaction rollback
		sco.askResetSpot(signal)
	case SPOT_CACHE_ASK_CLEAN:
		sco.askCleanSpots(signal)
	default:
		sco.log.ErrLog("Spot Cache Signal's ask does not exist.")
	}
}

// ask to reset the spot, and to next wait, for transaction rollback
func (sco *spotCacheOp) askResetSpot(signal *spotCacheSig) {
	spotc := sco.cache[signal.area][signal.id]
	go sco.resetTheSpot(signal, spotc)
}

// gro to reset spot
func (sco *spotCacheOp) resetTheSpot(signal *spotCacheSig, spotc *spotCache) {
	// 加锁
	spotc.op_lock.Lock()
	// 延时解锁
	defer spotc.op_lock.Unlock()
	// 修正到初始状态
	if spotc.be_change == true {
		if spotc.exist == false {
			spotc.spot = nil
		} else {
			var err error
			spotc.spot, err = sco.local_store.SpotRead(spotc.area, spotc.id)
			if err != nil {
				sco.log.ErrLog(err)
			}
		}
		spotc.be_change = false
	}
	spotc.be_delete = TRAN_SPOT_BE_DELETE_NO
	// 释放当前
	spotc.askToRelease()
}

// ask to store the spot's change, and to next wait
func (sco *spotCacheOp) askStoreSpot(signal *spotCacheSig) {
	spotc := sco.cache[signal.area][signal.id]
	go sco.storeTheSpot(signal, spotc)
}

// go store the spot's change
func (sco *spotCacheOp) storeTheSpot(signal *spotCacheSig, spotc *spotCache) {
	// 加锁
	spotc.op_lock.Lock()
	// 延时解锁
	defer spotc.op_lock.Unlock()
	// 保存
	if spotc.be_change == true && spotc.be_delete == TRAN_SPOT_BE_DELETE_NO {
		err := sco.local_store.SpotStore(spotc.area, spotc.spot)
		if err != nil {
			sco.log.ErrLog(err)
		}
	}
	// 或真正删除
	if spotc.be_delete == TRAN_SPOT_BE_DELETE_YES || spotc.be_delete == TRAN_SPOT_BE_DELETE_COMMIT {
		err := sco.local_store.SpotDelete(spotc.area, spotc.id)
		if err != nil {
			sco.log.ErrLog(err)
		}
		spotc.be_delete = TRAN_SPOT_BE_DELETE_COMMIT
	}
	// 释放
	spotc.askToRelease()
}

// ask get a spot
func (sco *spotCacheOp) askGetSpot(signal *spotCacheSig, write bool) {
	// 在缓存中找到或分配位置
	_, havearea := sco.cache[signal.area]
	if havearea == false {
		sco.cache[signal.area] = make(map[string]*spotCache)
	}
	_, spothave := sco.cache[signal.area][signal.id]
	if spothave == false {
		sco.cache[signal.area][signal.id] = initSpotC(signal.area, signal.id)
	}
	// 交给协程去处理，但要先加锁，getSpotFromStorage中要释放这个锁
	go sco.getSpotFromStorage(signal, sco.cache[signal.area][signal.id], write)
}

// get spot from hardstorage
func (sco *spotCacheOp) getSpotFromStorage(signal *spotCacheSig, spotc *spotCache, write bool) {
	spotc.op_lock.Lock()

	re := &spotCacheReturn{}
	if spotc.exist == false && spotc.spot == nil {
		// 如果确定是没有
		exist := sco.local_store.SpotExist(signal.area, signal.id)
		if exist == false {
			re.exist = false
			if write == false {
				re.err = fmt.Errorf("The Spot not exist.")
				re.status = SPOT_CACHE_RETURN_HANDLE_ERROR
			} else {
				re.status = SPOT_CACHE_RETURN_HANDLE_OK
				sco.clean_count++
			}
		} else {
			thespot, err := sco.local_store.SpotRead(signal.area, signal.id)
			if err != nil {
				re.exist = false
				if write == false {
					re.err = err
					re.status = SPOT_CACHE_RETURN_HANDLE_ERROR
				} else {
					re.status = SPOT_CACHE_RETURN_HANDLE_OK
					sco.clean_count++
				}
			} else {
				re.exist = true
				re.status = SPOT_CACHE_RETURN_HANDLE_OK
				spotc.spot = thespot
				spotc.exist = true
				re.spot = spotc
				sco.clean_count++
			}
		}
	} else {
		re.exist = true
		re.status = SPOT_CACHE_RETURN_HANDLE_OK
		re.spot = spotc
	}
	// 解锁这个缓存
	spotc.op_lock.Unlock()
	// 尝试排队
	askspot := &cacheAskSpot{
		tran_id:  signal.tranid,
		forwrite: signal.forwrite,
		approved: make(chan bool),
		ask_time: signal.ask_time,
	}
	approved := spotc.askToGet(askspot)
	if approved == false {
		// 进入排队的话就去监听等待
		<-askspot.approved
	}

	// 发送这个re
	signal.re <- re

	// 查看是否清理
	if sco.clean_count >= SPOT_CACHE_CLEAN_CYCLE {
		go sco.consCleanSig()
	}
}

// ask clean spots
func (sco *spotCacheOp) askCleanSpots(signal *spotCacheSig) {
	// 这里不能有协程了，也就是说在这个执行完所有的请求角色的都要等了
	tmpa := make(map[string][]string)
	for areaname, _ := range sco.cache {
		tmpa[areaname] = make([]string, 0)
		for spotname, spotc := range sco.cache[areaname] {
			if spotc.tran_id == "" && spotc.tran_time.Unix()+SPOT_CACHE_CLEAN_TIME_OUT < time.Now().Unix() {
				tmpa[areaname] = append(tmpa[areaname], spotname)
			}
		}
	}
	for areaname, _ := range tmpa {
		for _, spotname := range tmpa[areaname] {
			delete(sco.cache[areaname], spotname)
		}
	}
}
