// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spiders

import (
	"net/http"
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
	"fmt"
	
	iconv "github.com/djimenez/iconv-go"
	"github.com/saintfish/chardet"
	"github.com/PuerkitoBio/goquery"
	
	"github.com/idcsource/Insight-0-0-lib/rcontrol"
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
)

// 由运行中的Spider调用，新建对某个站点的抓取。
//
// 参数：站点名, 蜘蛛名, 站点的配置, 内部通讯桥, 角色管理器, 日志
//
// 将自己绑定进内部通讯桥中，绑定名为intercom。
func NewSiteMachine (sitename, spidername string, config *cpool.Section, insidebridge *bridges.Bridge, rolec *rcontrol.RolesControl, logs *ilogs.Logs) (*SiteMachine, error) {
	fmt.Println("设置站点");
	// 生成Site
	rootname := random.GetSha1Sum(sitename);
	rootrole , err_ro := rolec.GetRole(rootname);
	if err_ro != nil {
		rootrole = rolec.NewRole(rootname, &Site{});
	}
	
	sm := &SiteMachine{
		SiteName : sitename,
		Spider : spidername,
		Config : config,
		RolesControl: rolec,
		Site : rootrole,
		StatusSignal : make(chan uint8),
	};
	sm.New( random.Unid(1,"SiteMachine") );
	sm.SetLog(logs);
	sm.BridgeBind("intercom",insidebridge);
	rolesplus.StartBridge(sm);
	
	return sm, nil;
}

// 用go start()的方式调用真正的站点抓取启动
func (sm *SiteMachine) Start() {
	go sm.start();
}

// 启动SiteMachine的方法。
// 将首先对url设定进行抓取，如果抓取成功，则开始新进程，按照配置文件进行抓取
func (sm *SiteMachine) start () {
	url, err := sm.Config.GetConfig("url");
	if err != nil {
		sm.ErrLog(fmt.Errorf("spiders: [SiteMachine]start: %v", err));
		return;
	}
	fmt.Println("启动站点：", url);
	
	// 开始读取第一个url
	var links []string;
	for retry:=0; retry<=5; retry++ {
		var err2 error;
		links, _, _, _, err2 = sm.crawlOne(true, url);
		if err2 != nil {
			fmt.Println("首先抓取第一个页面的错误：", err2);
			sm.ErrLog(fmt.Errorf("spiders: [SiteMachine]start: %v",err2));
			if retry == 5 {
				sm.ErrLog(fmt.Errorf("spiders: [SiteMachine]start: Retry timeout."));
				return;
			}else{
				time.Sleep(time.Minute);
			}
		}else{
			break;
		}
	}
	
	// 正式开始了
	for {
		fmt.Println("正是开始", url);
		select {
			case signal := <-sm.StatusSignal :
				// 如果收到关闭和删除的信号量的处理
				if signal == INTERCOM_OPERATE_DEL || signal == INTERCOM_OPERATE_STOP {
					return;
				}
			default:
				// 默认的处理，当然就是抓了
				fmt.Println("抓第一个页面", url);
				sig, finished := sm.saveAndCrawl(links);
				if sig == true {
					return;
				}
				if finished == true {
					// 完成一遍就睡觉，一睡就睡一天，并且让角色管理器保存一次
					sm.RolesControl.ToStore();
					fmt.Println("睡觉：", url);
					time.Sleep(24 * time.Hour);
				}
		}
	}
}

// 抓取以及保存，如果途中接收到关停信号，则终止推出并返回bool为true，如果完成了一遍遍历，则第二个bool成为true
func (sm *SiteMachine) saveAndCrawl (urls []string) (bool, bool) {
	sleeptime , err2 := sm.Config.TranInt64("sleep");
	if err2 != nil {
		sleeptime = 10;
	}
	
	for _, url := range urls {
		fmt.Println(url);
		select {
			case signal := <-sm.StatusSignal :
				// 如果受到信号量的处理
				fmt.Println("caveAndCrawl的信号");
				if signal == INTERCOM_OPERATE_DEL || signal == INTERCOM_OPERATE_STOP {
					return true, false;
				}
			default:
				links, ifop, ifnosleep, iflog, err := sm.crawlOne(false, url);
				if err != nil {
					errl := errors.New("spiders: SiteMachine: Can't crawl this url : " + url + " and the error is : " + fmt.Sprint(err));
					sm.ErrLog(errl);
					//time.Sleep(time.Duration(sleeptime) * time.Second);
					//continue;
				} else if ifop == true {
					if iflog == true {
						runl := "spiders: SiteMachine: Success crawl this url : " + url;
						sm.RunLog(runl);
					}
					//time.Sleep(time.Duration(sleeptime) * time.Second);
					//continue;
				} else {
					runl := "spiders: SiteMachine: Success crawl this url : " + url;
					sm.RunLog(runl);
					if len(links) > 0 {
						ifsig, _ := sm.saveAndCrawl(links);
						if ifsig == true {
							return true, false;
						}
					}
				}
				if ifnosleep == false {
					time.Sleep(time.Duration(sleeptime) * time.Second);
				}
		}
	}
	return false, true;
}

