// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package nst

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/idcsource/Insight-0-0-lib/bridges"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// TCP的客户端
type TcpClient struct {
	runtimeid   string              // 运行时UNID
	bridge      *bridges.Bridge     // 通讯桥
	bridgeb     *bridges.BridgeBind // 绑定的通讯桥
	addr        string              // 地址
	logs        *ilogs.Logs         // 日志
	ccount      int                 // 启动几个连接
	tcpc        []*tcpC             // 连接管理池
	alloc_count int                 // 连接分配计数
	lock        *sync.RWMutex       // 连接分配计数的锁
	tls         bool                // 是否tls加密
}

// 进程的数据队列
type ProgressData struct {
	tcpc *tcpC
	logs *ilogs.Logs
}

type tcpC struct {
	id int
	// 地址，为重新连接准备
	addr string
	// TCP的连接返回
	tcp *TCP
	// 读写锁
	lock *sync.RWMutex
	// 自发读写锁
	slock chan bool
}

// 建立一个加密的TCP连接
func TcpClientTLS(ts *TcpClient) (err error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	for i := range ts.tcpc {
		ts.tcpc[i].tcp.tls = true
		ts.tcpc[i].tcp.tcp_tls = tls.Client(ts.tcpc[i].tcp.tcp, conf)
	}
	ts.tls = true
	return
}

// 建立一个TCP的客户端，并与addr的地址建立连接
func NewTcpClient(addr string, count int, logs *ilogs.Logs) (tc *TcpClient, err error) {
	tc = &TcpClient{
		runtimeid:   random.Unid(1, "TcpClient"),
		bridge:      bridges.NewBridge(logs),
		addr:        addr,
		logs:        logs,
		ccount:      count,
		tcpc:        make([]*tcpC, count),
		alloc_count: 0,
		lock:        new(sync.RWMutex),
		tls:         false,
	}
	tc.setBridgeBind()
	ipAdrr, err := net.ResolveTCPAddr("tcp", tc.addr)
	if err != nil {
		return
	}
	for i := 0; i < tc.ccount; i++ {
		connecter, err2 := net.DialTCP("tcp", nil, ipAdrr)
		if err2 != nil {
			return nil, err2
		}
		//err2 = connecter.SetKeepAlive(true);
		//if err2 != nil { return nil, err2; }
		tc.tcpc[i] = &tcpC{
			id:    i,
			addr:  addr,
			tcp:   NewTCP(connecter),
			lock:  new(sync.RWMutex),
			slock: make(chan bool, 1),
		}
	}
	go tc.checkConnRe()
	return
}

// 检查连接池的每个连接的状态，每30秒一次
func (tc *TcpClient) checkConnRe() {
	for {
		time.Sleep(30 * time.Second)
		for i := 0; i < tc.ccount; i++ {
			tc.checkOneConn(i)
		}
	}
}

// 检查某个连接的状态，发送心跳包，如果有问题就重新连接
func (tc *TcpClient) checkOneConn(cnum int) {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	select {
	case tc.tcpc[cnum].slock <- true:
		err := tc.tcpc[cnum].tcp.SendStat(HEART_BEAT)
		if err != nil {
			ipAdrr, _ := net.ResolveTCPAddr("tcp", tc.addr)
			connecter, err := net.DialTCP("tcp", nil, ipAdrr)
			if err != nil {
				tc.logerr(fmt.Errorf("nst[TcpClient]checkOneConn: Can't reconnect the server: %v", err))
			} else {
				tc.tcpc[cnum].tcp = NewTCP(connecter)
			}
		}
		<-tc.tcpc[cnum].slock
		tc.logrun(fmt.Errorf("心跳分配了一个连接： %v", cnum))
	default:
		tc.logrun(fmt.Errorf("心跳跳过了一个分配: %v", cnum))
	}
}

// 检查某个连接的状态，发送NORMAL_DATA，如果有问题就重新连接，再发送NORMAL_DATA
func (tc *TcpClient) checkOneConn2(cnum int) (err error) {
	err = tc.tcpc[cnum].tcp.SendStat(NORMAL_DATA)
	if err != nil {
		ipAdrr, _ := net.ResolveTCPAddr("tcp", tc.addr)
		connecter, err := net.DialTCP("tcp", nil, ipAdrr)
		if err != nil {
			tc.logerr(fmt.Errorf("nst[TcpClient]checkOneConn2: Can't reconnect the server: %v", err))
			return err
		} else {
			tc.tcpc[cnum].tcp = NewTCP(connecter)
			err = tc.tcpc[cnum].tcp.SendStat(NORMAL_DATA)
			if err != nil {
				tc.logerr(fmt.Errorf("nst[TcpClient]checkOneConn2: Can't reconnect the server: %v", err))
				return err
			}
		}
	}
	return nil
}

