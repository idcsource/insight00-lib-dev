// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spiders2

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"github.com/saintfish/chardet"
	opencc "github.com/stevenyao/go-opencc"

	"github.com/idcsource/Insight-0-0-lib/drule"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 启动站点机器
func (s *siteMachine) Start() (err error) {
	// 获取初始入口的地址
	url, err := s.config.GetConfig("url")
	if err != nil {
		return
	}
	// 开始抓第一个url，尝试连接5次
	var links []string
	for retry := 0; retry <= 5; retry++ {
		var err2 error
		links, _, _, _, err2 = s.crawlOne(true, url)
		if err2 != nil {
			return err2
			if retry == 5 {
				return fmt.Errorf("spiders2[SiteMachine]Start: Retry timeout.")
			} else {
				time.Sleep(30 * time.Second)
			}
		} else {
			break
		}
	}
	// 正式开始
	s.now_status = WORK_STATUS_WORKING
	fmt.Println(url + "抓玩了")
	go s.gostart(links)
	return
}

// 只进行一遍，一遍之后就设置为停止工作
func (s *siteMachine) gostart(links []string) {
	// 正式开始
	sig, finished := s.saveAndCrawl(links)
	if sig == true {
		return
		s.now_status = WORK_STATUS_STOPED
	}
	if finished == true {
		signal := siteStatusSignal{
			sitename: s.site_name,
			operate:  INTERCOM_OPERATE_FINISHED,
		}
		s.spider_signal <- signal
		s.now_status = WORK_STATUS_STOPED
		return
	}
}

// 抓取以及保存，如果途中接收到关停信号，则终止推出并返回bool为true，如果完成了一遍遍历，则第二个bool成为true
func (s *siteMachine) saveAndCrawl(urls []string) (bool, bool) {
	sleeptime, err2 := s.config.TranInt64("sleep")
	if err2 != nil {
		sleeptime = 10
	}

	for _, url := range urls {
		//fmt.Println(url)
		select {
		case signal := <-s.status_signal:
			// 如果收到信号量的处理
			fmt.Println("caveAndCrawl的信号")
			if signal.operate == INTERCOM_OPERATE_STOP {
				return true, false
			}
		default:
			links, ifop, ifnosleep, iflog, err := s.crawlOne(false, url)
			if err != nil {
				errl := errors.New("spiders: SiteMachine: Can't crawl this url : " + url + " and the error is : " + fmt.Sprint(err))
				s.logs.ErrLog(errl)
				//time.Sleep(time.Duration(sleeptime) * time.Second);
				//continue;
			} else if ifop == true {
				if iflog == true {
					runl := "spiders: SiteMachine: Success crawl this url : " + url
					s.logs.RunLog(runl)
				}
				//time.Sleep(time.Duration(sleeptime) * time.Second);
				//continue;
			} else {
				runl := "spiders: SiteMachine: Success crawl this url : " + url
				s.logs.RunLog(runl)
				if len(links) > 0 {
					ifsig, _ := s.saveAndCrawl(links)
					if ifsig == true {
						return true, false
					}
				}
			}
			if ifnosleep == false {
				time.Sleep(time.Duration(sleeptime) * time.Second)
			}
		}
	}
	return false, true
}