// 抓一个链接，enforce是强制执行，uint8为这个链接所代表的文件类型，roles.Roleer则是更新后的角色。
// []string为返回的所有链接。
// 如果第一个bool为true则这个链接不做理会。
// 如果第二个bool为true则不休息。
// 如果第三bool为true则强制写日志，主要是为了非HTML的情况。
func (sm *SiteMachine) crawlOne (enforce bool, url string) ([]string, bool, bool, bool, error) {
	if sm.domainAllow(url) == false {
		return nil, true, true, false, errors.New("The domain not be crawl.");
	}
	
	unid := random.GetSha1Sum(url);
	therole, err := sm.RolesControl.GetRole(unid);
	fmt.Println("crwalOne的找寻Role的错误：", err);
	if err == nil {
		// 找到了就走更新流程
		var thetime time.Time;
		err = sm.RolesControl.GetData(therole, "UpTime", &thetime);
		if err != nil {
			return nil, true, true, false, err;
		}
		var theUpInterval int64;
		err = sm.RolesControl.GetData(therole, "UpInterval", &theUpInterval);
		if err != nil {
			return nil, true, true, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err);
		}
		theUpInterval = theUpInterval * 60 * 60 * 24;
		if thetime.Unix() + theUpInterval > time.Now().Unix() || enforce == true {
			//走更新，否则放弃更新
			fmt.Println("是否走了更新", url);
			resp, err2 := sm.respGet(url);
			if err2 != nil {
				fmt.Println("resp的错误：", err2);
				return nil, true, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err2);
			}
			defer resp.Body.Close();
			if resp.StatusCode != 200 {
				//return nil, true, errors.New("Can't access the page " + url + " the code is " + resp.Status);
				fmt.Println("resp的状态：", resp.Status);
				return nil, true, false, false, nil;
			}
			matched, _ := regexp.MatchString("text/html", resp.Header.Get("Content-Type"));
			fmt.Println("HTML的头：", resp.Header.Get("Content-Type"));
			if matched == true {
				// 处理是HTML的情况
				fmt.Println("是否处理了HTML");
				link, ifop, err := sm.oprateHtml(enforce, url, resp, therole);
				fmt.Println("oprateHtml的错误：", err);
				if err != nil {
					return link, ifop, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err);
				}else{
					return link, ifop, false, false, nil;
				}
			} else {
				// 处理不是HTML的情况
				_, err3 := sm.oprateMedia(url, resp, therole);
				if err3 != nil {
					return nil, true, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err3);
				} else {
					return nil, true, false, false, nil;
				}
			}
		} else {
			return nil, true, true, false, nil;
		}
	} else {
		// 找不到就走新建流程
		resp, err2 := sm.respGet(url);
		if err2 != nil {
			return nil, true, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err);
		}
		defer resp.Body.Close();
		if resp.StatusCode != 200 {
			//return nil, true, errors.New("Can't access the page " + url + " the code is " + resp.Status);
			return nil, true, false, false, nil;
		}
		
		matched, _ := regexp.MatchString("text/html", resp.Header.Get("Content-Type"));
		if matched == true {
			// 处理是HTML的情况
			therole := sm.RolesControl.NewRole(unid, &PageData{
				Url : url,
				UpInterval : 1,
				IndexStatus : INDEX_STATUS_NO,
			});
			link, ifop, err := sm.oprateHtml(enforce, url, resp, therole);
			sm.RolesControl.RegChild(sm.Site, therole);
			if err != nil {
				return link, ifop, false, false, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err);
			}else {
				return link, ifop, false, false, nil;
			}
		} else {
			// 处理不是HTML的情况
			therole := sm.RolesControl.NewRole(unid, &MediaData{
				Url : url,
				UpInterval: 100,
				IndexStatus : INDEX_STATUS_NONEED,
			});
			ftype, err3 := sm.oprateMedia(url, resp, therole);
			sm.RolesControl.RegFriend(sm.Site, therole, ftype);
			if err3 != nil {
				return nil, true, false, true, fmt.Errorf("spider: [SiteMachine]crawlOne: %v",err3);
			} else {
				return nil, true, false, true, nil;
			}
		}
	}
	return nil, true, false, false, nil;
}

