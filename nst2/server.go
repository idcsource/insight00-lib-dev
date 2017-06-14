// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package nst2

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// A tcp server
type Server struct {
	execer     ConnExecer       // The Server's connect execution object
	logs       *ilogs.Logs      // the log
	port       string           // listen port
	tls        bool             // if use tls encryption
	tls_config *tls.Config      // the tls encryption config
	listen     *net.TCPListener // the tcp listen
	closed     bool             // if closed
}

// Create a new server for tcp
func NewServer(execer ConnExecer, port string, logs *ilogs.Logs) (s *Server, err error) {
	s = &Server{
		execer: execer,
		logs:   logs,
		port:   port,
		tls:    false,
		closed: true,
	}
	theport := ":" + s.port
	ipadrr, err := net.ResolveTCPAddr("tcp", theport)
	if err != nil {
		err = fmt.Errorf("nst2[Server]NewServer: %v", err)
		return nil, err
	}
	listens, err := net.ListenTCP("tcp", ipadrr)
	if err != nil {
		err = fmt.Errorf("nst2[Server]NewServer: %v", err)
		return nil, err
	}
	s.listen = listens
	return
}

// Let the Server to tls encryption
func (s *Server) ToTLS(pem, key string) (err error) {
	pem = pubfunc.LocalFile(pem)
	key = pubfunc.LocalFile(key)
	cert, err := tls.LoadX509KeyPair(pem, key)
	if err != nil {
		err = fmt.Errorf("nst2[Server]ToTLS: %v", err)
		return
	}
	s.tls_config = &tls.Config{Certificates: []tls.Certificate{cert}}
	s.tls = true
	return
}

func (s *Server) Start() (err error) {
	s.closed == false
	for {
		// check if closed
		if s.closed == true {
			return
		}
		connecter, err := s.listen.AcceptTCP()
		if err != nil {
			s.logs.ErrLog("nst2[Server]listen: ", err)
			continue
		}
		go s.doConnect(connecter)
	}
	return
}

func (s *Server) Close() {
	s.closed == true
}

func (s *Server) doConnect(conn *net.TCPConn) {
	var trans *Transmission
	if s.tls == false {
		trans = NewTransmission(conn)
	} else {
		trans = NewTransmissionTLS(tls.Server(conn, ts.tls_config))
	}
	conn_exec := NewConnExec(trans)

	// check if long connect or short connect.

}