// 抓一个链接，enforce是强制执行，uint8为这个链接所代表的文件类型，roles.Roleer则是更新后的角色。
// []string为返回的所有链接。
// 如果第一个bool为true则这个链接不做理会。
// 如果第二个bool为true则不休息。
// 如果第三bool为true则强制写日志，主要是为了非HTML的情况。
func (s *siteMachine) crawlOne(enforce bool, url string) ([]string, bool, bool, bool, error) {
	if s.domainAllow(url) == false {
		return nil, true, true, false, errors.New("The domain not be crawl: " + url)
	}
	_, err := s.roles_control.ReadRole(url)
	if err == nil {
		// 找到了就走更新流程，开启日志
		tran, _ := s.roles_control.Prepare(url)
		var thetime time.Time
		err = tran.ReadData(url, "UpTime", &thetime)
		if err != nil {
			tran.Rollback()
			return nil, true, true, false, err
		}
		var theUpInterval int64
		err = tran.ReadData(url, "UpInterval", &theUpInterval)
		if err != nil {
			tran.Rollback()
			return nil, true, true, false, err
		}
		theUpInterval = theUpInterval * 60 * 60 * 24
		if thetime.Unix()+theUpInterval > time.Now().Unix() || enforce == true {
			//走更新，否则放弃更新
			resp, err2 := s.respGet(url)
			if err2 != nil {
				tran.Rollback()
				return nil, true, false, false, err2
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				//return nil, true, errors.New("Can't access the page " + url + " the code is " + resp.Status);
				//fmt.Println("resp的状态：", resp.Status)
				tran.Rollback()
				return nil, true, false, false, nil
			}
			matched, _ := regexp.MatchString("text/html", resp.Header.Get("Content-Type"))
			//fmt.Println("HTML的头：", resp.Header.Get("Content-Type"))
			if matched == true {
				// 处理是HTML的情况
				//fmt.Println("是否处理了HTML")
				link, ifop, err := s.oprateHtml(enforce, url, resp, tran)
				fmt.Println("oprateHtml的错误：", err)
				if err != nil {
					tran.Rollback()
					return link, ifop, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err)
				} else {
					tran.Commit()
					return link, ifop, false, false, nil
				}
			} else {
				// 处理不是HTML的情况
				_, err3 := s.oprateMedia(url, resp, tran)
				if err3 != nil {
					tran.Rollback()
					return nil, true, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err3)
				} else {
					tran.Commit()
					return nil, true, false, false, nil
				}
			}
		} else {
			tran.Rollback()
			return nil, true, true, false, nil
		}
	} else {
		// 找不到就走新建流程
		resp, err2 := s.respGet(url)
		if err2 != nil {
			return nil, true, false, false, err2
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			//return nil, true, errors.New("Can't access the page " + url + " the code is " + resp.Status);
			return nil, true, false, false, nil
		}

		matched, _ := regexp.MatchString("text/html", resp.Header.Get("Content-Type"))
		if matched == true {
			// 处理是HTML的情况
			therole := &PageData{
				Url:         url,
				UpInterval:  1,
				IndexStatus: INDEX_STATUS_NO,
			}
			therole.New(url)
			therole.SetDataChanged()
			tran, _ := s.roles_control.Begin()
			tran.StoreRole(therole)
			link, ifop, err := s.oprateHtml(enforce, url, resp, tran)
			if err != nil {
				tran.Rollback()
				return link, ifop, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err)
			}
			err = tran.WriteChild(s.site_name, url)
			if err != nil {
				tran.Rollback()
				return link, ifop, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err)
			} else {
				tran.Commit()
				return link, ifop, false, false, nil
			}
		} else {
			// 处理不是HTML的情况
			therole := &MediaData{
				Url:         url,
				UpInterval:  100,
				IndexStatus: INDEX_STATUS_NONEED,
			}
			therole.New(url)
			therole.SetDataChanged()
			tran, _ := s.roles_control.Begin()
			tran.StoreRole(therole)
			ftype, err3 := s.oprateMedia(url, resp, tran)
			if err3 != nil {
				tran.Rollback()
				return nil, true, false, true, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err3)
			}
			err = tran.WriteFriend(s.site_name, url, ftype)
			if err != nil {
				tran.Rollback()
				return nil, true, false, true, fmt.Errorf("spider: [SiteMachine]crawlOne: %v", err)
			} else {
				tran.Commit()
				return nil, true, false, true, nil
			}
		}
	}
	return nil, true, false, false, nil
}

// 这个连接的domain是不是在允许内
func (s *siteMachine) domainAllow(url string) bool {
	allow := false
	fmt.Println(s.domains)
	for _, d := range s.domains {
		match, _ := regexp.MatchString("^(http|https|ftp)://"+d, url)
		if match == true {
			allow = true
			break
		}
	}
	return allow
}

// 一个http方法中Get的封装，使用的是Client
func (s *siteMachine) respGet(url string) (resp *http.Response, err error) {
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

// 处理html，enforce为是否强制处理，返回值[]string为搜集到的链接，bool如果为true则说明这个没有更新不用理会，error则是错误
func (s *siteMachine) oprateHtml(enforce bool, url string, resp *http.Response, thetran *drule.Transaction) ([]string, bool, error) {
	htmlbodyb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
	}
	// 将编码转为UTF-8
	htmlbody, err2 := s.charEncodeToUtf8(htmlbodyb)
	if err2 != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err2)
	}
	// trim掉html中的所有格式，只保留文字
	htmltrim := s.trimHtml(htmlbody)
	htmltrim_sha1 := random.GetSha1Sum(htmltrim)
	var old_signature string
	err = s.roles_control.ReadData(url, "Signature", &old_signature)
	if err != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
	}
	var old_UpInterval int64
	err = s.roles_control.ReadData(url, "UpInterval", &old_UpInterval)
	if err != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
	}
	if htmltrim_sha1 == old_signature && enforce == false {
		//如果发现内容没有变化，则让更新周期加长，并更新最新时间
		new_UpInterval := old_UpInterval + 1
		if new_UpInterval >= 100 {
			new_UpInterval = 100
		}
		err = s.roles_control.WriteData(url, "UpInterval", new_UpInterval)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "UpTime", time.Now())
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		return nil, true, nil
	} else {
		// 如果发现内容变化了，则继续处理
		var new_UpInterval int64
		new_UpInterval = old_UpInterval / 2
		if new_UpInterval < 1 {
			new_UpInterval = 1
		}
		err = s.roles_control.WriteData(url, "Url", url)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "UpInterval", new_UpInterval)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "UpTime", time.Now())
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "Signature", htmltrim_sha1)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "BodyContent", htmltrim)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "AllContent", htmlbody)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "Domain", resp.Request.URL.Host)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "Spider", s.spider_name)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "HeaderTitle", s.findTitle(htmlbody))
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		err = s.roles_control.WriteData(url, "KeyWord", s.findKeyword(htmlbody))
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}

		var old_IndexStatus uint8
		err = s.roles_control.ReadData(url, "IndexStatus", &old_IndexStatus)
		if err != nil {
			return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
		}
		if old_IndexStatus == INDEX_STATUS_OK || old_IndexStatus == INDEX_STATUS_RESTORE {
			err = s.roles_control.WriteData(url, "IndexStatus", INDEX_STATUS_UP)
			if err != nil {
				return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v", err)
			}
		}

		urls := s.findSrcHref(resp, htmlbodyb)
		return urls, false, nil
	}
}