func (sm *SiteMachine) domainAllow (url string) bool {
	allow := false;
	for _, d := range sm.Domain {
		match, _ := regexp.MatchString("^(http|https|ftp)://" + d, url);
		if match == true {
			allow = true;
			break;
		}
	}
	return allow;
}

// 处理html，enforce为是否强制处理，返回值[]string为搜集到的链接，bool如果为true则说明这个没有更新不用理会，error则是错误
func (sm *SiteMachine) oprateHtml (enforce bool, url string, resp *http.Response, therole roles.Roleer) ([]string, bool, error) {
	htmlbodyb, err := ioutil.ReadAll(resp.Body);
	if err != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v",err);
	}
	// 将编码转为UTF-8
	htmlbody, err2 := sm.charEncodeToUtf8(htmlbodyb);
	if err2 != nil {
		return nil, true, fmt.Errorf("spider: [SiteMachine]oprateHtml: %v",err2);
	}
	// trim掉html中的所有格式，只保留文字
	htmltrim := sm.trimHtml(htmlbody);
	htmltrim_sha1 := random.GetSha1Sum(htmltrim);
	var old_signature string;
	sm.RolesControl.GetData(therole,"Signature",&old_signature);
	var old_UpInterval int64;
	sm.RolesControl.GetData(therole,"UpInterval",&old_UpInterval);
	if htmltrim_sha1 == old_signature && enforce == false {
		//如果发现内容没有变化，则让更新周期加长，并更新最新时间
		new_UpInterval := old_UpInterval + 1;
		if new_UpInterval >= 100 {
			new_UpInterval = 100;
		}
		sm.RolesControl.SetData(therole,"UpInterval",new_UpInterval);
		sm.RolesControl.SetData(therole,"UpTime",time.Now());
		return nil, true, nil;
	} else {
		// 如果发现内容变化了，则继续处理
		var new_UpInterval int64;
		new_UpInterval = old_UpInterval / 2;
		if new_UpInterval < 1 {
			new_UpInterval = 1;
		}
		sm.RolesControl.SetData(therole,"Url",url);
		sm.RolesControl.SetData(therole,"UpInterval",new_UpInterval);
		sm.RolesControl.SetData(therole,"UpTime",time.Now());
		sm.RolesControl.SetData(therole,"Signature",htmltrim_sha1);
		sm.RolesControl.SetData(therole,"BodyContent",htmltrim);
		sm.RolesControl.SetData(therole,"AllContent",htmlbody);
		sm.RolesControl.SetData(therole,"Domain",resp.Request.URL.Host);
		sm.RolesControl.SetData(therole,"Spider",sm.Spider);
		sm.RolesControl.SetData(therole,"HeaderTitle",sm.findTitle(htmlbody));
		sm.RolesControl.SetData(therole,"KeyWord",sm.findKeyword(htmlbody));
		
		var old_IndexStatus uint8;
		sm.RolesControl.GetData(therole,"IndexStatus",&old_IndexStatus);
		if old_IndexStatus == INDEX_STATUS_OK || old_IndexStatus == INDEX_STATUS_RESTORE {
			sm.RolesControl.SetData(therole,"IndexStatus",INDEX_STATUS_UP);
		}
		
		urls := sm.findSrcHref(resp, htmlbodyb);
		return urls, false, nil;
	}
}

// 处理其他文件
func (sm *SiteMachine) oprateMedia (url string, resp *http.Response, therole roles.Roleer) (ftype int64, err error) {
	max_size, err := sm.Config.TranInt64("mediasize");
	if err != nil {
		max_size = 0;
	}
	max_size = max_size * 1000;
	filesize := resp.ContentLength;
	filetype := strings.Split(resp.Header.Get("Content-Type"),"/");
	ftype_t := strings.TrimSpace(filetype[0]);
	switch ftype_t {
		case "image":
			ftype = MEDIA_TYPE_IMG;
		case "application":
			ftype = MEDIA_TYPE_FILE;
		case "text":
			ftype = MEDIA_TYPE_TEXT;
		case "video":
			ftype = MEDIA_TYPE_VIDEO;
		case "audio":
			ftype = MEDIA_TYPE_AUDIO;
		default :
			ftype = MEDIA_TYPE_OTHER;
	}
	sm.RolesControl.SetData(therole,"Url",url);
	sm.RolesControl.SetData(therole,"MediaType",ftype);
	sm.RolesControl.SetData(therole,"Domain",resp.Request.URL.Host);
	sm.RolesControl.SetData(therole,"Spider",sm.Spider);
	filename := sm.getFileName(resp.Request.URL.Path);
	sm.RolesControl.SetData(therole,"MediaName",filename);
	if filesize < 0 || filesize > max_size {
		sm.RolesControl.SetData(therole,"DataSaved",false);
		return ftype, nil;
	}
	bodyb, err2 := ioutil.ReadAll(resp.Body);
	if err2 != nil {
		sm.RolesControl.SetData(therole,"DataSaved",false);
		return ftype, fmt.Errorf("spider: [SiteMachine]oprateMedia: %v",err2);
	}
	sm.RolesControl.SetData(therole,"DataSaved",true);
	sm.RolesControl.SetData(therole,"DataBody",bodyb);
	sha1 := random.GetSha1SumBytes(bodyb);
	sm.RolesControl.SetData(therole,"Signature",sha1);
	sm.RolesControl.SetData(therole,"UpTime",time.Now());
	return ftype, nil;
}

