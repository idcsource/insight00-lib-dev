// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/idcsource/Insight-0-0-lib/iendecode"
	"github.com/idcsource/Insight-0-0-lib/nst"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

type CrawlMachine struct {
	transport     *nst.TcpClient // the net transport
	identity_name string         // the identity name
	identity_code string         // the identity code
	closed        bool           // if closed the bool is true
}

func NewCrawlMachine(tcp *nst.TcpClient, name, code string) (c *CrawlMachine) {
	c = &CrawlMachine{
		transport:     tcp,
		identity_name: name,
		identity_code: code,
		closed:        true,
	}
	return
}

func (c *CrawlMachine) Start() {
	c.closed = false
	go c.gocrawl()
}

func (c *CrawlMachine) Close() {
	c.closed = true
}

func (c *CrawlMachine) gocrawl() {
	for {
		if c.closed == true {
			return
		}
		c.crawl()
		time.Sleep(time.Second * CRAWL_MACHINE_CRAWL_INTERVAL)
	}
}

func (c *CrawlMachine) crawl() {
	var err error
	var re *NetTransportDataRe
	// get a url
	re, err = c.sendandreturn(NET_TRANSPORT_OPERATE_URL_CRAWL_QUEUE_GET, NET_DATA_STATUS_NO, "", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// decode UrlBasic
	var url UrlBasic
	err = iendecode.BytesGobStruct(re.Data, &url)
	if err != nil {
		fmt.Println(err)
		return
	}
	// catch the page info
	resp, err := c.respGet(url.Url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	matched, _ := regexp.MatchString("text/html", resp.Header.Get("Content-Type"))
	if matched == true {
		// 处理是HTML的情况
		err = c.crawlHTML(resp, url)
	}
}

func (c *CrawlMachine) crawlHTML(resp *http.Response, url UrlBasic) (err error) {
	// get the urls in the page
	htmlbodyb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// to UTF-8
	htmlbody, err := charEncodeToUtf8(htmlbodyb)
	if err != nil {
		return
	}
	// trim the html, just leave the text
	htmltrim := trimHtml(htmlbody)
	htmltrim_sha1 := random.GetSha1Sum(htmltrim)
	if htmltrim_sha1 == url.Hash {
		// if the hash not change
		c.sendandreturn(NET_TRANSPORT_OPERATE_SEND_PAGE_DATA, NET_DATA_STATUS_PAGE_NOT_UPDATE, url.Domain, nil)
		return
	}
	// get the urls and send
	go c.crawlUrl(htmlbody)
	// get keywords
	keywords := getKeyword(htmlbody)
	// get headertitle
	title := getTitle(htmlbody)
	// send the page data
	page := &PageData{
		Url:         url.Url,
		UpTime:      time.Now(),
		Domain:      url.Domain,
		Spider:      c.identity_name,
		KeyWords:    keywords,
		HeaderTitle: title,
		BodyContent: htmltrim,
		Hash:        htmltrim_sha1,
	}
	page.New(url.Url)
	page_mid, err := roles.EncodeRoleToMiddle(page)
	if err != nil {
		return
	}
	page_mid_b, err := iendecode.StructGobBytes(page_mid)
	if err != nil {
		return
	}
	c.sendandreturn(NET_TRANSPORT_OPERATE_SEND_PAGE_DATA, NET_DATA_STATUS_PAGE_UPDATE, url.Domain, page_mid_b)
	return
}

func (c *CrawlMachine) crawlUrl(htmlbody string) {
	urls, err := getAllUrl(htmlbody)
	if err != nil {
		fmt.Println(err)
		return
	}
	// send the urls
	urls_b, err := iendecode.StructGobBytes(urls)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.sendandreturn(NET_TRANSPORT_OPERATE_URL_CRAWL_QUEUE_ADD, NET_DATA_STATUS_NO, "", urls_b)
	return
}

func (c *CrawlMachine) sendandreturn(operate NetTransportOperate, status NetDataStatus, domain string, data []byte) (re *NetTransportDataRe, err error) {
	ntd := NetTransportData{
		Name:    c.identity_name,
		Code:    c.identity_code,
		Operate: operate,
		Status:  status,
		Domain:  domain,
		Data:    data,
	}
	ntd_b, err := iendecode.StructGobBytes(ntd)
	if err != nil {
		return
	}
	process := c.transport.OpenProgress()
	defer process.Close()
	re_b, err := process.SendAndReturn(ntd_b)
	if err != nil {
		return
	}
	err = iendecode.BytesGobStruct(re_b, re)
	return
}

// http's Get
func (c *CrawlMachine) respGet(url string) (resp *http.Response, err error) {
	theAgentLen := len(UserAgent)
	theAgentNum := random.GetRandNum(theAgentLen - 1)
	theAgent := UserAgent[theAgentNum]
	theAcceptLanguageLen := len(AcceptLanguage)
	theAcceptLanguageNum := random.GetRandNum(theAcceptLanguageLen - 1)
	theAcceptLanguage := AcceptLanguage[theAcceptLanguageNum]

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", theAgent)
	req.Header.Set("Accept-Language", theAcceptLanguage)
	//req.Header.Set("Connection","keep-alive");
	resp, err = client.Do(req)
	return resp, err
}
