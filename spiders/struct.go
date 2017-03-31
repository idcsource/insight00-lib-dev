// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 一套蜘蛛抓取的实现构建，需要配合配置文件使用。
//
// 配置文件示例见源代码中的spider.cfg文件。
package spiders

import (
	"time"

	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/rcontrol"
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/smcs2"
)

const (
	// 索引状态
	INDEX_STATUS_NO      = iota // 没有索引
	INDEX_STATUS_OK             // 已经索引
	INDEX_STATUS_RESTORE        // 已经再保存，spider的数据已经交给后段
	INDEX_STATUS_UP             // 需要更新索引
	INDEX_STATUS_NONEED         // 不需要索引
)

const (
	// 内部通讯状态
	INTERCOM_OPERATE_STOP   = iota // 站点停止
	INTERCOM_OPERATE_DEL           // 站点删除
	INTERCOM_OPERATE_RUN           // 站点运行
	INTERCOM_OPERATE_CONFIG        // 配置文件更新
)

const (
	// 挂接的媒体类型，也就是Role中Friends的int，注意这里是int而不是uint8
	MEDIA_TYPE_OUTSIDE = iota // 站外链接（只要是连接到站外的，无论js、css、图片等，都被称为站外连接）
	MEDIA_TYPE_OTHER          // 没有表明的其他
	MEDIA_TYPE_IMG            // 图片文件
	MEDIA_TYPE_FILE           // 普通文件
	MEDIA_TYPE_TEXT           // txt
	MEDIA_TYPE_VIDEO          // video
	MEDIA_TYPE_AUDIO          // audio
)

// 抓取蜘蛛
type Spider struct {
	rolesplus.RolePlus
	Name         string
	Sites        map[string]*SiteMachine // 管理的站点，其中string是从配置文件而来的站点配置节点的名称
	NodeStatus   uint8                   // 结点状态，对应smcs包中的工作状态（NODE_STATUS_*）
	WorkStatus   uint8                   // 运行状态，对应smcs包中的工作状态（WORK_SET_*）
	Config       *cpool.ConfigPool       // 配置文件
	NodeSMCS     *smcs2.NodeSmcs         // 配置蔓延
	RolesControl *rcontrol.RolesControl  // 角色管理器
	insidebridge *bridges.Bridge         // 内部通讯桥
}

// 站点机器人
type SiteMachine struct {
	rolesplus.RolePlus
	SiteName     string                 // 站点名称
	Spider       string                 // 蜘蛛名称
	Config       *cpool.Section         // 配置信息
	RolesControl *rcontrol.RolesControl // 角色管理器
	Domain       []string               // 域名
	Site         roles.Roleer           // 站点角色
	NowStatus    uint8                  // 状态，对应INTERCOM_OPERATE_*
	StatusSignal chan uint8             // 状态信号量，对应INTERCOM_OPERATE_*
}

// 站点的数据信息，做一个站点所有页面和媒体角色的父角色
type Site struct {
	roles.Role
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
	UpInterval  int64     // 更新间隔
	Domain      string    // 域名
	Spider      string    // 爬虫机器身份标记，用以记录获取数据的机器是那一台
	MediaType   int       // 媒体类型
	MediaName   string    // 媒体的文件名
	DataSaved   bool      // 是否保存了数据体
	DataBody    []byte    // 媒体的数据体，根据扩展名可以还原当初文件，按照设置决定是否保存媒体的数据体
	Signature   string    // 内容签名，完整内容的SHA1散列值
	IndexStatus uint8     // 索引状态，默认是不需要索引INDEX_STATUS_NONEED
}

// 内部通讯
type intercom struct {
	ToWhom  string      // 给谁的
	Operate uint8       // 指令
	Data    interface{} // 数据
}

var UserAgent = []string{
	"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.73 Safari/537.36",
	"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.59 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 6.0; HUAWEI EVA-AL10 Build/HUAWEIEVA-AL10) AppleWebKit/537.36(KHTML,like Gecko) Version/4.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
}

var AcceptLanguage = []string{
	"en-US,en;q=0.5",
}
