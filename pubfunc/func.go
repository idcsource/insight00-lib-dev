// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Insight 0+0各包共同使用的辅助函数
package pubfunc

import(
	"regexp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 文件是否存在
func FileExist (filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// 返回执行文件的绝对路径
func GetCurrPath() string {
    file, _ := exec.LookPath(os.Args[0]);
    path, _ := filepath.Abs(file);
    index := strings.LastIndex(path, string(os.PathSeparator));
    ret := path[:index];
    return ret;
}

// 判断目录名，如果不是“/”结尾就加上“/”
func DirMustEnd (dir string) string {
	matched , _ := regexp.MatchString("/$", dir)
	if matched == false {
		dir = dir + "/"
	}
	return dir
}

// 处理给出的路径地址，如果为相对路径就加上绝对路径
func LocalPath (path string) string {
	matched, _ := regexp.MatchString("^/", path);
	if matched == false {
		local := DirMustEnd(GetCurrPath());
		path = local + path;
	}
	return DirMustEnd(path);
}

// 处理给出的文件地址，如果为相对路径就加上绝对路径
func LocalFile (path string) string {
	matched, _ := regexp.MatchString("^/", path);
	if matched == false {
		local := DirMustEnd(GetCurrPath());
		path = local + path;
	}
	return path;
}
// 路径必须以斜线开始
func PathMustBegin(path string) (string){
	matched , _ := regexp.MatchString("^/", path)
	if matched == false {
		path = "/" + path
	}
	return path
}
