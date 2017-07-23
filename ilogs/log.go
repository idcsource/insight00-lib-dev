// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 一套简单的运行日志和错误日志的记录和处理
package ilogs

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/idcsource/insight00-lib/pubfunc"
)

// ilogs日志结构
type Logs struct {
	runFile *forsmcslog //运行日志For SMCS
	runLog  *log.Logger //运行日志
	errFile *forsmcslog //错误日志For SMCS
	errLog  *log.Logger //错误日志
	forsmcs bool        //是否为smcs准备数据
}

func NewLog(run, el, prefix string) (logs *Logs, err error) {
	return newLog(run, el, prefix, false)
}

func NewLogForSmcs(run, el, prefix string) (logs *Logs, err error) {
	return newLog(run, el, prefix, true)
}

// 创建新的日志，forsmcs如果为true，则可以使用ReErrLog()和ReRunLog()方法返回在上次调用这两个函数时所产生的日志条目。
func newLog(run, el, prefix string, forsmcs bool) (logs *Logs, err error) {
	prefix = prefix + " "
	var rlogw io.Writer
	// 如果没有指定文件，则打开一个不存在的东西
	if len(run) == 0 {
		rlogw = &nolog{}
	} else {
		run = pubfunc.LocalFile(run)
		rlogw, err = os.OpenFile(run, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			return
		}
	}
	var runlogs *log.Logger
	smcslogrun := newForsmcslog(rlogw)
	if forsmcs == true {
		runlogs = log.New(smcslogrun, prefix, log.Ldate|log.Ltime)
	} else {
		runlogs = log.New(rlogw, prefix, log.Ldate|log.Ltime)
	}

	var elogw io.Writer
	if len(el) == 0 {
		elogw = &nolog{}
	} else {
		el = pubfunc.LocalFile(el)
		elogw, err = os.OpenFile(el, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			return
		}
	}
	var errlogs *log.Logger
	smcslogerr := newForsmcslog(elogw)
	if forsmcs == true {
		errlogs = log.New(smcslogerr, prefix, log.Ldate|log.Ltime)
	} else {
		errlogs = log.New(elogw, prefix, log.Ldate|log.Ltime)
	}

	if forsmcs == true {
		logs = &Logs{
			runFile: smcslogrun,
			runLog:  runlogs,
			errFile: smcslogerr,
			errLog:  errlogs,
			forsmcs: true,
		}
	} else {
		logs = &Logs{
			runLog:  runlogs,
			errLog:  errlogs,
			forsmcs: false,
		}
	}
	return
}

// 加入一条运行日志
func (l *Logs) RunLog(s ...interface{}) {
	l.runLog.Print(s...)
}

// 加入一条错误日志
func (l *Logs) ErrLog(s ...interface{}) {
	l.errLog.Print(s...)
}

// 返回期间的运行日志
func (l *Logs) ReRunLog() []string {
	if l.forsmcs == true {
		return l.runFile.ReSMCS()
	} else {
		return nil
	}
}

// 返回期间的错误日志
func (l *Logs) ReErrLog() []string {
	if l.forsmcs == true {
		return l.errFile.ReSMCS()
	} else {
		return nil
	}
}

type nolog struct {
	n string
}

func (nl *nolog) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	fmt.Println(string(p))
	return
}

type forsmcslog struct {
	file  io.Writer
	smcsd []string
}

func newForsmcslog(file io.Writer) *forsmcslog {
	return &forsmcslog{file: file, smcsd: make([]string, 0)}
}

func (f *forsmcslog) Write(p []byte) (n int, err error) {
	f.smcsd = append(f.smcsd, string(p))
	fmt.Fprint(f.file, string(p))
	n = len(p)
	err = nil
	return
}

func (f *forsmcslog) ReSMCS() []string {
	ther := make([]string, 0)
	for _, one := range f.smcsd {
		ther = append(ther, one)
	}

	f.smcsd = make([]string, 0)
	return ther
}
