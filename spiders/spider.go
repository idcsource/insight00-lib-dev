// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spiders

import (
	"time"
	"fmt"
	"encoding/gob"
	
	"github.com/idcsource/Insight-0-0-lib/random"
	"github.com/idcsource/Insight-0-0-lib/rcontrol"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/smcs"
	"github.com/idcsource/Insight-0-0-lib/rolesplus"
	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/cpool"
)

// 注册"encoding/gob"所用的类型，如果调用了NewSpider()方法，则不需要再使用这个方法
func RegInterfaceForGob(){
	gob.Register(&Site{});
	gob.Register(&PageData{});
	gob.Register(&MediaData{});
}

// 新建抓取蜘蛛。
//
// 参数说明：rolec为角色控制器，smcsn为smcs包提供的NodeSmcs，logs日志。
//
// 这里是作为启动的准备，所以只是处理了角色控制器（也就是存储）和日志。与SMCS进行通讯桥绑定，绑定名为SMCS_Node_Out_Bridge。
// 新建与站点机器（SiteMachine）沟通的内部通讯桥，绑定名为Spider_Inside_Bridge，发送出去的（给SiteMachine）要求处理的函数名为InsideCom。
// 而配置文件和其他的诸多状态都为空或关闭，等待运行时被动态修改。
func NewSpider (name string, rolec *rcontrol.RolesControl, smcsn *smcs.NodeSmcs, logs *ilogs.Logs) *Spider {
	gob.Register(&Site{});
	gob.Register(&PageData{});
	gob.Register(&MediaData{});
	
	insidebridge := bridges.NewBridge(logs);
	sp := &Spider{
		Name : name,
		RolesControl : rolec,
		NodeStatus : smcs.NODE_STATUS_NO_CONFIG,
		WorkStatus : smcs.WORK_SET_STOP,
		Sites : make(map[string]*SiteMachine),
		NodeSMCS : smcsn,
		insidebridge : insidebridge,
	};
	sp.New(random.Unid(1,"spider"));
	sp.SetLog(logs);
	sp.BridgeBind("SMCS_Node_Out_Bridge",smcsn.ReturnOutBridge());		// 与SMCS沟通的桥
	sp.BridgeBind("Spider_Inside_Bridge", insidebridge);				// 与内部其他进程沟通的桥
	rolesplus.StartBridge(sp);
	return sp;
}

// 开启蜘蛛。哪怕没有任何其他动作，也需要手动执行此方法才能让蜘蛛真正运行起来。
func (s *Spider) Start () {
	s.NodeSMCS.SetSendWorkSet(smcs.WORK_SET_STOP);
	s.NodeSMCS.SetSendType(smcs.NODE_TYPE_SPIDER);
	s.NodeSMCS.SetSendName(s.Name);
	s.NodeSMCS.SetSendStatus(smcs.NODE_STATUS_NO_CONFIG);
	go s.nodeSmcsLog();
}

// 接收SMCS-NODE回传服务器设置信息的方法，k和i这里被忽略不进行处理，这个函数名是在smcs包中指定的
func (s *Spider) SmcsReturn (k, i string, data smcs.CenterSend) {
	fmt.Println("收到了中心的信息");
	s.NodeStatus = smcs.NODE_STATUS_OK;
	s.NodeSMCS.SetSendStatus(smcs.NODE_STATUS_OK);
	ifrestart := false;
	ifstatus := false;
	ifconfig := false;
	if s.WorkStatus != data.NextWorkSet {
		s.WorkStatus = data.NextWorkSet;
		ifrestart = true;
		ifstatus = true;
	}
	//fmt.Println("收到的配置：",data);
	if data.NewConfig == true {
		spdata := cpool.NewBlock(s.Name,"");
		spdata.DecodeBlock(data.Config);
		s.Config = spdata;
		ifrestart = true;
		ifconfig = true;
	}
	/*
	if data.NextWorkSet == smcs.WORK_SET_NO {
		ifrestart = false;
	}*/
	if ifrestart == true {
		s.start(ifstatus, ifconfig);
	}
	
	// 如果状态改成了stop，则强制保存以下角色管理器
	if ifstatus == true && s.WorkStatus == smcs.WORK_SET_STOP {
		s.RolesControl.ToStore();
	}
}