// 处理其他文件
func (s *siteMachine) oprateMedia(url string, resp *http.Response, tran *drule.Transaction) (ftype int64, err error) {
	max_size, err := s.config.TranInt64("mediasize")
	if err != nil {
		max_size = 0
	}
	max_size = max_size * 1000
	filesize := resp.ContentLength
	filetype := strings.Split(resp.Header.Get("Content-Type"), "/")
	ftype_t := strings.TrimSpace(filetype[0])
	switch ftype_t {
	case "image":
		ftype = MEDIA_TYPE_IMG
	case "application":
		ftype = MEDIA_TYPE_FILE
	case "text":
		ftype = MEDIA_TYPE_TEXT
	case "video":
		ftype = MEDIA_TYPE_VIDEO
	case "audio":
		ftype = MEDIA_TYPE_AUDIO
	default:
		ftype = MEDIA_TYPE_OTHER
	}
	err = s.roles_control.WriteData(url, "Url", url)
	if err != nil {
		return
	}
	err = s.roles_control.WriteData(url, "MediaType", ftype)
	if err != nil {
		return
	}
	err = s.roles_control.WriteData(url, "Domain", resp.Request.URL.Host)
	if err != nil {
		return
	}
	err = s.roles_control.WriteData(url, "Spider", s.spider_name)
	if err != nil {
		return
	}
	filename := s.getFileName(resp.Request.URL.Path)
	err = s.roles_control.WriteData(url, "MediaName", filename)
	if err != nil {
		return
	}
	if filesize < 0 || filesize > max_size {
		err = s.roles_control.WriteData(url, "DataSaved", false)
		if err != nil {
			return
		}
		return ftype, nil
	}
	bodyb, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		err = s.roles_control.WriteData(url, "DataSaved", false)
		if err != nil {
			return
		}
		return ftype, fmt.Errorf("spider: [SiteMachine]oprateMedia: %v", err2)
	}
	err = s.roles_control.WriteData(url, "DataSaved", true)
	if err != nil {
		return
	}
	err = s.roles_control.WriteData(url, "DataBody", bodyb)
	if err != nil {
		return
	}
	sha1 := random.GetSha1SumBytes(bodyb)
	err = s.roles_control.WriteData(url, "Signature", sha1)
	if err != nil {
		return
	}
	err = s.roles_control.WriteData(url, "UpTime", time.Now())
	if err != nil {
		return
	}
	return ftype, nil
}

// 根据连接获取文件名
func (s *siteMachine) getFileName(path string) string {
	paths := strings.Split(path, "/")
	lens := len(paths)
	return paths[lens-1]
}

