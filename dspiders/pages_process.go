// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/reladb"
)

// Handle the store for pages when the crawler get the page.
type PagesProcess struct {
	drule      *operator.Operator // The DRule's operator
	config     *cpool.Block       // The PagesStore config
	crawlQueue *UrlCrawlQueue     // The url crawl queue
	reladb     *reladb.RelaDB     // The RelaDB
}

// Add a page data to store whitch crawler get.
func (p *PagesProcess) AddPage(page *PageData) (err error) {
	/*
		check if the url already exist.
		if exist {
			insert the new version.
		}else{
			create the url table.
			insert the new version.
		}
		send the page to words processor.
	*/
	return
}

// Add a media data to store whitch crawler get.
func (p *PagesProcess) AddMedia(media *MediaData) (err error) {
	/*
		check if the url already exist.
		if exist {
			insert the new version.
		}else{
			create the url table.
			insert the new version.
		}
	*/
	return
}

// Add urls to crawl queue whitch clawler get.
func (p *PagesProcess) AddUrls(urls []UrlBasic) (err error) {
	/*
		for {
			check the url if is in the domain
			if not {
				check if the url is in the around link
				if not {
					store this to around link
				}
			}else{
				check if the url is in the store.
				if yes {
					get the store's last version, look the update time.
					if can update {
						get the last hash, add the url to url crawl queue.
					}
				}else{
					add the url to url crwal queue.
				}
			}
		}
	*/
	return
}

// Add all entrance url for cyclical.
func (p *PagesProcess) addEntrUrls() {

}

// To index the text.
func (w *PagesProcess) DoIndex(page *PageData) (err error) {
	// key words index
	// content index
	return
}
