// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
)

// The words index process
type WordsIndexProcess struct {
	queue  chan *WordsIndexRequest // the queue
	count  uint                    // the queue's count
	closed bool                    // if closed it will true

	drule *operator.Operator // The DRule2 remote operator

	pagedb     *operator.Operator // The page data DRule2
	pagedbname string             // The page data DRule2 area name

	arounddb   *operator.Operator // The around link data index db
	aroundname string             // The around link data index area name

	keyworddb   *operator.Operator // The key word index db
	keywordname string             // The key word index area name
}

// return the index wait queue
func (w *WordsIndexProcess) ReturnQueue() chan *WordsIndexRequest {
	return w.queue
}

// Add a url basic information to the url crawl queue
func (w *WordsIndexProcess) Add(req *WordsIndexRequest) (err error) {
	if w.count == URL_CRAWL_QUEUE_CAP {
		err = fmt.Errorf("The queue is full.")
		return
	}
	w.queue <- req
	w.count++
	return
}

// go to index
func (w *WordsIndexProcess) goindex() {
	for {
		if w.closed == true {
			return
		}

		req := <-w.queue
		switch req.Type {
		case WORDS_INDEX_TYPE_PAGE:
			// if is the page
			w.indexPage(req)
		case WORDS_INDEX_TYPE_AROUND:
			// if is the around link
			w.indexAroundLink(req)
		default:
			continue
		}
	}
}

// the page index
func (w *WordsIndexProcess) indexPage(req *WordsIndexRequest) {

}

// the around link index
func (w *WordsIndexProcess) indexAroundLink(req *WordsIndexRequest) {

}
