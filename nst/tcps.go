// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package nst

import (
	"fmt"
	"net"
	"reflect"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 这是一个使用Tcp协议的服务器端监听组件。
// 根据设置接收tcp套接字传送来的信息并转交给注册的接收者。
type TcpServer struct {
	role reflect.Value // 将收到的数据返回给谁处理
	logs *ilogs.Logs   // 运行日志
	port string        // 监听的端口
}

// 新建一个Tcp的监听。注册一个符合ConnExecer接口的执行者负责真正的处理接口。
func NewTcpServer(role ConnExecer, port string, logs *ilogs.Logs) *TcpServer {
	ts := &TcpServer{role: reflect.ValueOf(role), logs: logs, port: port}
	go ts.startServer()
	return ts
}

// 启动服务器，在NewTcpServer中直接执行
func (ts *TcpServer) startServer() {
	theport := ":" + ts.port
	ipAdrr, err1 := net.ResolveTCPAddr("tcp", theport)
	if err1 != nil {
		ts.logerr(fmt.Errorf("StartServer: theport: %v", err1))
		return
	}
	listens, err2 := net.ListenTCP("tcp", ipAdrr)
	if err2 != nil {
		ts.logerr(fmt.Errorf("StartServer: ipAdrr: %v", err2))
		return
	}
	for {
		Connecter, err := listens.AcceptTCP()
		if err != nil {
			ts.logerr(fmt.Errorf("StartServer: AcceptTCP: %v", err))
			continue
		}
		go ts.doConn(Connecter)
	}
}

// 执行一个连接
func (ts *TcpServer) doConn(conn *net.TCPConn) {
	defer func() {
		if e := recover(); e != nil {
			ts.logerr(fmt.Errorf("nst[TcpServer]doConn: ", e))
		}
	}()
	tcp := NewTCP(conn)
	for {
		stat, err := tcp.GetStat()
		if err != nil {
			ts.logerr(fmt.Errorf("nst[TcpServer]doConn: %v", err))
			return
		}
		if stat == NORMAL_DATA {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(tcp)
			// 注册的方法需要符合ConnExecer，整个连接将交给注册的方法去执行
			rr := ts.role.MethodByName("ExecTCP").Call(in)
			erri := rr[0].Interface()
			var err error
			if erri != nil {
				err = erri.(error)
			}
			if fmt.Sprint(err) == "EOF" {
				tcp.Close()
				ts.logs.RunLog("nst[TcpServer]doConn: Connect Closed.")
				return
			}
		}
	}
}

// 做接收的信息处理，处理掉头部信息
func (ts *TcpServer) GetData(tcp *TCP) (data []byte, err error) {
	stat, err := tcp.GetStat()
	if err != nil {
		err = fmt.Errorf("nst[TcpServer]GetData: %v", err)
		ts.logerr(err)
		return nil, err
	}
	if stat == DATA_GOON {
		return nil, fmt.Errorf("DATA_CLOSE")
	}
	data, err = tcp.GetData()
	if err != nil {
		err = fmt.Errorf("nst[TcpServer]GetData: %v", err)
		ts.logerr(err)
	}
	return data, err
}

// 做发送的信息处理，加上发送头
func (ts *TcpServer) SendData(tcp *TCP, data []byte) (err error) {
	err = tcp.SendStat(DATA_GOON)
	if err != nil {
		err = fmt.Errorf("nst[TcpServer]SendData: %v", err)
		ts.logerr(err)
		return err
	}
	err = tcp.SendData(data)
	if err != nil {
		err = fmt.Errorf("nst[TcpServer]SendData: %v", err)
		ts.logerr(err)
		return err
	}
	return nil
}

// 发送关闭连接的处理
func (ts *TcpServer) SendClose(tcp *TCP) (err error) {
	err = tcp.SendStat(DATA_CLOSE)
	return err
}

// 处理错误和日志
func (ts *TcpServer) logerr(err interface{}) {
	if err == nil {
		return
	}
	if ts.logs != nil {
		ts.logs.ErrLog(fmt.Errorf("nst: TcpServer: %v", err))
	} else {
		fmt.Println(err)
	}
}
