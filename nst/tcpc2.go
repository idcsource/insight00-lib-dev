// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package nst

import (
	"crypto/tls"
	"fmt"
	"net"
)

type TcpClient2 struct {
	addr         string     // 地址
	iftls        bool       // if use tls
	max_count    int        // the max connect count
	connect_pool []*Connect // the connect pool
}

type Connect struct {
	tcp         *TCP   // the tcp data transmission
	addr        string // address
	iftls       bool   // if use tls
	lconnection bool   // if use long connection
	paused      bool   // if it paused
	closed      bool   // if is closed
}

func NewTcpClient2(addr string, max_count int, iftls bool) *TcpClient2 {
	return &TcpClient2{
		iftls:        iftls,
		addr:         addr,
		max_count:    max_count,
		connect_pool: make([]*Connect, 0),
	}
}

func (t *TcpClient2) OpenProgress() (c *Connect, err error) {
	return t.OpenConnect()
}

func (t *TcpClient2) OpenConnect() (c *Connect, err error) {
	c = &Connect{
		addr:        t.addr,
		iftls:       t.iftls,
		lconnection: false,
		paused:      true,
		closed:      false,
	}
	err = c.connect()
	return
}

func (c *Connect) connect() (err error) {
	var ipadrr *net.TCPAddr
	ipadrr, err = net.ResolveTCPAddr("tcp", c.addr)
	if err != nil {
		return
	}
	var connecter *net.TCPConn
	connecter, err = net.DialTCP("tcp", nil, ipadrr)
	if err != nil {
		return
	}
	if c.iftls == true {
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		c.tcp = NewTCPtls(tls.Client(connecter, conf))
	} else {
		c.tcp = NewTCP(connecter)
	}
	return
}

func (c *Connect) SendAndReturn(data []byte) (returndata []byte, err error) {
	if c.paused == true {
		err = c.tcp.SendStat(NORMAL_DATA)
		if err != nil {
			err = fmt.Errorf("nst: [Connect]SendAndReturn: %v", err)
			return
		}
		c.paused = false
	}

	err = c.tcp.SendStat(DATA_GOON)
	if err != nil {
		err = fmt.Errorf("nst: [Connect]SendAndReturn: %v", err)
		return
	}

	err = c.tcp.SendData(data)
	if err != nil {
		err = fmt.Errorf("nst[Connect]SendAndReturn: %v", err)
		return
	}

	// 接受是不是DATA_GOON或DATA_CLOSE
	var stat uint8
	stat, err = c.tcp.GetStat()
	if err != nil {
		err = fmt.Errorf("nst: [Connect]SendAndReturn: %v", err)
		return
	}
	if stat == DATA_CLOSE {
		err = fmt.Errorf("DATA_CLOSE")
		return
	}

	returndata, err = c.tcp.GetData()
	if err != nil {
		err = fmt.Errorf("nst[Connect]SendAndReturn: %v", err)
		return
	}
	return
}

func (c *Connect) Pause() (err error) {
	err = c.tcp.SendStat(DATA_CLOSE)
	if err != nil {
		return
	}
	c.paused = true
	return
}

func (c *Connect) Close() (err error) {
	defer func() {
		//c.tcp.Close()
		c.closed = true
		c = nil
	}()
	err = c.tcp.SendStat(DATA_CLOSE)
	if err != nil {
		return
	}
	err = c.tcp.SendStat(CONN_CLOSE)
	if err != nil {
		return err
	}
	_, err = c.tcp.GetStat()
	if err != nil {
		return err
	}
	return
}
