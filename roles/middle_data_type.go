// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

package roles

import (
	"bytes"

	"github.com/idcsource/insight00-lib/iendecode"
)

// 角色的版本，包含版本号和Id
type RoleVersion struct {
	Version uint32
	Id      string
}

func (r *RoleVersion) EncodeBinary() (b []byte, lens int64, err error) {
	version_b := iendecode.Uint32ToBytes(r.Version)
	id_b := []byte(r.Id)
	//b = append(version_b, id_b...)
	lens = int64(8 + len(id_b))
	b = make([]byte, lens)
	copy(b, version_b)
	copy(b[8:], id_b)
	return
}

func (r *RoleVersion) DecodeBinary(b []byte) (err error) {
	version_b := b[0:8]
	r.Version = iendecode.BytesToUint32(version_b)
	id_b := b[8:]
	r.Id = string(id_b)
	return
}

// 角色的关系，包含了父、子、朋友、上下文
type RoleRelation struct {
	Father   string             // 父角色（拓扑结构层面）
	Children []string           // 虚拟的子角色群，只保存键名
	Friends  map[string]Status  // 虚拟的朋友角色群，只保存键名，其余与朋友角色群一致
	Contexts map[string]Context // 上下文
}

func (r RoleRelation) EncodeBinary() (b []byte, lens int64, err error) {
	buf := bytes.Buffer{}
	// bytes the Father: the string length + string
	father_b := []byte(r.Father)
	father_b_len := int64(len(father_b))
	buf.Write(iendecode.Int64ToBytes(father_b_len))
	buf.Write(father_b)
	lens += 8 + father_b_len
	// bytes the Children: the children number + string length + string
	chilren_num := int64(len(r.Children))
	buf.Write(iendecode.Int64ToBytes(chilren_num))
	lens += 8
	for i := range r.Children {
		child_b := []byte(r.Children[i])
		child_b_len := int64(len(child_b))
		buf.Write(iendecode.Int64ToBytes(child_b_len))
		buf.Write(child_b)
		lens += 8 + child_b_len
	}
	// bytes Friends: the Friends number + key length + key + value length + value
	friends_num := int64(len(r.Friends))
	buf.Write(iendecode.Int64ToBytes(friends_num))
	lens += 8
	for key, _ := range r.Friends {
		// the key
		key_b := []byte(key)
		key_b_len := int64(len(key_b))
		buf.Write(iendecode.Int64ToBytes(key_b_len))
		buf.Write(key_b)
		lens += 8 + key_b_len
		// the value
		value_b, value_lens, err := r.Friends[key].EncodeBinary()
		if err != nil {
			return nil, 0, err
		}
		buf.Write(iendecode.Int64ToBytes(value_lens))
		buf.Write(value_b)
		lens += 8 + value_lens
	}
	// bytes Contexts: the Contexts number + key length + key + value length + value
	contexts_num := int64(len(r.Contexts))
	buf.Write(iendecode.Int64ToBytes(contexts_num))
	lens += 8
	for key, _ := range r.Contexts {
		// the key
		key_b := []byte(key)
		key_b_len := int64(len(key_b))
		buf.Write(iendecode.Int64ToBytes(key_b_len))
		buf.Write(key_b)
		lens += 8 + key_b_len
		// the value
		value_b, value_lens, err := r.Contexts[key].EncodeBinary()
		if err != nil {
			return nil, 0, err
		}
		buf.Write(iendecode.Int64ToBytes(value_lens))
		buf.Write(value_b)
		lens += 8 + value_lens
	}

	b = buf.Bytes()
	return
}

func (r *RoleRelation) DecodeBinary(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(b)
	// de bytes the Father: the string length + string
	father_len := iendecode.BytesToInt64(buf.Next(8))
	r.Father = string(buf.Next(int(father_len)))
	// de bytes the Children: the children number + string length + string
	children_num := iendecode.BytesToInt64(buf.Next(8))
	r.Children = make([]string, children_num)
	var i int64
	for i = 0; i < children_num; i++ {
		child_len := iendecode.BytesToInt64(buf.Next(8))
		r.Children[i] = string(buf.Next(int(child_len)))
	}
	// bytes Friends: the Friends number + key length + key + value length + value
	r.Friends = make(map[string]Status)
	friends_num := iendecode.BytesToInt64(buf.Next(8))
	for i = 0; i < friends_num; i++ {
		key_len := iendecode.BytesToInt64(buf.Next(8))
		key := string(buf.Next(int(key_len)))
		value_len := iendecode.BytesToInt64(buf.Next(8))
		value_b := buf.Next(int(value_len))
		value := Status{}
		err = value.DecodeBinary(value_b)
		if err != nil {
			return
		}
		r.Friends[key] = value
	}
	// bytes Contexts: the Contexts number + key length + key + value length + value
	r.Contexts = make(map[string]Context)
	contexts_num := iendecode.BytesToInt64(buf.Next(8))
	for i = 0; i < contexts_num; i++ {
		key_len := iendecode.BytesToInt64(buf.Next(8))
		key := string(buf.Next(int(key_len)))
		value_len := iendecode.BytesToInt64(buf.Next(8))
		value_b := buf.Next(int(value_len))
		value := Context{}
		err = value.DecodeBinary(value_b)
		if err != nil {
			return
		}
		r.Contexts[key] = value
	}
	return
}

type RoleData struct {
	Point map[string]*RoleDataPoint
}

type RoleDataPoint struct {
	Type string
	Data interface{}
}

// 角色的中期存储类型
type RoleMiddleData struct {
	Version        RoleVersion
	VersionChange  bool
	Relation       RoleRelation
	RelationChange bool
	Data           RoleData
	DataChange     bool
}