// 将自己绑定到自己创建的桥中
func (tc *TcpClient) setBridgeBind() {
	tc.bridgeb = tc.bridge.Register(tc.runtimeid)
}

// 返回自身提供的桥
func (tc *TcpClient) ReturnBridge() *bridges.Bridge {
	return tc.bridge
}

// 建立进程，将会固定在一个连接上进行
func (tc *TcpClient) OpenProgress() *ProgressData {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	var cnum int
	for {
		select {
		case tc.tcpc[tc.alloc_count].slock <- true:
			cnum = tc.alloc_count
			tc.alloc_count++
			if tc.alloc_count >= len(tc.tcpc) {
				tc.alloc_count = 0
			}
			err := tc.checkOneConn2(cnum)
			if err != nil {
				<-tc.tcpc[cnum].slock
				continue
			} else {
				tc.logrun(fmt.Errorf("分配了一个连接： %v", cnum))
				break
			}
		default:
			tc.alloc_count++
			if tc.alloc_count >= len(tc.tcpc) {
				tc.alloc_count = 0
			}
			tc.logrun("这个吗？跳过了一个分配")
		}
		break
	}
	return &ProgressData{
		tcpc: tc.tcpc[cnum],
		logs: tc.logs,
	}
}

// Send 向服务器端发送一个数据流。会创建连接进程，并首先发送DATA_GOON。
// 此方法将服务器的返回数据构造成一个指向TcpReturn方法的bridges.BridgeData发送给注册的桥。
// TcpReturn方法的原型为TcpReturn (key, id string, data []byte)。
// 此方法采用加锁的机制，防止在沟通的时候，被别的进程抢入。
func (tc *TcpClient) Send(data []byte) (err error) {
	if tc.tcpc == nil {
		err = errors.New("nst[TcpClient]Send: Connect not exiest.")
		return
	}
	onec := tc.OpenProgress()
	defer onec.Close()
	/*err = onec.checkOneConnInSend();
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]Send: %v", err);
		return ;
	}*/
	err = onec.tcpc.tcp.SendStat(DATA_GOON)
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]Send: %v", err)
		return
	}
	err = onec.tcpc.tcp.SendData(data)
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]Send: %v", err)
		return
	}
	// 接受是不是DATA_GOON或DATA_CLOSE
	var stat uint8
	stat, err = onec.tcpc.tcp.GetStat()
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]Send: %v", err)
		return
	}
	if stat == DATA_CLOSE {
		err = fmt.Errorf("DATA_CLOSE")
		return
	}

	var redata []byte
	redata, err = onec.tcpc.tcp.GetData()
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]Send: %v", err)
		return
	}
	bd := bridges.BridgeData{Id: random.Unid(1, time.Now().String()), Operate: "TcpReturn", Data: redata}
	tc.bridgeb.Send <- bd
	return
}

// 发送一段数据并返回服务端的数据，而不是构造桥，会创建连接进程，并首先发送DATA_GOON。
func (tc *TcpClient) SendAndReturn(data []byte) (returndata []byte, err error) {
	if tc.tcpc == nil {
		err = errors.New("nst[TcpClient]SendAndReturn: Connect not exiest.")
		return
	}
	onec := tc.OpenProgress()
	defer onec.Close()

	/*err = onec.checkOneConnInSend();
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]SendAndReturn: %v", err);
		return ;
	}*/
	err = onec.tcpc.tcp.SendStat(DATA_GOON)
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]SendAndReturn: %v", err)
		return
	}

	err = onec.tcpc.tcp.SendData(data)
	if err != nil {
		err = fmt.Errorf("nst[TcpClient]SendAndReturn: %v", err)
		return
	}

	// 接受是不是DATA_GOON或DATA_CLOSE
	var stat uint8
	stat, err = onec.tcpc.tcp.GetStat()
	if err != nil {
		err = fmt.Errorf("nst: [TcpClient]SendAndReturn: %v", err)
		return
	}
	if stat == DATA_CLOSE {
		err = fmt.Errorf("DATA_CLOSE")
		return
	}

	returndata, err = onec.tcpc.tcp.GetData()
	if err != nil {
		err = fmt.Errorf("nst[TcpClient]SendAndReturn: %v", err)
		return
	}
	return
}

