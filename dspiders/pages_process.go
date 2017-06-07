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
	"github.com/idcsource/Insight-0-0-lib/drule2/reladb"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// Handle the store for pages when the crawler get the page.
type PagesProcess struct {
	config        *cpool.Section     // The PagesStore config
	crawlQueue    *UrlCrawlQueue     // The url crawl queue
	indexQueue    *WordsIndexProcess // The words index queue
	drule         *operator.Operator // The DRule2 operator
	pagedb        *reladb.RelaDB     // The page data RelaDB
	mediadb       *reladb.RelaDB     // The media data RelaDB
	arounddb      *operator.Operator // The around link data's db
	aroundname    string             // The around link data's area name
	urlfilter     []string           // If the url in the url filter, it will not be store
	domains       []string           // Which domain can be store.
	entr_cycle_in int64              // entrance url for cyclical's intervals
	entr_url      []UrlBasic         // the entrance urls
	closed        bool               // If close the pages process, it will true
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
		return
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
	p.aroundname, err = config.GetConfig("aroundname")
	if err != nil {
		return
	}
	p.arounddb = drule
	// the page data db
	var pagedbname string
	pagedbname, err = config.GetConfig("pagedb")
	if err != nil {
		return
	}
	p.pagedb, err = reladb.NewRelaDBWithDRule(pagedbname, drule)
	if err != nil {
		return
	}
	// the media data db
	var mediadbname string
	mediadbname, err = config.GetConfig("mediadb")
	if err != nil {
		return
	}
	p.mediadb, err = reladb.NewRelaDBWithDRule(mediadbname, drule)
	if err != nil {
		return
	}

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
func (p *PagesProcess) AddPage(page *PageData) (err error) {
	/*
		check if the url is in urlfilter
		check if the url already exist.
		if exist {
			insert the new version.
		}else{
			create the url table.
			insert the new version.
		}
		send the page to words processor.
	*/
	if pubfunc.StringInSlice(p.urlfilter, page.Url) == true {
		return
	}
	var tableexit bool
	tableexit, err = p.pagedb.TableExist(page.Url)
	if err != nil {
		return
	}
	if tableexit == true {
		_, err = p.pagedb.InsertForAutoField(page.Url, page)
		if err != nil {
			return
		}
	} else {
		err = p.pagedb.NewTable(page.Url, &PageData{}, "Ver")
		if err != nil {
			return
		}
		_, err = p.pagedb.InsertForAutoField(page.Url, page)
		if err != nil {
			return
		}
	}
	// TODO send the page to words processor.
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
	var tableexit bool
	tableexit, err = p.mediadb.TableExist(media.Url)
	if err != nil {
		return
	}
	if tableexit == true {
		_, err = p.mediadb.InsertForAutoField(media.Url, media)
		if err != nil {
			return
		}
	} else {
		err = p.mediadb.NewTable(media.Url, &PageData{}, "Ver")
		if err != nil {
			return
		}
		_, err = p.mediadb.InsertForAutoField(media.Url, media)
		if err != nil {
			return
		}
	}
	return
}

// Add urls to crawl queue whitch clawler get.
func (p *PagesProcess) AddUrls(urls []UrlBasic) (err error) {
	for _, oneurl := range urls {
		// check the url if is in the domain
		if pubfunc.StringInSlice(p.domains, oneurl.Domain) == true {
			// check if the url is in the store.
			var exist bool
			exist, err = p.pagedb.TableExist(oneurl.Url)
			if err != nil {
				return
			}
			if exist == false {
				// not exisit
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
				var count uint64
				count, err = p.pagedb.Count(oneurl.Url)
				if err != nil {
					return
				}
				var UpTime time.Time
				var UpInterval int64
				var Hash string
				err = p.pagedb.SelectFields(oneurl.Url, count, "UpTime", &UpTime, "UpInterval", &UpInterval, "Hash", &Hash)
				if err != nil {
					return
				}
				if UpTime.Unix()+UpInterval < time.Now().Unix() {
					// if can update
					oneurl.Hash = Hash
					oneurl.Ver = count + 1
					err = p.crawlQueue.Add(oneurl)
					if err != nil {
						return
					}
				}
			}
		} else {
			// if the url is in the around link
			exist, errd := p.arounddb.ExistRole(p.aroundname, oneurl.Url)
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
			errd = p.arounddb.StoreRole(p.aroundname, thearound)
			if errd.IsError() != nil {
				err = errd.IsError()
				return
			}
			// TODO text index
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

// To index the text.
func (w *PagesProcess) DoIndex(page *PageData) (err error) {
	// key words index
	// content index
	return
}
