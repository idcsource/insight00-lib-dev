// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package dspiders

import (
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/reladb"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
)

// The network transport handle.
//
// It operate url crawl queue's request, the page's storage, the words's index and others.
type NetTransportHandle struct {
	urlCrawlQueue  *UrlCrawlQueue // The url crawl queue
	trule          *trule.TRule   // The TRule
	reladb         *reladb.RelaDB // The RelaDB
	identityConfig *cpool.Section // The indentity code's config, the *cpool.Section is name = code
	// TODO : Others
}
