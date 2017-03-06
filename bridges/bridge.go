// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Bridges 数据桥。负责在各个子进程间极端自由的扩散数据。
package bridges

import (
	"fmt"
	"reflect"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 桥结构
type Bridge struct {
	// 发送方
	send chan BridgeData
	// 接收方注册，string为接收方Id
	receive map[string]chan BridgeData
	// 关闭检测
	closeb chan bool
	// 日志
	logs *ilogs.Logs
}

// 桥的传递数据的结构
type BridgeData struct {
	// 判断是谁发送的信息
	Id string
	// 关闭吗？
	Close bool
	// 操作哪个函数
	Operate string
	// 实际的数据主体
	Data interface{}
}

// 桥绑定
type BridgeBind struct {
	// 发送
	Send chan<- BridgeData
	// 接收
	Receive <-chan BridgeData
}

// 桥操作
type BridgeOperate struct {
	// 桥中的绑定名
	id string
	// 桥绑定
	bridgeBind map[string]*BridgeBind
	// 对象
	class reflect.Value
	// 日志
	logs *ilogs.Logs
}

// 创建一个桥
func NewBridge(logs *ilogs.Logs) *Bridge {
	theb := &Bridge{
		send:    make(chan BridgeData),
		receive: make(map[string]chan BridgeData),
		closeb:  make(chan bool),
		logs:    logs,
	}
	go theb.runningBridge()
	return theb
}

// 将一个名字注册进桥中，返回桥绑定
func (b *Bridge) Register(id string) (bb *BridgeBind) {
	recerive := make(chan BridgeData)
	b.receive[id] = recerive
	return &BridgeBind{
		Send:    b.send,
		Receive: recerive,
	}
}

// 解开一个注册，这个角色将收不到消息
func (b *Bridge) UnRegister(id string) {
	delete(b.receive, id)
}

// 关闭一个桥，这个桥将再也无法被使用
func (b *Bridge) Close() {
	b.closeb <- true
	close(b.send)
	for _, v := range b.receive {
		close(v)
	}
	close(b.closeb)
}

// 实际运行的桥
func (b *Bridge) runningBridge() {
	var getdata BridgeData
	for {
		select {
		case <-b.closeb:
			return // 如果b.close有数据，则意味着关闭，这个桥，于是跳出这个方法
		case getdata = <-b.send:
			b.doSendData(b.receive, getdata)
		}
	}
}

// 执行发送队列
func (b *Bridge) doSendData(queue map[string]chan BridgeData, data BridgeData) {
	for key, br := range queue {
		if key != data.Id {
			go b.oneSendData(br, data)
		}
	}
}

// 为了防止堵塞
func (b *Bridge) oneSendData(br chan BridgeData, data BridgeData) {
	defer func() {
		if e := recover(); e != nil {
			b.logerr("bridges[Bridge]: " + fmt.Sprint(e))
		}
	}()
	br <- data
}

// 处理错误日志
func (b *Bridge) logerr(err interface{}) {
	if err == nil {
		return
	}
	if b.logs != nil {
		b.logs.ErrLog(fmt.Errorf("bridges[Bridge]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (b *Bridge) logrun(err interface{}) {
	if err == nil {
		return
	}
	if b.logs != nil {
		b.logs.RunLog(fmt.Errorf("bridges[Bridge]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 新建桥操作
func NewBridgeOperate(id string, class interface{}, logs *ilogs.Logs) (bridgeOperate *BridgeOperate) {
	bridgeOperate = &BridgeOperate{
		id:         id,
		bridgeBind: make(map[string]*BridgeBind),
		class:      reflect.ValueOf(class),
		logs:       logs,
	}
	return
}

// 绑定一个桥
func (b *BridgeOperate) Bind(bindname string, br *Bridge) {
	b.bridgeBind[bindname] = br.Register(b.id)
}

// 启动接收数据的桥，必须在全部桥绑定完成之后进行
func (b *BridgeOperate) StartBridge() {
	for key, br := range b.bridgeBind {
		go b.doRolesBridge(key, br)
	}
}

// 向为bindname的桥中发送数据
func (b *BridgeOperate) BridgeSend(bindname string, operate string, data interface{}) {

	bridgeData := BridgeData{
		Id:      b.id,
		Operate: operate,
		Data:    data,
	}
	b.bridgeBind[bindname].Send <- bridgeData
}

// doRolesBridge 一个个启动桥
func (b *BridgeOperate) doRolesBridge(key string, br *BridgeBind) {
	for {
		b_data, ok := <-br.Receive
		if ok == false {
			return
		} else {
			go b.doOneBridgeData(key, b_data)
		}
	}
}

// doOneBridgeData 实际的执行接收数据的函数。
//
// 根据接收到的结构体，转交给operate指定的处理函数。
// operate指定的函数原型为：funcname (key, id string, data interface{})。
// 其中key为发送方绑定的桥名，id为发送方的id，data为数据体，在设计函数时interface{}可以是你所确定的任何具体类型。
func (b *BridgeOperate) doOneBridgeData(key string, data BridgeData) {
	defer func() {
		if e := recover(); e != nil {
			b.logerr(e)
		}
	}()
	operate := data.Operate
	id := data.Id
	databody := data.Data

	params := make([]reflect.Value, 3)
	params[0] = reflect.ValueOf(key)
	params[1] = reflect.ValueOf(id)
	params[2] = reflect.ValueOf(databody)
	b.class.MethodByName(operate).Call(params)
}

// 处理错误日志
func (b *BridgeOperate) logerr(err interface{}) {
	if err == nil {
		return
	}
	if b.logs != nil {
		b.logs.ErrLog(fmt.Errorf("bridges[BridgeOperate]: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (b *BridgeOperate) logrun(err interface{}) {
	if err == nil {
		return
	}
	if b.logs != nil {
		b.logs.RunLog(fmt.Errorf("bridges[BridgeOperate]: %v", err))
	} else {
		fmt.Println(err)
	}
}
