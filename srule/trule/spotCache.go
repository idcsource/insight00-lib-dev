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
)

// 初始化一个空的角色缓存
func initSpotC(area, id string) (spotc *spotCache) {
	spotc = &spotCache{
		area:           area,
		id:             id,
		exist:          false,
		forwrite:       false,
		be_delete:      TRAN_SPOT_BE_DELETE_NO,
		be_change:      false,
		tran_time:      time.Now(),
		wait_line:      make([]*cacheAskSpot, 0),
		wait_line_lock: new(sync.RWMutex),
		op_lock:        new(sync.RWMutex),
		spot_lock:      new(sync.RWMutex),
	}
	//go spotc.listen()
	return
}

// 排队信号监听
//func (r *spotCache) listen() {
//	for {
//		wait_sig := <-s.wait_line_sig
//		switch wait_sig.optype {
//		case SPOT_CACHE_ASK_GET:
//			// 请求获取
//			s.askToGet(wait_sig)
//		case SPOT_CACHE_ASK_RELEASE:
//			// 请求释放
//			s.askToRelease(wait_sig)
//		}
//	}
//}

// 处理请求获取的信号
func (s *spotCache) askToGet(wait *cacheAskSpot) (approved bool) {
	s.wait_line_lock.Lock()
	defer s.wait_line_lock.Unlock()
	towait := false
	if wait.forwrite == true {
		if s.tran_id == "" || s.tran_id == wait.tran_id {
			s.tran_id = wait.tran_id
			s.tran_time = time.Now()
			s.forwrite = true
		} else {
			towait = true
		}
	} else {
		if s.tran_id != "" || s.forwrite == true {
			towait = true
		} else {
			s.tran_time = time.Now()
		}
	}
	if towait == false {
		// 如果可以给予就发送
		approved = true
	} else {
		// 否则就加入等待，并返回false
		s.wait_line = append(s.wait_line, wait)
		approved = false
	}
	return
}

// 处理请求释放的信号，等待队列空就返回true，否则返回false
func (s *spotCache) askToRelease() (approved bool) {
	s.wait_line_lock.Lock()
	defer s.wait_line_lock.Unlock()

	s.tran_id = ""
	s.forwrite = false

	waitlen := len(s.wait_line)
	if waitlen == 0 {
		// 队列空，就发true给发出释放信号的家伙
		approved = true
	} else {
		thenext := s.wait_line[0]
		// 队列没空，就发false给发出释放信号的家伙
		approved = false
		if thenext.forwrite == true {
			s.tran_id = thenext.tran_id
			s.forwrite = true
		} else {
			s.tran_id = ""
			s.forwrite = false
		}
		s.tran_time = time.Now()

		if waitlen == 1 {
			s.wait_line = make([]*cacheAskSpot, 0)
		} else {
			new_wait_line := s.wait_line[1:]
			s.wait_line = new_wait_line
		}
		thenext.approved <- true
	}
	return
}

// 设置角色
//func (s *spotCache) setSpot(spot *spots.SpotMiddleData) {
//	s.spot = spot
//}

// 加入某角色缓存的等待队列
func (s *spotCache) addWait(wait *cacheAskSpot) {
	s.wait_line = append(s.wait_line, wait)
}
