// Copyright 2016-2018
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package p2p

import (
	"bytes"
	"fmt"

	"github.com/cznic/zappy"
	"github.com/idcsource/insight00-lib/iendecode"
	"github.com/idcsource/insight00-lib/nst2"
)

// The message operator interface.
type P2PMessageOperator interface {
	TableSet(table *NodesTable)
	OperateRequest(data *bytes.Buffer, ce *nst2.ConnExec) (err error)
	OperateResponse(data *bytes.Buffer, ce *nst2.ConnExec) (err error)
}

// Rescipient the node's ask an message.
type Rescipient struct {
	NodesTable *NodesTable        // The nodes table.
	MessageOp  P2PMessageOperator // The P2P message operator.
}

// Set the node table to P2P message operator
func (o *Rescipient) TableSet() {
	o.MessageOp.TableSet(o.NodesTable)
}

// The nst2 ConnExecer interface.
func (o *Rescipient) NSTexec(ce *nst2.ConnExec) (ss nst2.SendStat, err error) {
	alldata, err := ce.GetData()
	if err != nil {
		return
	}

	// Get the P2P_OPERATE_*
	buf := bytes.NewBuffer(alldata)
	theop := P2Poperate(iendecode.BytesToUint(buf.Next(8)))

	switch theop {
	case P2P_OPERATE_NODES_DISCOVER: // Nodes discover.
		err = o.sendNodesTable(buf, ce)
	case P2P_OPERATE_MESSAGE_REQUEST: // Get the message which other node send request.
		err = o.MessageOp.OperateRequest(buf, ce)
	case P2P_OPERATE_MESSAGE_RESPONSE: // Get the message which other node response you.
		err = o.MessageOp.OperateResponse(buf, ce)
	default:
		err = fmt.Errorf("The P2P operate not exist.")
	}

	return
}

// Send self node table to other node who need it.
func (o *Rescipient) sendNodesTable(buf *bytes.Buffer, ce *nst2.ConnExec) (err error) {
	remoteip := ce.Transmission.RemoteAddr()
	remoteip_str := remoteip.String()

	b := buf.Bytes()
	remoteinfo := SelfNodesInfo{}
	err = remoteinfo.UnmarshalBinary(b)
	if err != nil {
		return
	}
	// Add the remote to self table.
	if remoteinfo.Type == NODE_TYPE_NORMAL || remoteinfo.Type == NODE_TYPE_SERVER {
		o.NodesTable.AddOneNode(remoteinfo.Hash, remoteip_str, remoteinfo.Port, remoteinfo.Type)
	}

	// Send the table.
	theresult := NDiscoverResult{
		Status:     P2P_OPERATE_OK,
		NodesTable: o.NodesTable.OutputTable(),
	}
	theresult_b, err := theresult.MarshalBinary()
	if err != nil {
		return
	}
	// zip the result.
	theresult_b_zip, err := zappy.Encode(nil, theresult_b)
	if err != nil {
		return
	}
	// send the result.
	err = ce.SendData(theresult_b_zip)

	return
}
