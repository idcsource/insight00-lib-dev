// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
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
func NewTcpServer(role ConnExecer, port string, logs *ilogs.Logs) (ts *TcpServer, err error) {
	ts = &TcpServer{
		role: reflect.ValueOf(role),
		logs: logs,
		port: port,
	}

	theport := ":" + ts.port
	ipAdrr, err := net.ResolveTCPAddr("tcp", theport)
	if err != nil {
		ts.logerr(fmt.Errorf("StartServer: theport: %v", err))
		return nil, err
	}
	listens, err := net.ListenTCP("tcp", ipAdrr)
	if err != nil {
		ts.logerr(fmt.Errorf("StartServer: ipAdrr: %v", err))
		return nil, err
	}

	go ts.startServer(listens)
	return ts, nil
}

// 启动服务器，在NewTcpServer中直接执行
func (ts *TcpServer) startServer(listens *net.TCPListener) {
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
	conn_exec := NewConnExec(tcp)

	for {
		stat, err := tcp.GetStat()
		if err != nil {
			ts.logerr(fmt.Errorf("nst[TcpServer]doConn: %v", err))
			return
		}
		if stat == NORMAL_DATA {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(conn_exec)
			// 注册的方法需要符合ConnExecer，整个连接将交给注册的方法去执行
			rr := ts.role.MethodByName("ExecTCP").Call(in)
			erri := rr[0].Interface()
			var err error
			if erri != nil {
				err = erri.(error)
				ts.logerr(err)
			}
			if fmt.Sprint(err) == "EOF" {
				tcp.Close()
				ts.logs.RunLog("nst[TcpServer]doConn: Connect Closed.")
				return
			}
		}
	}
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

// 服务器端的连接执行类型
type ConnExec struct {
	tcp *TCP
}

// 创建一个连接执行
func NewConnExec(tcp *TCP) (connExec *ConnExec) {
	return &ConnExec{tcp: tcp}
}

// 做接收的信息处理，处理掉头部信息
func (ce *ConnExec) GetData() (data []byte, err error) {
	stat, err := ce.tcp.GetStat()
	if err != nil {
		if fmt.Sprint(err) == "EOF" {
			return nil, err
		} else if err != nil {
			err = fmt.Errorf("nst[ConnExec]GetData: %v", err)
			return nil, err
		}
	}
	if stat != DATA_GOON {
		return nil, fmt.Errorf("DATA_CLOSE")
	}
	data, err = ce.tcp.GetData()
	if err != nil {
		if fmt.Sprint(err) == "EOF" {
			return nil, err
		} else if err != nil {
			err = fmt.Errorf("nst[ConnExec]GetData: %v", err)
			return nil, err
		}
	}
	return data, err
}

// 做发送的信息处理，加上发送头
func (ce *ConnExec) SendData(data []byte) (err error) {
	err = ce.tcp.SendStat(DATA_GOON)
	if err != nil {
		if fmt.Sprint(err) == "EOF" {
			return err
		} else if err != nil {
			err = fmt.Errorf("nst[ConnExec]SendData: %v", err)
			return err
		}
	}
	err = ce.tcp.SendData(data)
	if err != nil {
		if fmt.Sprint(err) == "EOF" {
			return err
		} else if err != nil {
			err = fmt.Errorf("nst[ConnExec]SendData: %v", err)
			return err
		}
	}
	return nil
}

// 发送关闭连接的处理
func (ce *ConnExec) SendClose() (err error) {
	err = ce.tcp.SendStat(DATA_CLOSE)
	if fmt.Sprint(err) == "EOF" {
		return err
	} else if err != nil {
		err = fmt.Errorf("nst[ConnExec]SendData: %v", err)
		return err
	}
	return nil
}

// 返回其中的TCP
func (ce *ConnExec) Tcp() (tcp *TCP) {
	return ce.tcp
}