// 启动蜘蛛的内部函数
func (s *Spider) start (ifstatus, ifconfig bool) {
	fmt.Println("启动蜘蛛");
	if ifconfig == true {
		// 获得配置中的站点名称
		sites_from_config, err := s.Config.GetEnum("main.site");
		if err != nil {
			s.ErrLog(fmt.Errorf("spiders: Spider: %v",err));
			return;
		}
		
		// 站点群与配置文件对比，将配置文件中没有的站点删除（发送DEL）
		for name, _ := range s.Sites {
			have := s.configHaveSite(name, sites_from_config);
			if have == false {
				inside_del_data := bridges.BridgeData{
					Id : s.ReturnId(),
					Operate : "InsideCom",
					Data : intercom{
						ToWhom : name,
						Operate : INTERCOM_OPERATE_DEL,
					},
				};
				s.BridgeSend("Spider_Inside_Bridge",inside_del_data);
				delete(s.Sites, name);
			}
		}
		
		// 对所有已经有的站点发送新的配置文件
		for name, _ := range s.Sites {
			siteconfig , err := s.Config.GetSection(name);
			if err != nil {
				continue;
			}
			inside_data := bridges.BridgeData{
				Id : s.ReturnId(),
				Operate : "InsideCom",
				Data : intercom{
					ToWhom : name,
					Operate : INTERCOM_OPERATE_CONFIG,
					Data : siteconfig,
				},
			};
			s.BridgeSend("Spider_Inside_Bridge",inside_data);
		}
		
		// 配置文件与站点群对比，将配置文件中有却没有新建的站点进行新建
		for _, name := range sites_from_config {
			_, find := s.Sites[name];
			if find == false {
				siteconfig , err := s.Config.GetSection(name);
				if err != nil {
					continue;
				}
				sitemachine, err2 := NewSiteMachine(name, s.Name, siteconfig, s.insidebridge, s.RolesControl, s.ReturnLog());
				if err2 != nil {
					s.ErrLog(fmt.Errorf("spiders: Spider: %v",err2));
					continue;
				}
				s.Sites[name] = sitemachine;
			}
		}
	}
	
	// 如果状态是STOP|RUN，则逐个向Site发送关闭指令
	if ifstatus == true {
		if s.WorkStatus == smcs.WORK_SET_START || s.WorkStatus == smcs.WORK_SET_STOP {
			var oprate uint8;
			if s.WorkStatus == smcs.WORK_SET_START {
				oprate = INTERCOM_OPERATE_RUN;
			} else {
				oprate = INTERCOM_OPERATE_STOP;
			}
			for name, _ := range s.Sites {
				siteconfig , err := s.Config.GetSection(name);
					if err != nil {
					continue;
				}
				inside_data := bridges.BridgeData{
					Id : s.ReturnId(),
					Operate : "InsideCom",
					Data : intercom{
						ToWhom : name,
						Operate : oprate,
						Data: siteconfig,
					},
				};
				s.BridgeSend("Spider_Inside_Bridge",inside_data);
			}
		}
	}
}

// 处理将日志交给SMCS
func (s *Spider) nodeSmcsLog () {
	for {
		runlog := s.ReRunLog();
		if runlog != nil {
			for _, one := range runlog {
				s.NodeSMCS.SetSendRunLog(one);
			}
		}
		errlog := s.ReErrLog();
		if errlog != nil {
			for _, one := range errlog {
				s.NodeSMCS.SetSendErrLog(one);
			}
		}
		time.Sleep(30 * time.Second);
	}
}

// 配置文件中是否指定了某个站点
func (s *Spider) configHaveSite(sitename string, config []string) bool {
	have := false;
	for _, name := range config {
		if name == sitename {
			have = true;
			break;
		}
	}
	return have;
}
