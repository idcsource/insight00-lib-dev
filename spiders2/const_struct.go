// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 这是抓取爬虫的第二个版本
package spiders2

import (
	"encoding/gob"
	"time"

	"github.com/idcsource/insight00-lib/cpool"
	"github.com/idcsource/insight00-lib/drule"
	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/roles"
)

// 索引状态
const (
	INDEX_STATUS_NO      = iota // 没有索引
	INDEX_STATUS_OK             // 已经索引
	INDEX_STATUS_RESTORE        // 已经再保存，spider的数据已经交给后段
	INDEX_STATUS_UP             // 需要更新索引
	INDEX_STATUS_NONEED         // 不需要索引
)

// 内部通讯状态
const (
	INTERCOM_OPERATE_NO       = iota // 没有
	INTERCOM_OPERATE_STOP            // 站点停止
	INTERCOM_OPERATE_RUN             // 站点运行
	INTERCOM_OPERATE_FINISHED        // 站点完成了一遍抓取
)

// 挂接的媒体类型，也就是Role中Friends的int，注意这里是int而不是uint8
const (
	MEDIA_TYPE_OUTSIDE = iota // 站外链接（只要是连接到站外的，无论js、css、图片等，都被称为站外连接）
	MEDIA_TYPE_OTHER          // 没有表明的其他
	MEDIA_TYPE_IMG            // 图片文件
	MEDIA_TYPE_FILE           // 普通文件
	MEDIA_TYPE_TEXT           // txt
	MEDIA_TYPE_VIDEO          // video
	MEDIA_TYPE_AUDIO          // audio
)

// 工作状态
const (
	WORK_STATUS_NO          = iota // 未知工作状态
	WORK_STATUS_NO_CONFIG          // 没有配置文件
	WORK_STATUS_HAVE_CONFIG        // 有配置文件
	WORK_STATUS_WORKING            // 工作中
	WORK_STATUS_STOPED             // 已经停止
)

// 抓取蜘蛛
type Spider struct {
	name          string
	sites         map[string]*siteMachine // 管理的站点，其中string是从配置文件而来的站点配置节点的名称
	work_status   uint8                   // 运行状态，对应WORK_STATUS_*
	config        *cpool.Block            // 配置文件
	roles_control *drule.TRule            // 角色管理器，这里用支持事务的drule包的TRule
	spider_signal chan siteStatusSignal   // 状态信号量，可以让站点机器在运行时给spider发送信息
	finishedlist  map[string]finishedSite // 完成抓取正在等待的站点列表
	logs          *ilogs.Logs             // 日志系统
}

// 完成一次抓取的站点列表
type finishedSite struct {
	finished_time time.Time // 完成时间
	next_start    int64     //下次开始时间，这里是小时
}

// 站点机器人
type siteMachine struct {
	site_name     string                // 站点名称
	spider_name   string                // 蜘蛛名称，来自Spider.name
	config        *cpool.Section        // 配置信息
	roles_control *drule.TRule          // 角色管理器，从Spider类型过来
	domains       []string              // 域名
	site_role     string                // 站点角色的名字
	now_status    uint8                 // 状态，对应WORK_STATUS_*
	spider_signal chan siteStatusSignal // 状态信号量，可以让站点机器在运行时给spider发送信息
	status_signal chan siteStatusSignal // 状态信号量
	logs          *ilogs.Logs           // 日志系统
}

// 站点的数据信息，做一个站点所有页面和媒体角色的父角色
type Site struct {
	roles.Role
	Domain []string // 来自SiteMachine
}

// 一个页面的数据信息
type PageData struct {
	roles.Role
	Url         string    // 页面的完整地址，带http等协议
	UpTime      time.Time // Update Time 最近更新时间
	UpInterval  int64     // 更新间隔
	Domain      string    // 域名
	Spider      string    // 爬虫机器身份标记，用以记录获取数据的机器是那一台
	KeyWord     []string  // 关键词，来自于页面
	Signature   string    // 内容签名，内容的SHA1散列值，主体，去除HTML标签后的
	HeaderTitle string    // 页面的标题，来自<header><title>
	BodyContent string    // 页面内容信息，来自<body>体
	AllContent  string    // 完整的页面内容
	IndexStatus uint8     // 索引状态
}

// 一个挂接的媒体数据信息
type MediaData struct {
	roles.Role
	Url         string    // 页面的完整地址
	UpTime      time.Time // Update Time 最近更新时间
	UpInterval  int64     // 更新间隔,单位天
	Domain      string    // 域名
	Spider      string    // 爬虫机器身份标记，用以记录获取数据的机器是那一台
	MediaType   int       // 媒体类型
	MediaName   string    // 媒体的文件名
	DataSaved   bool      // 是否保存了数据体
	DataBody    []byte    // 媒体的数据体，根据扩展名可以还原当初文件，按照设置决定是否保存媒体的数据体
	Signature   string    // 内容签名，完整内容的SHA1散列值
	IndexStatus uint8     // 索引状态，默认是不需要索引INDEX_STATUS_NONEED
}

// 内部statusSignal
type siteStatusSignal struct {
	sitename      string                // 站点名称
	operate       uint8                 // 操作，对应INTERCOM_OPERATE_*
	return_handle chan siteReturnHandle // 站点把返回的信号交给spider
}

// 站点机器的回执信号
type siteReturnHandle struct {
	status bool  // 状态，true一切正常，false就看err
	err    error // 错误
}

var UserAgent = []string{
	// "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.73 Safari/537.36",
	// "Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.59 Safari/537.36",
	// "Mozilla/5.0 (Linux; Android 6.0; HUAWEI EVA-AL10 Build/HUAWEIEVA-AL10) AppleWebKit/537.36(KHTML,like Gecko) Version/4.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
}

var AcceptLanguage = []string{
	"en-US,en;q=0.5",
}

// 注册"encoding/gob"所用的类型，如果调用了NewSpider()方法，则不需要再使用这个方法
func RegInterfaceForGob() {
	gob.Register(&Site{})
	gob.Register(&PageData{})
	gob.Register(&MediaData{})
}
