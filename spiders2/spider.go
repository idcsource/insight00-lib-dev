// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spiders2

import (
	"fmt"
	"strings"
	"time"

	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/drule"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 新建一个Spider
// 执行了RegInterfaceForGob()，初始化工作状态，不接收配置文件，返回*Spider
func NewSpider(name string, rc *drule.TRule, logs *ilogs.Logs) (spider *Spider) {
	RegInterfaceForGob()
	spider = &Spider{
		name:          name,
		roles_control: rc,
		logs:          logs,
		work_status:   WORK_STATUS_NO,
		spider_signal: make(chan siteStatusSignal),
		finishedlist:  make(map[string]finishedSite),
	}
	return spider
}

// 查看工作状态
func (s *Spider) ReturnWorkStatus() uint8 {
	return s.work_status
}

// 更新配置信息
func (s *Spider) UpdateConfig(config *cpool.Block) (err error) {
	if s.work_status == WORK_STATUS_WORKING {
		return fmt.Errorf("spiders2[Spider]UpdateConfig: The Spider is working, please first stop it using the method Stop().")
	}
	s.config = config
	s.work_status = WORK_STATUS_HAVE_CONFIG
	return
}

// 运行蜘蛛，如果在运行就直接返回
func (s *Spider) Start() (err error) {
	if s.work_status == WORK_STATUS_NO || s.work_status == WORK_STATUS_NO_CONFIG {
		return fmt.Errorf("spiders2[Spider]Start: The Spider have no any config.")
	}
	if s.work_status == WORK_STATUS_WORKING {
		fmt.Println("工作中")
		return
	}
	/* 下面是一步步开启蜘蛛的过程 */
	// 获取配置文件中的站点名称
	sites_name_list, err := s.config.GetEnum("main.sites")
	if err != nil {
		return fmt.Errorf("spiders2[Spider]Start: %v", err)
	}
	// 在这里初始化站点机器人
	s.sites = make(map[string]*siteMachine)
	// 遍历找出所有的站点具体的配置文件
	inside_err_array := make([]string, 0)
	fmt.Println("开始准备初始化机器")
	for i := range sites_name_list {
		// 获取站点的配置文件
		site_config, errn := s.config.GetSection(sites_name_list[i])
		if errn != nil {
			inside_err_array = append(inside_err_array, errn.Error())
			continue
		}
		// 创建站点机器，但这里并不启动
		// 站点机器人
		domains, errn := site_config.GetEnum("domains")
		if errn != nil {
			inside_err_array = append(inside_err_array, errn.Error())
			continue
		}
		site_role_name := s.name + "_" + sites_name_list[i]
		// 看有没有这个站点的角色
		have := s.roles_control.ExistRole(site_role_name)
		if have == false {
			// 出错被理解为没有，那就新建
			site_role := &Site{}
			// 站点的角色名为：spidername_sitename
			site_role.New(site_role_name)
			site_role.Domain = domains
			site_role.SetDataChanged()
			// 将站点角色保存进trule存储
			errn = s.roles_control.StoreRole(site_role)
			if errn != nil {
				inside_err_array = append(inside_err_array, errn.Error())
				continue
			}
		}
		s.sites[sites_name_list[i]] = &siteMachine{
			site_name:     sites_name_list[i],
			spider_name:   s.name,
			config:        site_config,
			roles_control: s.roles_control,
			domains:       domains,
			site_role:     site_role_name,
			now_status:    WORK_STATUS_STOPED,
			spider_signal: s.spider_signal,
			status_signal: make(chan siteStatusSignal),
			logs:          s.logs,
		}
	}
	fmt.Println("设置完成站点机器")
	// 如果有无法建立的，就返回错误
	if len(inside_err_array) != 0 {
		errstr := strings.Join(inside_err_array, " | ")
		err = fmt.Errorf(errstr)
		return
	}
	// 从现在开始才是启动所有的站点机器人
	// 开启站点监控，接收哪个站点完成了一次抓取遍历
	go s.runSiteFinishHandle()
	// 开启重启站点监控
	go s.runSiteRestartHandle()
	fmt.Println("开启站点机器")
	// 开启所有站点
	inside_err_array = make([]string, 0)
	for key := range s.sites {
		errn := s.sites[key].Start()
		if errn != nil {
			inside_err_array = append(inside_err_array, errn.Error())
		}
	}
	if len(inside_err_array) != 0 {
		errstr := strings.Join(inside_err_array, " | ")
		err = fmt.Errorf(errstr)
		return
	}
	fmt.Println("开启站点机器完成")
	// 工作状态修改为WORKING
	s.work_status = WORK_STATUS_WORKING
	return
}

// 开启站点监控
func (s *Spider) runSiteFinishHandle() {
	for {
		finished_sig := <-s.spider_signal
		// 如果不在工作就略过去
		if s.work_status != WORK_STATUS_WORKING {
			continue
		}
		finishedSite := finishedSite{
			finished_time: time.Now(),
			next_start:    24,
		}

		s.finishedlist[finished_sig.sitename] = finishedSite

	}
}

// 开启站点重启监控，默认每小时启动一次
func (s *Spider) runSiteRestartHandle() {
	for {
		time.Sleep(time.Hour)
		// 如果不在工作就略过去
		if s.work_status != WORK_STATUS_WORKING {
			continue
		}
		for sitename, finish := range s.finishedlist {
			thetime := finish.next_start * 60 * 60
			if finish.finished_time.Unix()+thetime > time.Now().Unix() {
				// 如果可以启动那就启动
				s.sites[sitename].Start()
				delete(s.finishedlist, sitename)
			}
		}
	}
}

// 停止蜘蛛，如果已经是停止状态就直接返回
func (s *Spider) Stop() (err error) {
	if s.work_status == WORK_STATUS_STOPED {
		return
	}
	// 置于STOP状态
	s.work_status = WORK_STATUS_STOPED
	// 清空等待重启列表
	s.finishedlist = make(map[string]finishedSite)
	// 创建发送信号
	signal := siteStatusSignal{
		operate: INTERCOM_OPERATE_STOP,
	}
	// 想办法对所有的站点发送停止命令
	for key := range s.sites {
		s.sites[key].status_signal <- signal
	}
	return
}