// 关闭连接
func (tc *TcpClient) Close() (err error) {
	for i := 0; i < tc.ccount; i++ {
		err = tc.tcpc[i].tcp.Close()
		tc.tcpc[i] = nil
	}
	return
}

// 连接分配
/*
func (tc *TcpClient) connAlloc () (num int) {
	tc.lock.Lock();
	defer tc.lock.Unlock();
	for {
		select {
			case tc.tcpc[tc.alloc_count].slock <- true :
				num = tc.alloc_count;
				err = p.checkOneConnInSend();
				tc.alloc_count++;
				if tc.alloc_count >= len(tc.tcpc){
					tc.alloc_count = 0;
				}
				tc.logs.RunLog("分配了一个连接：", num);
				return num;
			default:
				tc.alloc_count++;
				if tc.alloc_count >= len(tc.tcpc){
					tc.alloc_count = 0;
				}
				tc.logs.RunLog("跳过了一个分配");
		}
	}
}
*/

// 处理错误和日志
func (tc *TcpClient) logerr(err interface{}) {
	if err == nil {
		return
	}
	if tc.logs != nil {
		tc.logs.ErrLog(fmt.Errorf("nst: TcpClient: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (tc *TcpClient) logrun(err interface{}) {
	if err == nil {
		return
	}
	if tc.logs != nil {
		tc.logs.RunLog(fmt.Errorf("nst: TcpClient: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 发送一段数据并返回服务端的数据，而不是构造桥，会创建连接进程，并首先发送DATA_GOON。
func (p *ProgressData) SendAndReturn(data []byte) (returndata []byte, err error) {
	/*err = p.checkOneConnInSend();
	if err != nil {
		err = fmt.Errorf("nst: [ProgressData]SendAndReturn: %v", err);
		return ;
	}*/
	err = p.tcpc.tcp.SendStat(DATA_GOON)
	if err != nil {
		err = fmt.Errorf("nst: [ProgressData]SendAndReturn: %v", err)
		return
	}
	err = p.tcpc.tcp.SendData(data)
	if err != nil {
		err = fmt.Errorf("nst[ProgressData]SendAndReturn: %v", err)
		return
	}

	// 接受是不是DATA_GOON或DATA_CLOSE
	var stat uint8
	stat, err = p.tcpc.tcp.GetStat()
	if err != nil {
		err = fmt.Errorf("nst: [ProgressData]SendAndReturn: %v", err)
		return
	}
	if stat == DATA_CLOSE {
		err = fmt.Errorf("DATA_CLOSE")
		return
	}

	returndata, err = p.tcpc.tcp.GetData()
	if err != nil {
		err = fmt.Errorf("nst[ProgressData]SendAndReturn: %v", err)
		return
	}
	return
}

// 运行时检查服务端连接，并且会发送NORMAL_DATA位
/*
func (p *ProgressData) checkOneConnInSend () (err error) {
	err = p.tcpc.tcp.SendStat(NORMAL_DATA);
	if err != nil {
		ipAdrr, _ := net.ResolveTCPAddr("tcp", p.tcpc.addr);
		connecter, err := net.DialTCP("tcp", nil, ipAdrr);
		if err != nil {
			p.logs.ErrLog("nst[TcpClient]checkOneConnInSend: Can't reconnect the server: " , err);
			return err;
		} else {
			p.tcpc.tcp = NewTCP(connecter);
			err = p.tcpc.tcp.SendStat(NORMAL_DATA);
			return err;
		}
	} else {
		return nil;
	}
}
*/

// 关闭分配的连接进程，并发送DATA_CLOSE
func (p *ProgressData) Close() {
	fmt.Println("释放连接")
	p.logrun(fmt.Errorf("释放了一个连接？ %v", p.tcpc.id))
	p.tcpc.tcp.SendStat(DATA_CLOSE)
	<-p.tcpc.slock
	p.logrun(fmt.Errorf("释放了一个连接： %v", p.tcpc.id))
}

// 处理错误和日志
func (p *ProgressData) logerr(err interface{}) {
	if err == nil {
		return
	}
	if p.logs != nil {
		p.logs.ErrLog(fmt.Errorf("nst: TcpClient: %v", err))
	} else {
		fmt.Println(err)
	}
}

// 处理运行日志
func (p *ProgressData) logrun(err interface{}) {
	if err == nil {
		return
	}
	if p.logs != nil {
		p.logs.RunLog(fmt.Errorf("nst: TcpClient: %v", err))
	} else {
		fmt.Println(err)
	}
}
