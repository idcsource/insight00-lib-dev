// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs

import (
	"errors"
	"runtime"
	"net/http"
	"reflect"
	"os"
	
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
	"github.com/idcsource/Insight-0-0-lib/cpool"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/idb"
)

// 建立一个新的Web，其中name是config中block的名字
func NewWeb (name string, config *cpool.ConfigPool, logs *ilogs.Logs) (*Web, error) {
	// 查看有没有这个name的block配置
	webc, werr := config.GetBlock(name);
	if werr != nil {
		return nil, werr;
	}
	
	// 查看配置里有无数据库配置文件
	var db *idb.DB;
	dbc, have := webc.GetSection("database");
	if have == nil {
		var err error;
		db, err = idb.NewDatabase(dbc, logs);
		if err != nil {
			return nil, err;
		}
	}
	
	// 创建路由
	router := NewRouter(config, logs);
	
	// 准备最大并发
	var max int64;
	var ok1 error;
	max, ok1 = webc.TranInt64("main.max_routine");
	if ok1 != nil {
		max = int64(runtime.NumCPU()) * MAX_ROUTINE_RATIO;
	}
	maxRoutine := make(chan bool,max);
	
	var web = &Web{
		Local : pubfunc.LocalPath(""),
		Database : db,
		Ext: make(map[string]interface{}),
		MultiDB : make(map[string]*idb.DB),
		Config : config,
		Log : logs,
		Router : router,
		SelfConfig : webc,
		MaxRoutine : maxRoutine,
	};
	return web, nil;
}

// 启动Web，需要在设置好路由、扩展等之后再调用
func (web *Web) Start() error {
	if web.Router.RouterOk == false {
		errs := errors.New("The router can not be set !");
		web.Log.ErrLog(errs);
		return errs;
	}
	go web.httpGo();
	return nil;
}

func (web *Web) httpGo () {
	var ifHttps bool;
	_, e1 := web.SelfConfig.GetSection("https");
	if e1 != nil {
		ifHttps = false;
	}
	var thecert, thekey string;
	if ifHttps == true {
		var e2, e3 error;
		thecert, e2 = web.SelfConfig.GetConfig("https.sslcert");
		thekey, e3 = web.SelfConfig.GetConfig("https.sslkey");
		if e2 != nil || e3 != nil {
			errs := errors.New("The SSL cert or key not be set !");
			web.Log.ErrLog(errs);
			return;
		}
	}
	var thePort string;
	thePort, e4 := web.SelfConfig.GetConfig("main.port");
	if e4 != nil {
		if ifHttps == false {
			thePort = "80"
		}else{
			thePort = "443";
		}
	}
	thePort = ":" + thePort;
	
	var err error;
	if ifHttps == true {
		err = http.ListenAndServeTLS(thePort, pubfunc.LocalFile(thecert), pubfunc.LocalFile(thekey), web);
	}else{
		err = http.ListenAndServe(thePort, web);
	}
	if err != nil {
		errs := errors.New("Can not start web service ! ");
		web.Log.ErrLog(errs);
		return;
	}
}

//HTTP的路由，提供给"net/http"包使用
func (web *Web) ServeHTTP(httpw http.ResponseWriter, httpr *http.Request) {
	//对进程数的控制
	web.MaxRoutine <- true;
	defer func(){
		<- web.MaxRoutine ;
	}()
	
	var runfloor reflect.Value;	//要运行的Floor
	urla, parameter := SplitUrl(httpr.URL.Path); //将获得的URL用斜线拆分成[]string
	//rt := Runtime{ AllRoutePath: httpr.URL.Path, NowRoutePath : urla, UrlRequest : parameter, ConfigTree : make(map[string]*goconfig.ConfigFile) };  //准备基本的RunTime
	rt := Runtime{ AllRoutePath: httpr.URL.Path, NowRoutePath : urla, UrlRequest : parameter };  //准备基本的RunTime
	
	
	//静态路由
	static, have := web.Router.GetStatic(httpr.URL.Path);
	if have == true {
		_, err := os.Stat(static);
		if err != nil {
			web.ToNotFoundHttp(httpw, httpr, rt);
		}else{
			http.ServeFile(httpw, httpr, static);
		}
		return;
	}
	
	//如果为0,则肯定为首页，则先处理掉首页
	if len(urla) == 0 {
		rt.RealNode = "";
		runfloor = web.Router.NodeTree.FloorValue;
	}else{
		runfloor, rt = web.Router.GetRunFloor(rt);
	}

	//开始执行
	in := make([]reflect.Value, 4);
	in[0] = reflect.ValueOf(httpw);
	in[1] = reflect.ValueOf(httpr);
	in[2] = reflect.ValueOf(web);
	in[3] = reflect.ValueOf(rt);
	runfloor.MethodByName("InitHTTP").Call(in);
	runfloor.MethodByName("ExecHTTP").Call(nil);
	return;
}

// 去执行NotFound，不要直接调用这个方法
func (web *Web) ToNotFoundHttp (w http.ResponseWriter, r *http.Request, rt Runtime) {
	runfloor := web.Router.NotFound;
	in := make([]reflect.Value, 4);
	in[0] = reflect.ValueOf(w);
	in[1] = reflect.ValueOf(r);
	in[2] = reflect.ValueOf(web);
	in[3] = reflect.ValueOf(rt);
	runfloor.MethodByName("InitHTTP").Call(in);
	runfloor.MethodByName("ExecHTTP").Call(nil);
	return;
}
