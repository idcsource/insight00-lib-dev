// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"net/url"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// Handle the store for pages when the crawler get the page.
type PagesProcess struct {
	config *cpool.Section // The PagesStore config

	crawlQueue *UrlCrawlQueue     // The url crawl queue
	indexQueue *WordsIndexProcess // The words index queue

	drule *operator.Operator // The DRule2 operator

	pagedb     *operator.Operator // The page data DRule2
	pagedbname string             // The page data DRule2 area name

	mediadb     *operator.Operator // The media data DRule2
	mediadbname string             // the media data DRule2 area name

	arounddb     *operator.Operator // The around link data's db
	arounddbname string             // The around link data's area name

	urlfilter []string // If the url in the url filter, it will not be store
	domains   []string // Which domain can be store.

	entr_cycle_in int64      // entrance url for cyclical's intervals
	entr_url      []UrlBasic // the entrance urls

	closed bool // If close the pages process, it will true
}

func NewPagesProcess(config *cpool.Section, crawlQueue *UrlCrawlQueue, indexQueue *WordsIndexProcess, drule *operator.Operator) (p *PagesProcess, err error) {
	p = &PagesProcess{
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
			Domain: theurl.Hostname(),
			Url:    one,
		}
	}
	// the around link db
	p.arounddbname, err = config.GetConfig("arounddbname")
	if err != nil {
		return
	}
	p.arounddb = drule
	// the page data db
	p.pagedbname, err = config.GetConfig("pagedbname")
	if err != nil {
		return
	}
	p.pagedb = drule
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
	p.mediadbname, err = config.GetConfig("mediadbname")
	if err != nil {
		return
	}
	p.mediadb = drule
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
		roleexist, errd := p.pagedb.ExistRole(p.pagedbname, page.Url)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		if roleexist == true {
			// if exist
			tran, errd := p.pagedb.Begin()
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			// update the UpInterval.
			var upInterval int64
			errd = tran.ReadData(p.pagedbname, page.Url, "UpInterval", &upInterval)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			if upInterval > UP_INTERVAL_MIN {
				upInterval = upInterval - ((upInterval - UP_INTERVAL_MIN) / 2)
				page.UpInterval = upInterval
			}
			errd = tran.StoreRole(p.pagedbname, page)
			if errd.IsError() != nil {
				tran.Rollback()
				err = errd.IsError()
				return
			}
			tran.Commit()
		} else {
			// if not exist
			page.UpInterval = UP_INTERVAL_DEFAULT
			errd := p.pagedb.StoreRole(p.pagedbname, page)
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
		roleexist, errd := p.pagedb.ExistRole(p.pagedbname, page.Url)
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		// if the role not exist
		if roleexist == false {
			return
		}
		tran, errd := p.pagedb.Begin()
		if errd.IsError() != nil {
			err = errd.IsError()
			return
		}
		var upInterval int64
		errd = tran.ReadData(p.pagedbname, page.Url, "UpInterval", &upInterval)
		if errd.IsError() != nil {
			tran.Rollback()
			err = errd.IsError()
			return
		}
		if upInterval < UP_INTERVAL_MAX {
			upInterval = ((UP_INTERVAL_MAX - upInterval) / 2) + upInterval
			errd = tran.WriteData(p.pagedbname, page.Url, "UpInterval", upInterval)
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
			exist, errd := p.pagedb.ExistRole(p.pagedbname, oneurl.Url)
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
				errd = p.pagedb.ReadData(p.pagedbname, oneurl.Url, "UpTime", &UpTime)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
				errd = p.pagedb.ReadData(p.pagedbname, oneurl.Url, "UpInterval", &UpInterval)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
				errd = p.pagedb.ReadData(p.pagedbname, oneurl.Url, "Hash", &Hash)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
				errd = p.pagedb.ReadData(p.pagedbname, oneurl.Url, "Ver", &Ver)
				if errd.IsError() != nil {
					err = errd.IsError()
					return
				}
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
			exist, errd := p.arounddb.ExistRole(p.arounddbname, oneurl.Url)
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
			errd = p.arounddb.StoreRole(p.arounddbname, thearound)
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
		p.AddUrls(p.entr_url)
		time.Sleep(time.Second * time.Duration(p.entr_cycle_in))
	}
}