// 将HTML中的乱七八遭标记全都删除掉，尽量只保留文本。
func (s *siteMachine) trimHtml(html string) string {
	html = strings.Replace(html, "\r", "\n", -1)

	b1, _ := regexp.Compile(`(?isU)<HEAD>(.*)</head>`)
	html = b1.ReplaceAllString(html, "\n")

	b4, _ := regexp.Compile(`(?isU)<script(.*)</script>`)
	html = b4.ReplaceAllString(html, "\n")
	b6, _ := regexp.Compile(`(?isU)<style(.*)</style>`)
	html = b6.ReplaceAllString(html, "\n")
	//b2, _ := regexp.Compile("<(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b2, _ := regexp.Compile("<([^>]+)>")
	//b3, _ := regexp.Compile("</(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b3, _ := regexp.Compile("</([^>]+)>")
	html = b2.ReplaceAllString(html, "\n")
	html = b3.ReplaceAllString(html, "\n")

	//html = strings.Replace(html,"\t"," ",-1);
	//html = strings.Replace(html,"\n"," ",-1);

	b5, _ := regexp.Compile("([ ]{2,})")
	html = b5.ReplaceAllString(html, " ")
	b7, _ := regexp.Compile(`([\t]{2,})`)
	html = b7.ReplaceAllString(html, "\n")
	b8, _ := regexp.Compile(`([　]{2,})`)
	html = b8.ReplaceAllString(html, " ")
	b9, _ := regexp.Compile(`([ \n]{2,})`)
	html = b9.ReplaceAllString(html, "\n")

	b5_1, _ := regexp.Compile(`([\n]{2,})`)
	html = b5_1.ReplaceAllString(html, "\n")

	// 将繁体中文转成简体中文
	config_t2s := "/usr/share/opencc/t2s.json"
	c := opencc.NewConverter(config_t2s)
	defer c.Close()
	html = c.Convert(html)

	return html
}

// 将字符编码转换成UTF-8
func (s *siteMachine) charEncodeToUtf8(body []byte) (string, error) {
	cd := chardet.NewHtmlDetector()
	ccode, err := cd.DetectBest(body)
	if err != nil {
		return "", err
	}
	var bodys string
	if ccode.Charset != "UTF-8" {
		thecode := ccode.Charset
		if strings.Contains(thecode, "GB-") {
			thecode = "GBK"
		}
		var err2 error
		bodys, err2 = iconv.ConvertString(string(body), thecode, "utf-8")
		if err2 != nil {
			return "", err2
		}
	} else {
		bodys = string(body)
	}
	return bodys, nil
}

// 找到所有的页面链接
func (s *siteMachine) findSrcHref(resp *http.Response, body []byte) (href []string) {
	html := string(body)
	src, _ := regexp.Compile("(src|href)=([^ ><\t\n]+)")
	all := src.FindAllStringSubmatch(html, -1)
	href = make([]string, 0)
	for _, one := range all {
		oneurl := one[2]
		oneurl = strings.TrimSpace(oneurl)
		oneurl = strings.Trim(oneurl, "'")
		oneurl = strings.Trim(oneurl, "\"")
		oneurl = strings.TrimSpace(oneurl)
		if oneurl != "#" {
			oneurl, _ = s.linkDomain(resp, oneurl)
			href = append(href, oneurl)
		}
	}
	return
}

// 将所有的连接都变成绝对的，也就是头上都加上http什么的东西
func (s *siteMachine) linkDomain(resp *http.Response, hrefs string) (string, error) {
	host := resp.Request.URL.Host
	path := resp.Request.URL.Path
	//pattern := "^(http://|https://|ftp://)" + host + "(.*)";
	pattern := "^(http://|https://|ftp://)(.*)"
	href := hrefs
	match, err := regexp.MatchString(pattern, href)
	if err != nil {
		return "", err
	}
	protocol, find := s.config.GetConfig("protocol")
	if find != nil {
		protocol = "http"
	}
	if match == true {
		return href, nil
	} else {
		pattern1 := "^/"
		match2, err2 := regexp.MatchString(pattern1, href)
		if err2 != nil {
			return "", err2
		}
		if match2 == true {
			href = protocol + "://" + host + href
			return href, nil
		} else {
			realpath := s.realPath(path)
			href = protocol + "://" + host + realpath + href
			return href, nil
		}
	}
	return href, nil
}

// 返回HTML的title部分
func (s *siteMachine) findTitle(html string) string {
	bodysreader := strings.NewReader(html)
	jquery, _ := goquery.NewDocumentFromReader(bodysreader)
	title := jquery.Find("title").Text()
	return title
}

// 返回HTML中的meta name=keywords部分
func (s *siteMachine) findKeyword(html string) []string {
	bodysreader := strings.NewReader(html)
	jquery, _ := goquery.NewDocumentFromReader(bodysreader)
	keyword, _ := jquery.Find("meta[name=keywords]").Attr("content")
	keywords := strings.Split(keyword, ",")
	if len(keywords) == 1 {
		keywords = strings.Split(keyword, ";")
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, "，")
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, " ")
	}
	return keywords
}

func (s *siteMachine) realPath(path string) string {
	path_a := strings.Split(path, "/")
	var realpath string
	con := len(path_a)
	if len(path_a[con-1]) != 0 {
		con = con - 1
	}
	for i, one := range path_a {
		if i == con {
			break
		}
		if len(one) != 0 {
			realpath += "/"
			realpath += one
		}
	}
	realpath += "/"
	return realpath
}
