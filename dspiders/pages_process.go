// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"
	"net/url"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// Handle the store for pages when the crawler get the page.
type PagesProcess struct {
	sitename string         // The site name.
	config   *cpool.Section // The PagesStore config

	crawlQueue *UrlCrawlQueue     // The url crawl queue
	indexQueue *WordsIndexProcess // The words index queue

	drule *operator.Operator // The DRule2 operator

	pagedb   *dbinfo // The page data information
	mediadb  *dbinfo // The media data information
	arounddb *dbinfo // The around link data information

	urlfilter []string // If the url in the url filter, it will not be store
	domains   []string // Which domain can be store.

	entr_cycle_in int64      // entrance url for cyclical's intervals
	entr_url      []UrlBasic // the entrance urls

	closed bool // If close the pages process, it will true
}

func NewPagesProcess(sitename string, config *cpool.Section, crawlQueue *UrlCrawlQueue, indexQueue *WordsIndexProcess, drule *operator.Operator) (p *PagesProcess, err error) {
	p = &PagesProcess{
		sitename:   sitename,
		config:     config,
		crawlQueue: crawlQueue,
		indexQueue: indexQueue,
		drule:      drule,
		closed:     true,
	}
	// the url filter
	p.urlfilter, err = config.GetEnum("urlfilter")
	if err != nil {
		p.urlfilter = make([]string, 0)
		err = nil
	}
	// the domains url
	p.domains, err = config.GetEnum("domains")
	if err != nil {
		return
	}
	// the entrance url for cyclical's intervals
	p.entr_cycle_in, err = config.TranInt64("entr_cycle_in")
	if err != nil {
		return
	}
	// the entrance urls
	var entr_url []string
	entr_url, err = config.GetEnum("entr_url")
	if err != nil {
		return
	}
	p.entr_url = make([]UrlBasic, len(entr_url))
	for i, one := range entr_url {
		var theurl *url.URL
		theurl, err = url.Parse(one)
		if err != nil {
			return
		}
		p.entr_url[i] = UrlBasic{
			SiteName: p.sitename,
			Domain:   theurl.Hostname(),
			Url:      one,
		}
	}
	// the around link db
	p.arounddb = &dbinfo{
		drule: drule,
	}
	p.arounddb.area, err = config.GetConfig("arounddbname")
	if err != nil {
		return
	}
	// the page data db
	p.pagedb = &dbinfo{
		drule: drule,
	}
	p.pagedb.area, err = config.GetConfig("pagedbname")
	if err != nil {
		return
	}
	/*
		var pagedbname string
		pagedbname, err = config.GetConfig("pagedb")
		if err != nil {
			return
		}
		p.pagedb, err = reladb.NewRelaDBWithDRule(pagedbname, drule)
		if err != nil {
			return
		}
	*/
	// the media data db
	p.mediadb = &dbinfo{
		drule: drule,
	}
	p.mediadb.area, err = config.GetConfig("mediadbname")
	if err != nil {
		return
	}
	/*
		var mediadbname string
		mediadbname, err = config.GetConfig("mediadb")
		if err != nil {
			return
		}
		p.mediadb, err = reladb.NewRelaDBWithDRule(mediadbname, drule)
		if err != nil {
			return
		}
	*/
	return
}

func (p *PagesProcess) Start() {
	p.closed = false
	go p.addEntrUrls()
}

func (p *PagesProcess) Close() {
	p.closed = true
}

