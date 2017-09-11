// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package spots

import (
	"bytes"
	"fmt"

	"github.com/idcsource/insight00-lib/iendecode"
)

type Spots struct {
	// ID
	Id string
	// Body
	Body DataBodyer
	// bbody
	bbody    map[string][]byte
	is_bbody bool
	// 父点（拓扑结构层面）
	father string
	// 虚拟的子点群，只保存键名
	children []string
	// 虚拟的朋友点群，只保存键名，其余与朋友点群一致
	friends map[string]Status
	// 上下文关系列表
	context map[string]Context
	// 父点被更改
	father_changed bool
	// 子点关系被改变
	children_changed bool
	// 朋友点被改变
	friends_changed bool
	// 上下文关系改变
	context_changed bool
	// 是否被删除
	be_delete bool
}

func NewSpot(id string) (spot *Spots) {
	spot = &Spots{
		Id:               id,
		bbody:            make(map[string][]byte),
		is_bbody:         false,
		children:         make([]string, 0),
		friends:          make(map[string]Status),
		context:          make(map[string]Context),
		father_changed:   false,
		children_changed: false,
		friends_changed:  false,
		context_changed:  false,
		be_delete:        false,
	}
	return
}

func NewEmptySpot() (spot *Spots) {
	spot = &Spots{
		bbody:            make(map[string][]byte),
		is_bbody:         false,
		children:         make([]string, 0),
		friends:          make(map[string]Status),
		context:          make(map[string]Context),
		father_changed:   false,
		children_changed: false,
		friends_changed:  false,
		context_changed:  false,
		be_delete:        false,
	}
	return
}

func NewSpotWithBody(id string, body DataBodyer) (spot *Spots) {
	spot = NewSpot(id)
	spot.SetBody(body)
	return
}

func (s *Spots) UseBody() {
	s.is_bbody = false
}

func (s *Spots) UseBbody() {
	s.is_bbody = true
}

func (s *Spots) GetId() (id string) {
	return s.Id
}

func (s *Spots) SetBody(body DataBodyer) {
	s.Body = body
	s.is_bbody = false
}

func (s *Spots) BtoDataBody(body DataBodyer) (err error) {
	err = body.DecodeBbody(s.bbody)
	if err != nil {
		return
	}
	s.Body = body
	s.bbody = make(map[string][]byte)
	s.is_bbody = false
	return
}

func (s *Spots) DataBodyToB() (err error) {
	if s.bbody != nil {
		s.bbody, err = s.Body.EncodeBbody()
		if err == nil {
			s.is_bbody = true
		}
		return
	}
	return
}

func (s *Spots) GetBody() (body DataBodyer) {
	return s.Body
}

func (s *Spots) GetFather() string {
	return s.father
}

func (s *Spots) ResetFather() {
	s.father = ""
	s.father_changed = true
}

func (s *Spots) SetFather(id string) {
	s.father = id
	s.father_changed = true
}

func (s *Spots) GetChildren() []string {
	return s.children
}

func (s *Spots) ResetChilren() {
	s.children = make([]string, 0)
	s.children_changed = true
}

func (s *Spots) SetChildren(children []string) {
	s.children = children
	s.children_changed = true
}

func (s *Spots) ExistChild(name string) bool {
	for _, v := range s.children {
		if v == name {
			return true
			break
		}
	}
	return false
}

func (s *Spots) AddChild(id string) {
	exist := s.ExistChild(id)
	if exist == true {
		return
	} else {
		s.children = append(s.children, id)
		s.children_changed = true
		return
	}
}

func (s *Spots) DeleteChild(child string) (err error) {
	exist := s.ExistChild(child)
	if exist != true {
		err = fmt.Errorf("The child id not exist.")
		return
	} else {
		var count int
		for i, v := range s.children {
			if v == child {
				count = i
				break
			}
		}
		s.children = append(s.children[:count], s.children[count+1:]...)
		s.children_changed = true
		return nil
	}
}

func (s *Spots) GetFriends() map[string]Status {
	return s.friends
}

func (s *Spots) ResetFriends() {
	s.friends = make(map[string]Status)
	s.friends_changed = true
}

func (s *Spots) SetFriends(friends map[string]Status) {
	s.friends = friends
	s.friends_changed = true
}

func (s *Spots) ExistFriend(name string) bool {
	_, FindV := s.friends[name]
	if FindV == true {
		return true
	}
	return false
}