// 根据连接获取文件名
func (sm *SiteMachine) getFileName(path string) string {
	paths := strings.Split(path,"/");
	lens := len(paths);
	return paths[lens - 1];
}

// 内部通讯系统的接收。
// 因为所有的Site都被注册进同一个内部通讯桥中，所以Site收到信息后首先判断是不是发给自己的。
func (sm *SiteMachine) InsideCom (k, i string, data intercom) {
	if data.ToWhom == sm.SiteName {
		switch data.Operate {
			case INTERCOM_OPERATE_CONFIG:
				// 配置文件更新的处理方法
				fmt.Println("配置文件更新的处理方法");
				sm.Config = data.Data.(*cpool.Section);
				sm.Domain, _ = sm.Config.GetEnum("domain");
			case INTERCOM_OPERATE_STOP:
				// 站点停止的方法
				sm.StatusSignal <- INTERCOM_OPERATE_STOP;
				sm.NowStatus = INTERCOM_OPERATE_STOP;
			case INTERCOM_OPERATE_DEL:
				// 站点删除的方法
				sm.StatusSignal <- INTERCOM_OPERATE_DEL;
				sm.NowStatus = INTERCOM_OPERATE_DEL;
			case INTERCOM_OPERATE_RUN:
				fmt.Println("站点运行的方法");
				if sm.NowStatus != INTERCOM_OPERATE_RUN {
					// 站点运行的方法
					sm.NowStatus = INTERCOM_OPERATE_RUN;
					sm.Config = data.Data.(*cpool.Section);
					sm.Domain, _ = sm.Config.GetEnum("domain");
					sm.start();
				}
		}
	}
}

// 将HTML中的乱七八遭标记全都删除掉，尽量只保留文本。
func (sm *SiteMachine) trimHtml (html string) string {
	html = strings.Replace(html,"\r","\n",-1);
	
	b1, _ := regexp.Compile(`(?isU)<HEAD>(.*)</head>`);
	html = b1.ReplaceAllString(html,"\n");
	
	b4, _ := regexp.Compile(`(?isU)<script(.*)</script>`);
	html = b4.ReplaceAllString(html,"\n");
	b6, _ := regexp.Compile(`(?isU)<style(.*)</style>`);
	html = b6.ReplaceAllString(html,"\n");
	//b2, _ := regexp.Compile("<(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b2, _ := regexp.Compile("<([^>]+)>");
	//b3, _ := regexp.Compile("</(html|body|div|span|ul|li|a|script|img)([^>]{0,})>");
	b3, _ := regexp.Compile("</([^>]+)>");
	html = b2.ReplaceAllString(html,"\n");
	html = b3.ReplaceAllString(html,"\n");
	
	//html = strings.Replace(html,"\t"," ",-1);
	//html = strings.Replace(html,"\n"," ",-1);
	
	b5, _ := regexp.Compile("([ ]{2,})");
	html = b5.ReplaceAllString(html," ");
	b7, _ := regexp.Compile(`([\t]{2,})`);
	html = b7.ReplaceAllString(html,"\n");
	b8, _ := regexp.Compile(`([　]{2,})`);
	html = b8.ReplaceAllString(html," ");
	b9, _ := regexp.Compile(`([ \n]{2,})`);
	html = b9.ReplaceAllString(html,"\n");
	
	b5_1, _ := regexp.Compile(`([\n]{2,})`);
	html = b5_1.ReplaceAllString(html,"\n");
	return html;
}

