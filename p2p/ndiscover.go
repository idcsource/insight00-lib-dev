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

	"github.com/idcsource/insight00-lib/iendecode"
	"github.com/idcsource/insight00-lib/nst2"

	"github.com/cznic/zappy"
)

// The nodes discover result.
//
// It's must be zip by cznic/zappy. So you must unzip before use.
type NDiscoverResult struct {
	Status     P2Poperate
	Err        string
	NodesTable string
}

func (o NDiscoverResult) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// Status 8
	status_b := iendecode.UintToBytes(uint(o.Status))
	buf.Write(status_b)

	// Err
	err_b := []byte(o.Err)
	err_b_len := len(err_b)
	err_b_len_b := iendecode.IntToBytes(err_b_len)
	buf.Write(err_b_len_b)
	buf.Write(err_b)

	// NodesTable
	nodestable_b := []byte(o.NodesTable)
	nodestable_b_len := len(nodestable_b)
	nodestable_b_len_b := iendecode.IntToBytes(nodestable_b_len)
	buf.Write(nodestable_b_len_b)
	buf.Write(nodestable_b)

	data = buf.Bytes()
	return
}

func (o *NDiscoverResult) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(data)

	// Status
	status_b := buf.Next(8)
	o.Status = P2Poperate(iendecode.BytesToUint(status_b))

	// Err
	err_b_len := iendecode.BytesToInt(buf.Next(8))
	err_b := buf.Next(err_b_len)
	o.Err = string(err_b)

	// NodesTable
	nodestable_b_len := iendecode.BytesToInt(buf.Next(8))
	nodestable_b := buf.Next(nodestable_b_len)
	o.NodesTable = string(nodestable_b)

	return
}

// Discover nodes table from one node.
//
// The nt is nodes table file.
func NodesDiscover(ip, port string) (nt string, err error) {
	addr := ip + ":" + port
	client, err := nst2.NewClient(addr, 1, true)
	if err != nil {
		return
	}
	pro, err := client.OpenConnect()
	if err != nil {
		return
	}
	send_operate := iendecode.UintToBytes(uint(P2P_OPERATE_DISCOVER))
	red_zip, err := pro.SendAndReturn(send_operate) // send P2P_OPERATE_DISCOVER
	if err != nil {
		return
	}

	pro.Close()
	client.Close()

	// unzip red
	red, err := zappy.Decode(nil, red_zip)
	if err != nil {
		return
	}

	// decode red
	ndiscover_r := NDiscoverResult{}
	err = ndiscover_r.UnmarshalBinary(red)
	if err != nil {
		return
	}
	if ndiscover_r.Status != P2P_OPERATE_OK {
		err = fmt.Errorf(ndiscover_r.Err)
		return
	}
	nt = ndiscover_r.NodesTable

	return
}
