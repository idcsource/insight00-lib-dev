// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package dspiders

const (
	URL_CRAWL_QUEUE_CAP uint = 1000 // The url crawl queue's capacity.
)

// The network transport operate code
type NetTransportOperate uint

const (
	NET_TRANSPORT_OPERATE_NO                  NetTransportOperate = iota // the code is null
	NET_TRANSPORT_OPERATE_URL_CRAWL_QUEUE_ADD                            // add a basic url information to url crawl queue
	NET_TRANSPORT_OPERATE_URL_CRAWL_QUEUE_GET                            // get a basic url information from url crawl queue
)