// 将字符编码转换成UTF-8
func (sm *SiteMachine) charEncodeToUtf8 (body []byte) (string, error) {
	cd := chardet.NewHtmlDetector();
	ccode , err := cd.DetectBest(body);
	if err != nil {
		return "", err;
	}
	var bodys string;
	if ccode.Charset != "UTF-8" {
		thecode := ccode.Charset;
		if strings.Contains(thecode,"GB-") {
			thecode = "GBK";
		}
		var err2 error;
		bodys, err2 = iconv.ConvertString(string(body), thecode, "utf-8");
		if err2 != nil {
			return "", err2;
		}
	}else{
		bodys = string(body);
	}
	return bodys, nil;
}

// 找到所有的页面链接
func (sm *SiteMachine) findSrcHref (resp *http.Response, body []byte) (href []string) {
	html := string(body);
	src, _ := regexp.Compile("(src|href)=([^ ><\t\n]+)");
	all := src.FindAllStringSubmatch(html,-1);
	href = make([]string, 0);
	for _, one := range all {
		oneurl := one[2];
		oneurl = strings.TrimSpace(oneurl);
		oneurl = strings.Trim(oneurl,"'");
		oneurl = strings.Trim(oneurl,"\"");
		oneurl = strings.TrimSpace(oneurl);
		if oneurl != "#" {
			oneurl, _ = sm.linkDomain(resp, oneurl);
			href = append(href, oneurl);
		}
	}
	return;
}

// 将所有的连接都变成绝对的，也就是头上都加上http什么的东西
func (sm *SiteMachine) linkDomain (resp *http.Response, hrefs string) (string, error) {
	host := resp.Request.URL.Host;
	path := resp.Request.URL.Path;
	//pattern := "^(http://|https://|ftp://)" + host + "(.*)";
	pattern := "^(http://|https://|ftp://)(.*)";
	href := hrefs;
	match , err := regexp.MatchString(pattern, href);
	if err != nil {
		return "", err;
	}
	protocol, find := sm.Config.GetConfig("protocol");
	if find != nil {
		protocol = "http";
	}
	if match == true {
		return href, nil;
	} else {
		pattern1 := "^/";
		match2 , err2 := regexp.MatchString(pattern1, href);
		if err2 != nil {
			return "", err2;
		}
		if match2 == true {
			href = protocol + "://" + host + href;
			return href, nil;
		} else {
			realpath := sm.realPath(path);
			href = protocol + "://" + host + realpath + href;
			return href, nil;
		}
	}
	return href, nil;
}

// 返回HTML的title部分
func (sm *SiteMachine) findTitle (html string) string {
	bodysreader := strings.NewReader(html);
	jquery, _ := goquery.NewDocumentFromReader(bodysreader);
	title := jquery.Find("title").Text();
	return title;
}

// 返回HTML中的meta name=keywords部分
func (sm *SiteMachine) findKeyword (html string) []string {
	bodysreader := strings.NewReader(html);
	jquery, _ := goquery.NewDocumentFromReader(bodysreader);
	keyword, _ := jquery.Find("meta[name=keywords]").Attr("content");
	keywords := strings.Split(keyword, ",");
	if len(keywords) == 1 {
		keywords = strings.Split(keyword, ";");
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, "，");
	} else if len(keywords) == 1 {
		keywords = strings.Split(keyword, " ");
	}
	return keywords;
}

func (sm *SiteMachine) realPath (path string) string {
	path_a := strings.Split(path,"/");
	var realpath string;
	con := len(path_a);
	if len(path_a[con-1]) != 0 {
		con = con - 1;
	}
	for i, one := range path_a {
		if i == con {
			break;
		}
		if len(one) != 0 {
			realpath += "/";
			realpath += one;
		}
	}
	realpath += "/";
	return realpath;
}

// 一个http方法中Get的封装，使用的是Client
func (sm *SiteMachine) respGet (url string) (resp *http.Response, err error) {
	theAgentLen := len(UserAgent);
	theAgentNum := random.GetRandNum(theAgentLen - 1);
	theAgent := UserAgent[theAgentNum];
	theAcceptLanguageLen := len(AcceptLanguage);
	theAcceptLanguageNum := random.GetRandNum(theAcceptLanguageLen - 1);
	theAcceptLanguage := AcceptLanguage[theAcceptLanguageNum];
	
	client := &http.Client{};
	req, err := http.NewRequest("GET", url, nil);
	if err != nil {
		return nil, err;
	}
	req.Header.Set("User-Agent",theAgent);
	req.Header.Set("Accept-Language",theAcceptLanguage);
	//req.Header.Set("Connection","keep-alive");
	resp, err = client.Do(req);
	return resp, err;
}