func (s *Spots) AddFriend(id string) {
	// 检查这个friend是否存在
	ifexist := s.ExistFriend(id)
	if ifexist == true {
		return
	}
	s.friends[id] = NewStatus()
	s.friends_changed = true
	return
}

func (s *Spots) DeleteFriend(id string) (err error) {
	ifexist := s.ExistFriend(id)
	if ifexist == true {
		err = fmt.Errorf("The friend id not exist.")
		return
	}
	delete(s.friends, id)
	s.friends_changed = true
	return nil
}

func (s *Spots) SetFriendIntStatus(id string, bit int, value int64) (err error) {
	_, findf := s.friends[id]
	if findf == false {
		s.friends[id] = NewStatus()
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	s.friends[id].Int[bit] = value
	s.friends_changed = true
	return nil
}

func (s *Spots) SetFriendFloatStatus(id string, bit int, value float64) (err error) {
	_, findf := s.friends[id]
	if findf == false {
		s.friends[id] = NewStatus()
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	s.friends[id].Float[bit] = value
	s.friends_changed = true
	return nil
}

func (s *Spots) SetFriendComplexStatus(id string, bit int, value complex128) (err error) {
	_, findf := s.friends[id]
	if findf == false {
		s.friends[id] = NewStatus()
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	s.friends[id].Complex[bit] = value
	s.friends_changed = true
	return nil
}

func (s *Spots) SetFriendStringStatus(id string, bit int, value string) (err error) {
	_, findf := s.friends[id]
	if findf == false {
		s.friends[id] = NewStatus()
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	s.friends[id].String[bit] = value
	s.friends_changed = true
	return nil
}

func (s *Spots) GetFriendIntStatus(id string, bit int) (value int64, have bool, err error) {
	have = true
	_, findf := s.friends[id]
	if findf == false {
		have = false
		err = fmt.Errorf("The friend id not exist.")
		return
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	value = s.friends[id].Int[bit]
	return
}

func (s *Spots) GetFriendFloatStatus(id string, bit int) (value float64, have bool, err error) {
	have = true
	_, findf := s.friends[id]
	if findf == false {
		have = false
		err = fmt.Errorf("The friend id not exist.")
		return
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	value = s.friends[id].Float[bit]
	return
}

func (s *Spots) GetFriendComplexStatus(id string, bit int) (value complex128, have bool, err error) {
	have = true
	_, findf := s.friends[id]
	if findf == false {
		have = false
		err = fmt.Errorf("The friend id not exist.")
		return
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	value = s.friends[id].Complex[bit]
	return
}

func (s *Spots) GetFriendStringStatus(id string, bit int) (value string, have bool, err error) {
	have = true
	_, findf := s.friends[id]
	if findf == false {
		have = false
		err = fmt.Errorf("The friend id not exist.")
		return
	}
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		return
	}
	value = s.friends[id].String[bit]
	return
}

func (s *Spots) SetContexts(context map[string]Context) {
	s.context = context
}

// 获取全部上下文，存储实例调用
func (s *Spots) GetContexts() map[string]Context {
	return s.context
}

func (s *Spots) NewContext(contextname string) (err error) {
	_, find := s.context[contextname]
	if find == false {
		s.context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
	} else {
		err = fmt.Errorf("The context already exist.")
	}
	return
}

func (s *Spots) ExistContext(contextname string) (have bool) {
	_, have = s.context[contextname]
	return have
}

func (s *Spots) AddContextUp(contextname, upname string) {
	_, find := s.context[contextname]
	if find == false {
		s.context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		s.context[contextname].Up[upname] = NewStatus()
	}
	s.context_changed = true
}

func (s *Spots) AddContextDown(contextname, downname string) {
	_, find := s.context[contextname]
	if find == false {
		s.context[contextname] = Context{
			Up:   make(map[string]Status),
			Down: make(map[string]Status),
		}
		s.context[contextname].Down[downname] = NewStatus()
	}
	s.context_changed = true
}

func (s *Spots) DelContextUp(contextname, upname string) {
	_, find := s.context[contextname]
	if find == true {
		if _, find2 := s.context[contextname].Up[upname]; find2 == true {
			delete(s.context[contextname].Up, upname)
		}
	}
	s.context_changed = true
}

func (s *Spots) DelContextDown(contextname, downname string) {
	_, find := s.context[contextname]
	if find == true {
		if _, find2 := s.context[contextname].Down[downname]; find2 == true {
			delete(s.context[contextname].Down, downname)
		}
	}
	s.context_changed = true
}

func (s *Spots) DelContext(contextname string) {
	_, find := s.context[contextname]
	if find == true {
		delete(s.context, contextname)
	}
}

func (s *Spots) GetContext(contextname string) (context Context, have bool) {
	if context, have = s.context[contextname]; have == true {
		return
	} else {
		have = false
		return
	}
}

func (s *Spots) SetContext(contextname string, context Context) {
	s.context[contextname] = context
	s.context_changed = true
}

func (s *Spots) GetContextsName() (names []string) {
	lens := len(s.context)
	names = make([]string, lens)
	i := 0
	for name, _ := range s.context {
		names[i] = name
		i++
	}
	return names
}

func (s *Spots) SetContextIntStatus(contextname string, upordown ContextUpDown, id string, bit int, value int64) (err error) {
	if bit > 9 {
		return fmt.Errorf("The bit must less than 10.")
	}
	_, findc := s.context[contextname]
	if findc == false {
		s.context[contextname] = NewContext()
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			s.context[contextname].Up[id] = NewStatus()
		}
		s.context[contextname].Up[id].Int[bit] = value
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			s.context[contextname].Down[id] = NewStatus()
		}
		s.context[contextname].Down[id].Int[bit] = value
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

func (s *Spots) SetContextFloatStatus(contextname string, upordown ContextUpDown, id string, bit int, value float64) (err error) {
	if bit > 9 {
		return fmt.Errorf("The bit must less than 10.")
	}
	_, findc := s.context[contextname]
	if findc == false {
		s.context[contextname] = NewContext()
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			s.context[contextname].Up[id] = NewStatus()
		}
		s.context[contextname].Up[id].Float[bit] = value
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			s.context[contextname].Down[id] = NewStatus()
		}
		s.context[contextname].Down[id].Float[bit] = value
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

func (s *Spots) SetContextComplexStatus(contextname string, upordown ContextUpDown, id string, bit int, value complex128) (err error) {
	if bit > 9 {
		return fmt.Errorf("The bit must less than 10.")
	}
	_, findc := s.context[contextname]
	if findc == false {
		s.context[contextname] = NewContext()
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			s.context[contextname].Up[id] = NewStatus()
		}
		s.context[contextname].Up[id].Complex[bit] = value
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			s.context[contextname].Down[id] = NewStatus()
		}
		s.context[contextname].Down[id].Complex[bit] = value
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

func (s *Spots) SetContextStringStatus(contextname string, upordown ContextUpDown, id string, bit int, value string) (err error) {
	if bit > 9 {
		return fmt.Errorf("The bit must less than 10.")
	}
	_, findc := s.context[contextname]
	if findc == false {
		s.context[contextname] = NewContext()
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			s.context[contextname].Up[id] = NewStatus()
		}
		s.context[contextname].Up[id].String[bit] = value
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			s.context[contextname].Down[id] = NewStatus()
		}
		s.context[contextname].Down[id].String[bit] = value
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
	}
	return
}

func (s *Spots) GetContextIntStatus(contextname string, upordown ContextUpDown, id string, bit int) (value int64, have bool, err error) {
	have = true
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		have = false
		return
	}
	_, findc := s.context[contextname]
	if findc == false {
		have = false
		err = fmt.Errorf("The context name not exist.")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Up[id].Int[bit]
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Down[id].Int[bit]
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
		have = false
	}
	return
}

func (s *Spots) GetContextFloatStatus(contextname string, upordown ContextUpDown, id string, bit int) (value float64, have bool, err error) {
	have = true
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		have = false
		return
	}
	_, findc := s.context[contextname]
	if findc == false {
		have = false
		err = fmt.Errorf("The context name not exist.")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Up[id].Float[bit]
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Down[id].Float[bit]
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
		have = false
	}
	return
}

func (s *Spots) GetContextComplexStatus(contextname string, upordown ContextUpDown, id string, bit int) (value complex128, have bool, err error) {
	have = true
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		have = false
		return
	}
	_, findc := s.context[contextname]
	if findc == false {
		have = false
		err = fmt.Errorf("The context name not exist.")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Up[id].Complex[bit]
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Down[id].Complex[bit]
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
		have = false
	}
	return
}
func (s *Spots) GetContextStringStatus(contextname string, upordown ContextUpDown, id string, bit int) (value string, have bool, err error) {
	have = true
	if bit > 9 {
		err = fmt.Errorf("The bit must less than 10.")
		have = false
		return
	}
	_, findc := s.context[contextname]
	if findc == false {
		have = false
		err = fmt.Errorf("The context name not exist.")
		return
	}
	if upordown == CONTEXT_UP {
		_, findr := s.context[contextname].Up[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Up[id].String[bit]
	} else if upordown == CONTEXT_DOWN {
		_, findr := s.context[contextname].Down[id]
		if findr == false {
			have = false
			err = fmt.Errorf("The id not exist.")
			return
		}
		value = s.context[contextname].Down[id].String[bit]
	} else {
		err = fmt.Errorf("The upordown must CONTEXT_UP or CONTEXT_DOWN.")
		have = false
	}
	return
}

func (s *Spots) SetBbody(name string, data []byte) (err error) {
	if s.is_bbody == false {
		err = fmt.Errorf("Not use Bbody.")
		return
	}
	s.bbody[name] = data
	return
}

func (s *Spots) GetBbody(name string) (data []byte, err error) {
	if s.is_bbody == false {
		err = fmt.Errorf("Not use Bbody.")
		return
	}
	var have bool
	data, have = s.bbody[name]
	if have == false {
		err = fmt.Errorf("The Bbody not exist.")
		return
	}
	return
}

func (s *Spots) ReturnDelete() bool {
	return s.be_delete
}

func (s *Spots) SetDelete(del bool) {
	s.be_delete = del
}

func (s *Spots) MarshalBinary() (b []byte, err error) {
	buf := bytes.Buffer{}
	// id
	id_b := []byte(s.Id)
	id_b_len := int64(len(id_b))
	buf.Write(iendecode.Int64ToBytes(id_b_len))
	buf.Write(id_b)
	// relation
	var re_b []byte
	var re_len int64
	re_b, re_len, err = s.relationToBytes()
	if err != nil {
		return
	}
	buf.Write(iendecode.Int64ToBytes(re_len))
	buf.Write(re_b)
	// body data
	if s.Body != nil && s.is_bbody == false {
		var bbody map[string][]byte
		var body_b []byte
		var body_len int64
		bbody, err = s.Body.EncodeBbody()
		if err != nil {
			return
		}
		body_b = s.bbodyToBytes(bbody)
		body_len = int64(len(body_b))
		buf.Write(iendecode.Int64ToBytes(body_len))
		buf.Write(body_b)
	} else if s.is_bbody == true {
		var body_b []byte
		var body_len int64
		body_b = s.bbodyToBytes(s.bbody)
		body_len = int64(len(body_b))
		buf.Write(iendecode.Int64ToBytes(body_len))
		buf.Write(body_b)
	} else {
		buf.Write(iendecode.Int64ToBytes(0))
	}
	b = buf.Bytes()
	return
}

func (s *Spots) bbodyToBytes(bbody map[string][]byte) (b []byte) {
	var buf bytes.Buffer
	// map count
	thecount := int64(len(bbody))
	buf.Write(iendecode.Int64ToBytes(thecount))
	for key, _ := range bbody {
		// key
		key_b := []byte(key)
		key_b_len := int64(len(key_b))
		buf.Write(iendecode.Int64ToBytes(key_b_len))
		buf.Write(key_b)
		// value
		value_len := int64(len(bbody[key]))
		buf.Write(iendecode.Int64ToBytes(value_len))
		buf.Write(bbody[key])
	}
	b = buf.Bytes()
	return
}

func (s *Spots) relationToBytes() (b []byte, lens int64, err error) {
	buf := bytes.Buffer{}
	// bytes the Father: the string length + string
	father_b := []byte(s.father)
	father_b_len := int64(len(father_b))
	buf.Write(iendecode.Int64ToBytes(father_b_len))
	buf.Write(father_b)
	lens += 8 + father_b_len
	// bytes the Children: the children number + string length + string
	chilren_num := int64(len(s.children))
	buf.Write(iendecode.Int64ToBytes(chilren_num))
	lens += 8
	for i := range s.children {
		child_b := []byte(s.children[i])
		child_b_len := int64(len(child_b))
		buf.Write(iendecode.Int64ToBytes(child_b_len))
		buf.Write(child_b)
		lens += 8 + child_b_len
	}
	// bytes Friends: the Friends number + key length + key + value length + value
	friends_num := int64(len(s.friends))
	buf.Write(iendecode.Int64ToBytes(friends_num))
	lens += 8
	for key, _ := range s.friends {
		// the key
		key_b := []byte(key)
		key_b_len := int64(len(key_b))
		buf.Write(iendecode.Int64ToBytes(key_b_len))
		buf.Write(key_b)
		lens += 8 + key_b_len
		// the value
		value_b, err := s.friends[key].MarshalBinary()
		value_lens := int64(len(value_b))
		if err != nil {
			return nil, 0, err
		}
		buf.Write(iendecode.Int64ToBytes(value_lens))
		buf.Write(value_b)
		lens += 8 + value_lens
	}
	// bytes Contexts: the Contexts number + key length + key + value length + value
	contexts_num := int64(len(s.context))
	buf.Write(iendecode.Int64ToBytes(contexts_num))
	lens += 8
	for key, _ := range s.context {
		// the key
		key_b := []byte(key)
		key_b_len := int64(len(key_b))
		buf.Write(iendecode.Int64ToBytes(key_b_len))
		buf.Write(key_b)
		lens += 8 + key_b_len
		// the value
		value_b, err := s.context[key].MarshalBinary()
		if err != nil {
			return nil, 0, err
		}
		value_lens := int64(len(value_b))
		buf.Write(iendecode.Int64ToBytes(value_lens))
		buf.Write(value_b)
		lens += 8 + value_lens
	}

	b = buf.Bytes()
	return
}

func (s *Spots) UnmarshalBinary(b []byte) (err error) {
	buf := bytes.NewBuffer(b)
	// id
	id_len := iendecode.BytesToInt64(buf.Next(8))
	s.Id = string(buf.Next(int(id_len)))
	// relation
	rela_len := iendecode.BytesToInt64(buf.Next(8))
	err = s.bytesToRelation(buf.Next(int(rela_len)))
	if err != nil {
		return
	}
	// bbody
	bbody_len := iendecode.BytesToInt64(buf.Next(8))
	if bbody_len != 0 {
		err = s.bytesToBbody(buf.Next(int(bbody_len)))
		if err != nil {
			return
		}
		s.is_bbody = true
	} else {
		s.is_bbody = false
	}
	return
}

func (s *Spots) bytesToRelation(b []byte) (err error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	buf := bytes.NewBuffer(b)
	// de bytes the Father: the string length + string
	father_len := iendecode.BytesToInt64(buf.Next(8))
	s.father = string(buf.Next(int(father_len)))
	// de bytes the Children: the children number + string length + string
	children_num := iendecode.BytesToInt64(buf.Next(8))
	s.children = make([]string, children_num)
	var i int64
	for i = 0; i < children_num; i++ {
		child_len := iendecode.BytesToInt64(buf.Next(8))
		s.children[i] = string(buf.Next(int(child_len)))
	}
	// bytes Friends: the Friends number + key length + key + value length + value
	s.friends = make(map[string]Status)
	friends_num := iendecode.BytesToInt64(buf.Next(8))
	for i = 0; i < friends_num; i++ {
		key_len := iendecode.BytesToInt64(buf.Next(8))
		key := string(buf.Next(int(key_len)))
		value_len := iendecode.BytesToInt64(buf.Next(8))
		value_b := buf.Next(int(value_len))
		value := Status{}
		err = value.UnmarshalBinary(value_b)
		if err != nil {
			return
		}
		s.friends[key] = value
	}
	// bytes Contexts: the Contexts number + key length + key + value length + value
	s.context = make(map[string]Context)
	contexts_num := iendecode.BytesToInt64(buf.Next(8))
	for i = 0; i < contexts_num; i++ {
		key_len := iendecode.BytesToInt64(buf.Next(8))
		key := string(buf.Next(int(key_len)))
		value_len := iendecode.BytesToInt64(buf.Next(8))
		value_b := buf.Next(int(value_len))
		value := Context{}
		err = value.UnmarshalBinary(value_b)
		if err != nil {
			return
		}
		s.context[key] = value
	}
	return
}

func (s *Spots) bytesToBbody(b []byte) (err error) {
	s.bbody = make(map[string][]byte)
	buf := bytes.NewBuffer(b)
	thecount := iendecode.BytesToInt64(buf.Next(8))
	var i int64 = 0
	for {
		if i >= thecount {
			break
		}

		key_len := iendecode.BytesToInt64(buf.Next(8))
		key := string(buf.Next(int(key_len)))

		b_len := iendecode.BytesToInt64(buf.Next(8))
		s.bbody[key] = buf.Next(int(b_len))

		i++
	}
	return
}
