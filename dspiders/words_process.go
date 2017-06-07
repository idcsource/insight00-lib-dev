// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import "fmt"

// The words index process
type WordsIndexProcess struct {
	queue  chan WordsIndexRequest // the queue
	count  uint                   // the queue's count
	closed bool                   // if closed it will true
}

// return the index wait queue
func (w *WordsIndexProcess) ReturnQueue() chan WordsIndexRequest {
	return w.queue
}

// Add a url basic information to the url crawl queue
func (w *WordsIndexProcess) Add(req WordsIndexRequest) (err error) {
	if w.count == URL_CRAWL_QUEUE_CAP {
		err = fmt.Errorf("The queue is full.")
		return
	}
	w.queue <- req
	w.count++
	return
}
