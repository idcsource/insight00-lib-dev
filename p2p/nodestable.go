// Copyright 2016-2018
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package p2p

import (
	"fmt"
	"strings"

	"github.com/lukechampine/randmap"
)

// The node type
type NodeType uint

const (
	NODE_TYPE_NORMAL NodeType = iota // The node type is normal.
	NODE_TYPE_SERVER                 // The node is a server.
	NODE_TYPE_DEAL                   // The node is only for business deal, it will not be remembered in p2p net work node table.
)

// The node's status.
type NodeStatus struct {
	Hash    string   // The node hash.
	Ip      string   // The IP
	Port    string   // The port
	Type    NodeType // The node type
	Failure uint     // The failure count.
}

// The nodes table. The string is node's hash.
type NodesTable struct {
	Table map[string]*NodeStatus
	File  string // the local nodes table file.
}

func NewNodesTable() *NodesTable {
	return &NodesTable{Table: make(map[string]*NodeStatus)}
}

// Set the table file, if the file can not create or access, return error.
func (n *NodesTable) SetTableFile(filename string) (err error) {
	return
}

// Return the number of nodes in the table.
func (n *NodesTable) ReturnCount() (count int) {
	return len(n.Table)
}

// Add the node status information to the table.
//
// Ths s is a nodes table file. The s' every line is like this: hash,ip,port,type.
func (n *NodesTable) AddToTable(s string) {
	thesplit := strings.Split(s, "\n")
	for line := range thesplit {
		thelinea := strings.Split(thesplit[line], ",")
		if len(thelinea) != 4 {
			break
		}
		if _, ok := n.Table[thelinea[0]]; ok == true {
			continue
		}
		thestatus := &NodeStatus{
			Hash:    thelinea[0],
			Ip:      thelinea[1],
			Port:    thelinea[2],
			Failure: 0,
		}
		switch thelinea[3] {
		case "NODE_TYPE_NORMAL":
			thestatus.Type = NODE_TYPE_NORMAL
		case "NODE_TYPE_SERVER":
			thestatus.Type = NODE_TYPE_SERVER
		}
		n.Table[thelinea[0]] = thestatus
	}
}

// Add one node status to table.
func (n *NodesTable) AddOneNode(hash, ip, port string, thetype NodeType) {
	n.Table[hash] = &NodeStatus{
		Hash:    hash,
		Ip:      ip,
		Port:    port,
		Type:    thetype,
		Failure: 0,
	}
}

// Return all table nodes status to a string.
//
// It like:
// hash1111111111111,192.168.1.200,34,NODE_TYPE_SERVER\nhash12221221111111,192.133.1.200,34,NODE_TYPE_NORMAL
func (n *NodesTable) OutputTable() (s string) {
	for key := range n.Table {
		s += n.Table[key].Hash + "," + n.Table[key].Ip + "," + n.Table[key].Port + ","
		switch n.Table[key].Type {
		case NODE_TYPE_NORMAL:
			s += "NODE_TYPE_NORMAL"
		case NODE_TYPE_SERVER:
			s += "NODE_TYPE_SERVER"
		}
		s += "\n"
	}
	return
}

// Random return some nodes status.
func (n *NodesTable) Random(c int) (o map[string]*NodeStatus, err error) {
	lenc := len(n.Table)
	if lenc < c {
		err = fmt.Errorf("The c too big.")
		return
	}
	if lenc == c {
		o = n.Table
		return
	}
	o = make(map[string]*NodeStatus)
	for {
		k := randmap.Key(n.Table).(string)
		o[k] = n.Table[k]
		if len(o) == c {
			break
		}
	}

	return
}

// Delete node which was not connect (The Failure >= some number).
func (n *NodesTable) Delete(hash string) {
	if _, ok := n.Table[hash]; ok == true {
		delete(n.Table, hash)
	}
}
