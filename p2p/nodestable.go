// Copyright 2016-2018
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package p2p

// The node type
type NodeType uint

const (
	NODE_TYPE_NORMAL NodeType = iota // The node type is normal.
	NODE_TYPE_SERVER                 // The node is a server.
	NODE_TYPE_DEAL                   // The node is only for business deal, it will not be remembered in p2p net work node table.
)

// The node's status.
type NodeStatus struct {
	Ip   string   // The IP
	Port string   // The port
	Hash string   // The node hash.
	Type NodeType // The node type
}

// The nodes table. The string is node's hash.
type NodesTable map[string]NodeStatus
