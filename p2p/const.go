// Copyright 2016-2018
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package p2p

const (
	BROKEN_NODE_COUNT = 10 // The number of consecutive failed attempts, and the node will be delete from nodes table.
)

type P2Poperate uint

const (
	P2P_OPERATE_NOTHING          P2Poperate = iota // P2P Operate do nothing.
	P2P_OPERATE_OK                                 // P2P Operate status ok.
	P2P_OPERATE_ERR                                // P2P Operate status err.
	P2P_OPERATE_NODES_DISCOVER                     // P2P Operate discover nodes.
	P2P_OPERATE_MESSAGE_REQUEST                    // P2P Operate send one message to nodes.
	P2P_OPERATE_MESSAGE_RESPONSE                   // P2P Operate message response.
)
