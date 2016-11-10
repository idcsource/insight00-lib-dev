// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// Bridges 数据桥。负责在各个子进程（角色）间传递数据。
package bridges

import(
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 桥结构
type Bridge struct {
	send		chan BridgeData											// 发送方
	receive		map[string] chan BridgeData								// 接收方注册，string为接收方Id
	closeb		chan bool		 										// 关闭检测
	logs		*ilogs.Logs												// 日志
}

// 桥的传递数据的结构
type BridgeData struct {
	Id			string													// 判断是谁发送的信息
	Operate		string													// 供反射使用，确定操作那个类型的志或函数
	Data		interface{}												// 实际的数据主体
}

// 桥绑定
type BridgeBind struct {
	Send		chan<- BridgeData
	Receive		<-chan BridgeData	
}

// 创建一个桥
func NewBridge (logs *ilogs.Logs) *Bridge {
	theb := &Bridge{send : make(chan BridgeData), receive : make(map[string] chan BridgeData), closeb : make(chan bool), logs : logs};
	go theb.runningBridge();
	return theb;
}

// 将一个名字注册进桥中，返回桥绑定
func (b *Bridge) Register (id string) (bb *BridgeBind) {
	recerive := make(chan BridgeData);
	b.receive[id] = recerive;
	return &BridgeBind{Send : b.send, Receive : recerive};
}

// 解开一个注册，这个角色将收不到消息
func (b *Bridge) UnRegister (id string) {
	delete(b.receive, id);
}

// 关闭一个桥，这个桥将再也无法被使用
func (b *Bridge) Close() {
	b.closeb <- true;
	close(b.send);
	for _, v := range b.receive {
		close(v);
	}
	close(b.closeb);
}

// 实际运行的桥
func (b *Bridge) runningBridge() {
	var getdata BridgeData;
	for {
		select {
			case <-b.closeb :
				return;			// 如果b.close有数据，则意味着关闭，这个桥，于是跳出这个方法
			case getdata = <-b.send :
				b.doSendData(b.receive, getdata);
		}
	}
}

// 执行发送队列
func (b *Bridge) doSendData (queue map[string] chan BridgeData, data BridgeData){
	for key, br := range queue {
		if key != data.Id {
			go b.oneSendData(br, data);
		}
	}
}

// 为了防止堵塞
func (b *Bridge) oneSendData (br chan BridgeData, data BridgeData){
	defer func(){
		if e := recover(); e != nil {
			b.logs.ErrLog("bridges: ", e);
		}
	}()
	br <- data;
}
