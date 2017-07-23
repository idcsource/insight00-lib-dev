// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 增强的角色类型。
//
// 继承自roles.Role，符合roles.Roleer接口，可以被rcontrol包管理，但无法保证能被正确保存。
// 如果不是用rcontrol中的NewRole()方法创建，则需要自行执行RolePlus的New()方法。
//
// 可处理birdges包中提供的数据桥的功能。
// 实现向桥发送数据，或者在收到桥中数据的时候按需要调用函数。
// 调用函数的原型为：funcname (key, id string, data interface{})。
// 其中key为发送方绑定的桥名，id为发送方的id，data为数据体，在设计函数时interface{}可以是你所确定的任何具体类型。
// 在执行StartBridge()启动桥功能之前，需要使用BridgeBind()将所有桥绑定好。
// 在执行StartBridge()的所有绑定都将无效。
//
// 实现nst包中的ConnExecer接口，也就是ExecTCP()方法，只是做个样子，无实际用处。
package rolesplus

import (
	"fmt"
	"reflect"

	"github.com/idcsource/insight00-lib/bridges"
	"github.com/idcsource/insight00-lib/ilogs"
	"github.com/idcsource/insight00-lib/nst"
	"github.com/idcsource/insight00-lib/roles"
)

type RolePlus struct {
	roles.Role                                // 继承roles.Roleer接口
	_bridge    map[string]*bridges.BridgeBind // 通讯桥
	_logs      *ilogs.Logs                    // 运行日志
}

// 新建自己
func (r *RolePlus) New(id string) {
	r.Role.New(id)
	r._bridge = make(map[string]*bridges.BridgeBind)
}

// 设置日志的接收
func (r *RolePlus) SetLog(logs *ilogs.Logs) {
	r._logs = logs
}

// 绑定一个桥
func (r *RolePlus) BridgeBind(name string, br *bridges.Bridge) {
	r._bridge[name] = br.Register(r.ReturnId())
}

// 向为name的桥中发送数据
func (r *RolePlus) BridgeSend(name string, data bridges.BridgeData) {
	r._bridge[name].Send <- data
}

// 返回全部的桥绑定
func (r *RolePlus) ReturnBridges() map[string]*bridges.BridgeBind {
	return r._bridge
}

// 返回日志
func (r *RolePlus) ReturnLog() *ilogs.Logs {
	return r._logs
}

// 启动接收数据的桥
func StartBridge(r RolePluser) {
	role_value := reflect.ValueOf(r)
	for key, br := range r.ReturnBridges() {
		go doRolesBridge(key, br, role_value, r.ReturnLog())
	}
}

// doRolesBridge 一个个启动桥
func doRolesBridge(key string, br *bridges.BridgeBind, role_value reflect.Value, logs *ilogs.Logs) {
	for {
		b_data, ok := <-br.Receive
		if ok == false {
			return
		} else {
			go doOneBridgeData(key, b_data, role_value, logs)
		}
	}
}

// doOneBridgeData 实际的执行接收数据的函数。
//
// 根据接收到的结构体，转交给operate指定的处理函数。
// operate指定的函数原型为：funcname (key, id string, data interface{})。
// 其中key为发送方绑定的桥名，id为发送方的id，data为数据体，在设计函数时interface{}可以是你所确定的任何具体类型。
func doOneBridgeData(key string, data bridges.BridgeData, role_value reflect.Value, logs *ilogs.Logs) {
	defer func() {
		if e := recover(); e != nil {
			if logs != nil {
				logs.ErrLog(e)
			} else {
				fmt.Println(e)
			}
		}
	}()
	operate := data.Operate
	id := data.Id
	databody := data.Data

	params := make([]reflect.Value, 3)
	params[0] = reflect.ValueOf(key)
	params[1] = reflect.ValueOf(id)
	params[2] = reflect.ValueOf(databody)
	role_value.MethodByName(operate).Call(params)
}

// ExecTCP nst的ConnExecer接口，只做样子
func (r *RolePlus) ExecTCP(ce *nst.ConnExec) error {
	return nil
}

// Logerr 做错误日志
func (r *RolePlus) ErrLog(err interface{}) {
	if err == nil {
		return
	}
	if r._logs != nil {
		r._logs.ErrLog(err)
	} else {
		fmt.Println(err)
	}
}

// Logerr 做运行日志
func (r *RolePlus) RunLog(err interface{}) {
	if err == nil {
		return
	}
	if r._logs != nil {
		r._logs.RunLog(err)
	} else {
		fmt.Println(err)
	}
}

// 返回期间的运行日志
func (r *RolePlus) ReRunLog() []string {
	return r._logs.ReRunLog()
}

// 返回期间的错误日志
func (r *RolePlus) ReErrLog() []string {
	return r._logs.ReErrLog()
}
