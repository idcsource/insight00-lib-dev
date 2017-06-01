// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package dspiders

// The url crawl queue
type UrlCrawlQueue struct {
	urlchan chan UrlBasic
	count   uint64
}

// Initialize the url crawl queue, the chan's length is const URL_CRAWL_QUEUE_CAP
func NewUrlCrawlQueue() (u *UrlCrawlQueue) {
	u = &UrlCrawlQueue{
		urlchan: make(chan UrlBasic, URL_CRAWL_QUEUE_CAP),
		count:   0,
	}
	return
}

// Add a url basic information to the url crawl queue
func (u *UrlCrawlQueue) Add(ub UrlBasic) {
	u.urlchan <- ub
	u.count++
}

// Get one url basic information from the url crawl queue
func (u *UrlCrawlQueue) Get() (ub UrlBasic) {
	ub = <-u.urlchan
	u.count--
	return
}

// Get the queue's length
func (u *UrlCrawlQueue) Count() (count uint64) {
	return u.count
}
