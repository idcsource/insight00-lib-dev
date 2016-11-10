// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package webs

import(
	"net/http"
	"fmt"
	"strings"
	"os"
	
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// FloorInterface 此为控制器接口的定义
type FloorInterface interface{
	InitHTTP(w http.ResponseWriter, r *http.Request, b *Web, rt Runtime)
	ExecHTTP()
}



//控制器原型的数据类型
type Floor struct {
	W http.ResponseWriter
	R *http.Request
	RT Runtime
	B *Web
}

// 函数负责对控制器进行初始化。
// 根据Runtime中的NowRoutePath，使用strings.Split函数根据“/”分割并作为Runtime中的UrlRequest（已清除可能的为空的字符串）
func (f *Floor) InitHTTP (w http.ResponseWriter, r *http.Request, b *Web,  rt Runtime) {
	f.W = w;
	f.R = r;
	f.RT = rt;
	f.B = b;
}

func (f *Floor) ExecHTTP() {
	
}


// 无法找到页面的系统内默认处理手段 
type NotFoundFloor struct{
	Floor
}

func (n *NotFoundFloor) ExecHTTP() {
	n.W.WriteHeader(404);
	fmt.Fprint(n.W,"404 Page Not Found");
	return;
}




// 静态文件的系统内默认处理手段
type StaticFileFloor struct{
	Floor
}

func (f *StaticFileFloor) ExecHTTP() {
	thepath, _ := f.RT.MyConfig.GetConfig("main.path");
	thepath = pubfunc.LocalPath(thepath);
		
	thefile := strings.Join(f.RT.NowRoutePath,"/");
	thefile = thepath + thefile;
		
	_, err := os.Stat(thefile);
	if err != nil {
		f.B.ToNotFoundHttp(f.W, f.R, f.RT);
	}else{
		http.ServeFile(f.W, f.R, thefile);
	}
}




// 空节点的处理手段 
type EmptyFloor struct{
	Floor
}

func (f *EmptyFloor) ExecHTTP() {
	f.B.ToNotFoundHttp(f.W, f.R, f.RT);
}




// FloorDoor的接口和数据类型
type FloorDoor map[string]FloorInterface

type FloorDoorInterface interface {
	FloorList()(FloorDoor)
}