// Add a page data to store whitch crawler get.
func (p *PagesProcess) AddPage(page *PageData, status NetDataStatus) (err error) {

	// check if the data change(the hash change)
	if status == NET_DATA_STATUS_PAGE_UPDATE {
		// if update
		roleexist, errd := p.pagedb.drule.ExistRole(p.pagedb.area, page.Url)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		if roleexist == true {
			// if exist
			tran, errd := p.pagedb.drule.Begin()
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			// update the UpInterval.
			var upInterval int64
			errd = tran.ReadData(p.pagedb.area, page.Url, "UpInterval", &upInterval)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			if upInterval > UP_INTERVAL_MIN {
				upInterval = upInterval - ((upInterval - UP_INTERVAL_MIN) / 2)
				page.UpInterval = upInterval
			}
			errd = tran.StoreRole(p.pagedb.area, page)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			tran.Commit()
		} else {
			// if not exist
			page.UpInterval = UP_INTERVAL_DEFAULT
			errd := p.pagedb.drule.StoreRole(p.pagedb.area, page)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
		}
		// if the url in urlfilter, just do not be index.
		if pubfunc.StringInSlice(p.urlfilter, page.Url) != true {
			// if the url not in urlfilter, send the page to words processor.
			index_req := &WordsIndexRequest{
				Url:      page.Url,
				Domain:   page.Domain,
				Type:     WORDS_INDEX_TYPE_PAGE,
				PageData: page,
			}
			err = p.indexQueue.Add(index_req)
			return
		}
	} else if status == NET_DATA_STATUS_PAGE_NOT_UPDATE {
		// if not update, just update the UpInterval.
		roleexist, errd := p.pagedb.drule.ExistRole(p.pagedb.area, page.Url)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		// if the role not exist
		if roleexist == false {
			return
		}
		tran, errd := p.pagedb.drule.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		var upInterval int64
		errd = tran.ReadData(p.pagedb.area, page.Url, "UpInterval", &upInterval)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if upInterval < UP_INTERVAL_MAX {
			upInterval = ((UP_INTERVAL_MAX - upInterval) / 2) + upInterval
			errd = tran.WriteData(p.pagedb.area, page.Url, "UpInterval", upInterval)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
		}
		tran.Commit()
	}

	return
}

// Add a media data to store whitch crawler get.
func (p *PagesProcess) AddMedia(media *MediaData) (err error) {
	// TODO or not TODO?
	return
}

// Add urls to crawl queue whitch clawler get.
func (p *PagesProcess) AddUrls(urls []UrlBasic) (err error) {
	for _, oneurl := range urls {
		// check the url if is in the domain
		if pubfunc.StringInSlice(p.domains, oneurl.Domain) == true {
			// check if the url is in the store.
			exist, errd := p.pagedb.drule.ExistRole(p.pagedb.area, oneurl.Url)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			if exist == false {
				// not exist
				if pubfunc.StringInSlice(p.urlfilter, oneurl.Url) == true {
					oneurl.Filter = true
				} else {
					oneurl.Filter = false
				}
				oneurl.Hash = ""
				oneurl.Ver = 0
				err = p.crawlQueue.Add(oneurl)
				if err != nil {
					return
				}
			} else {
				// get the store's last version, look the update time.
				var UpTime time.Time
				var UpInterval int64
				var Hash string
				var Ver uint64
				tran, errd := p.pagedb.drule.Begin()
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
				errd = tran.ReadData(p.pagedb.area, oneurl.Url, "UpTime", &UpTime)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
				errd = tran.ReadData(p.pagedb.area, oneurl.Url, "UpInterval", &UpInterval)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
				errd = tran.ReadData(p.pagedb.area, oneurl.Url, "Hash", &Hash)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
				errd = tran.ReadData(p.pagedb.area, oneurl.Url, "Ver", &Ver)
				if errd.IsError() != nil {
					tran.Rollback()
					err = errd.IsError()
					return
				}
				tran.Commit()
				if UpTime.Unix()+UpInterval < time.Now().Unix() {
					// if can update
					oneurl.Hash = Hash
					oneurl.Ver = Ver + 1
					err = p.crawlQueue.Add(oneurl)
					if err != nil {
						return
					}
				}
			}
		} else {
			// if the url is in the around link
			exist, errd := p.arounddb.drule.ExistRole(p.arounddb.area, oneurl.Url)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			if exist == true {
				continue
			}
			thearound := &AroundLink{
				Url:  oneurl.Url,
				Text: oneurl.Text,
			}
			thearound.New(oneurl.Url)
			errd = p.arounddb.drule.StoreRole(p.arounddb.area, thearound)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			// text index
			index_req := &WordsIndexRequest{
				Url:        oneurl.Url,
				Domain:     oneurl.Domain,
				Type:       WORDS_INDEX_TYPE_AROUND,
				AroundLink: thearound,
			}
			err = p.indexQueue.Add(index_req)
			if err != nil {
				return
			}
		}
	}
	return
}

// Add all entrance url for cyclical. use: go p.addEntrUrls()
func (p *PagesProcess) addEntrUrls() {
	for {
		if p.closed == true {
			return
		}
		err := p.AddUrls(p.entr_url)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second * time.Duration(p.entr_cycle_in))
	}
}
