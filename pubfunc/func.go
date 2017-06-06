// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Insight 0+0各包共同使用的辅助函数
package pubfunc

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// 文件是否存在
func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// 返回执行文件的绝对路径
func GetCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return ret
}

// 判断目录名，如果不是“/”结尾就加上“/”
func DirMustEnd(dir string) string {
	matched, _ := regexp.MatchString("/$", dir)
	if matched == false {
		dir = dir + "/"
	}
	return dir
}

// 处理给出的路径地址，如果为相对路径就加上绝对路径
func LocalPath(path string) string {
	matched, _ := regexp.MatchString("^/", path)
	if matched == false {
		local := DirMustEnd(GetCurrPath())
		path = local + path
	}
	return DirMustEnd(path)
}

// 处理给出的文件地址，如果为相对路径就加上绝对路径
func LocalFile(path string) string {
	matched, _ := regexp.MatchString("^/", path)
	if matched == false {
		local := DirMustEnd(GetCurrPath())
		path = local + path
	}
	return path
}

// 路径必须以斜线开始
func PathMustBegin(path string) string {
	matched, _ := regexp.MatchString("^/", path)
	if matched == false {
		path = "/" + path
	}
	return path
}

// 生成绝对路径，如果给定的path不是绝对路径（以/开头），则用给定的abs补充并返回绝对路径（以/结尾）,否则直接返回path（并以/结尾）
func AbsolutePath(path, abs string) (absolute string) {
	matched, _ := regexp.MatchString("^/", path)
	if matched == false {
		abs = DirMustEnd(abs)
		path = abs + path
	}
	absolute = DirMustEnd(path)
	return
}

// 生成绝对文件路径，如果给定的path不是绝对路径（以/开头），则用给定的abs补充并返回绝对文件地址，否则直接返回file
func AbsoluteFile(file, abs string) (absolute string) {
	matched, _ := regexp.MatchString("^/", file)
	if matched == false {
		abs = DirMustEnd(abs)
		file = abs + file
	}
	absolute = file
	return
}

//将Url用斜线/拆分
// :id=dfad/:type=dafa/
func SplitUrl(url string) (urla []string, parameter map[string]string) {
	parameter = make(map[string]string)
	urlRequest := strings.Split(url, "/")
	matchP, _ := regexp.Compile("^:([A-Za-z0-9_-]+)=(.*)")
	for _, v := range urlRequest {
		if len(v) != 0 {
			if matchP.MatchString(v) {
				pa := matchP.FindStringSubmatch(v)
				parameter[pa[1]] = pa[2]
			} else {
				urla = append(urla, v)
			}
		}
	}
	return
}

// Is odd number
func IsOdd(num int) bool {
	if num%2 == 0 {
		return false
	}
	return true
}

func StringInSlice(list []string, s string) bool {
	sort.Strings(list)
	llen := len(list)
	i := sort.SearchStrings(list, s)
	if i < llen && list[i] == s {
		return true
	} else {
		return false
	}
}
