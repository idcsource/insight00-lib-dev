// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// This source code is governed by GNU LGPL v3 license

// The Configure Pool - load and manage config files
//
// The configure information have three levels: Block, Section, Config.
//
// The Block use curly braces to define its name, the Section use square brackets to define its name.
//
// All lines can add comments after the line. The comments must start with #, and after # must have a space.
//
// The enumerate config, each item is searated by a space.
//
// The config files like this:
//     # a comment line
//     {block_name}
//     [section_name]
//         config_key_name = config
//         # a comment line
//         ...
//     [other_section_name]    # a comment
//         other_config_key_name = config
//         ...
//
//     {other_block_name}
//     [section_name]
//         config_key_name = config
//         ...
//     [other_section_name]
//         other_config_key_name = enum1 enum2 enum3  # a enumerate config
//         ...
package cpool

import (
	"bytes"
	"regexp"

	"github.com/idcsource/insight00-lib/iendecode"
)

// The Configure Pool
type ConfigPool struct {
	block map[string]*Block
	rege  map[string]*regexp.Regexp
	files []string   // config file's name
	lines [][]string // every files and every lines
}

func (c ConfigPool) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// block
	block_count := len(c.block)
	buf.Write(iendecode.IntToBytes(block_count))
	for key, _ := range c.block {
		key_b := []byte(key)
		key_b_len := len(key_b)
		buf.Write(iendecode.IntToBytes(key_b_len))
		buf.Write(key_b)

		var block_b []byte
		block_b, err = c.block[key].MarshalBinary()
		if err != nil {
			return
		}
		block_b_len := len(block_b)
		buf.Write(iendecode.IntToBytes(block_b_len))
		buf.Write(block_b)
	}

	// files
	files_b := iendecode.SliceStringToBytes(c.files)
	files_b_len := len(files_b)
	buf.Write(iendecode.IntToBytes(files_b_len))
	buf.Write(files_b)

	// lines
	lines_count := len(c.lines)
	buf.Write(iendecode.IntToBytes(lines_count))
	for i := range c.lines {
		one_b := iendecode.SliceStringToBytes(c.lines[i])
		one_b_len := len(one_b)
		buf.Write(iendecode.IntToBytes(one_b_len))
		buf.Write(one_b)
	}

	data = buf.Bytes()
	return
}

func (c *ConfigPool) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	// block
	c.block = make(map[string]*Block)
	block_count := iendecode.BytesToInt(buf.Next(8))
	for i := 0; i < block_count; i++ {
		key_b_len := iendecode.BytesToInt(buf.Next(8))
		key := string(buf.Next(key_b_len))

		block_b_len := iendecode.BytesToInt(buf.Next(8))
		block_b := buf.Next(block_b_len)
		block := &Block{}
		err = block.UnmarshalBinary(block_b)
		if err != nil {
			return
		}
		c.block[key] = block
	}

	// files
	files_b_len := iendecode.BytesToInt(buf.Next(8))
	c.files = iendecode.BytesToSliceString(buf.Next(files_b_len))

	// lines
	lines_count := iendecode.BytesToInt(buf.Next(8))
	c.lines = make([][]string, lines_count)
	for i := 0; i < lines_count; i++ {
		one_b_len := iendecode.BytesToInt(buf.Next(8))
		one := iendecode.BytesToSliceString(buf.Next(one_b_len))
		c.lines[i] = one
	}

	return
}

// The Configure Pool's encode mode
type PoolEncode struct {
	Blocks map[string]BlockEncode
}

// The Block
type Block struct {
	section map[string]*Section
	key     string // the key (the block name)
	notes   string // the comment text
	file    int    // which file
	index   int    // the line count
	new     bool   // if is new, it is true
	del     bool   // if is true, it be deleted
}

func (b Block) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// section
	section_count := len(b.section)
	buf.Write(iendecode.IntToBytes(section_count))
	for key, _ := range b.section {
		key_b := []byte(key)
		key_b_len := len(key_b)
		buf.Write(iendecode.IntToBytes(key_b_len))
		buf.Write(key_b)

		var section_b []byte
		section_b, err = b.section[key].MarshalBinary()
		if err != nil {
			return
		}
		section_b_len := len(section_b)
		buf.Write(iendecode.IntToBytes(section_b_len))
		buf.Write(section_b)
	}

	key_b := []byte(b.key)
	key_b_len := len(key_b)
	buf.Write(iendecode.IntToBytes(key_b_len))
	buf.Write(key_b)

	notes_b := []byte(b.notes)
	notes_b_len := len(notes_b)
	buf.Write(iendecode.IntToBytes(notes_b_len))
	buf.Write(notes_b)

	// 定长部分
	file_b := iendecode.IntToBytes(b.file)
	buf.Write(file_b) // 8

	index_b := iendecode.IntToBytes(b.index)
	buf.Write(index_b) //8

	new_b := iendecode.BoolToBytes(b.new)
	buf.Write(new_b) //1

	del_b := iendecode.BoolToBytes(b.del)
	buf.Write(del_b) //1

	data = buf.Bytes()
	return
}

func (b *Block) UnmarshalBinary(data []byte) (err error) {

	buf := bytes.NewBuffer(data)

	b.section = make(map[string]*Section)
	section_count := iendecode.BytesToInt(buf.Next(8))
	for i := 0; i < section_count; i++ {
		key_b_len := iendecode.BytesToInt(buf.Next(8))
		key := string(buf.Next(key_b_len))

		section_b_len := iendecode.BytesToInt(buf.Next(8))
		section_b := buf.Next(section_b_len)
		section := &Section{}
		err = section.UnmarshalBinary(section_b)
		if err != nil {
			return
		}
		b.section[key] = section
	}

	key_b_len := iendecode.BytesToInt(buf.Next(8))
	b.key = string(buf.Next(key_b_len))

	notes_b_len := iendecode.BytesToInt(buf.Next(8))
	b.notes = string(buf.Next(notes_b_len))

	b.file = iendecode.BytesToInt(buf.Next(8))

	b.index = iendecode.BytesToInt(buf.Next(8))

	b.new = iendecode.BytesToBool(buf.Next(1))

	b.del = iendecode.BytesToBool(buf.Next(1))

	return
}

// The Block's encode mode
type BlockEncode struct {
	Key      string
	Notes    string
	Sections map[string]SectionEncode
}

// The Section
type Section struct {
	config map[string]*Config
	key    string // the key (the section name)
	notes  string // the comment text
	file   int    // which file
	index  int    // the line count
	new    bool   // if is new, it is true
	del    bool   // if is true, it be deleted
}

func (s Section) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// config
	config_count := len(s.config)
	buf.Write(iendecode.IntToBytes(config_count))
	for key, _ := range s.config {
		key_b := []byte(key)
		key_b_len := len(key_b)
		buf.Write(iendecode.IntToBytes(key_b_len))
		buf.Write(key_b)

		var config_b []byte
		config_b, err = s.config[key].MarshalBinary()
		if err != nil {
			return
		}
		config_b_len := len(config_b)
		buf.Write(iendecode.IntToBytes(config_b_len))
		buf.Write(config_b)
	}

	key_b := []byte(s.key)
	key_b_len := len(key_b)
	buf.Write(iendecode.IntToBytes(key_b_len))
	buf.Write(key_b)

	notes_b := []byte(s.notes)
	notes_b_len := len(notes_b)
	buf.Write(iendecode.IntToBytes(notes_b_len))
	buf.Write(notes_b)

	// 定长部分
	file_b := iendecode.IntToBytes(s.file)
	buf.Write(file_b) // 8

	index_b := iendecode.IntToBytes(s.index)
	buf.Write(index_b) //8

	new_b := iendecode.BoolToBytes(s.new)
	buf.Write(new_b) //1

	del_b := iendecode.BoolToBytes(s.del)
	buf.Write(del_b) //1

	data = buf.Bytes()
	return
}

func (s *Section) UnmarshalBinary(data []byte) (err error) {

	buf := bytes.NewBuffer(data)

	s.config = make(map[string]*Config)
	config_count := iendecode.BytesToInt(buf.Next(8))
	for i := 0; i < config_count; i++ {
		key_b_len := iendecode.BytesToInt(buf.Next(8))
		key := string(buf.Next(key_b_len))

		config_b_len := iendecode.BytesToInt(buf.Next(8))
		config_b := buf.Next(config_b_len)
		config := &Config{}
		err = config.UnmarshalBinary(config_b)
		if err != nil {
			return
		}
		s.config[key] = config
	}

	key_b_len := iendecode.BytesToInt(buf.Next(8))
	s.key = string(buf.Next(key_b_len))

	notes_b_len := iendecode.BytesToInt(buf.Next(8))
	s.notes = string(buf.Next(notes_b_len))

	s.file = iendecode.BytesToInt(buf.Next(8))

	s.index = iendecode.BytesToInt(buf.Next(8))

	s.new = iendecode.BytesToBool(buf.Next(1))

	s.del = iendecode.BytesToBool(buf.Next(1))

	return
}

// The Section's encode mode
type SectionEncode struct {
	Key     string
	Notes   string
	Configs map[string]ConfigEncode
}

// One Config
type Config struct {
	key   string   // the key
	value string   // the value
	enum  []string // the enum
	notes string   // if is new, it is true
	file  int      // which file
	index int      // the line count
	new   bool     // if is new, it is true
	del   bool     // if is true, it be deleted
}

func (c Config) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	// 不定长部分
	key_b := []byte(c.key)
	key_b_len := len(key_b)
	buf.Write(iendecode.IntToBytes(key_b_len))
	buf.Write(key_b)

	value_b := []byte(c.value)
	value_b_len := len(value_b)
	buf.Write(iendecode.IntToBytes(value_b_len))
	buf.Write(value_b)

	enum_b := iendecode.SliceStringToBytes(c.enum)
	enum_b_len := len(enum_b)
	buf.Write(iendecode.IntToBytes(enum_b_len))
	buf.Write(enum_b)

	notes_b := []byte(c.notes)
	notes_b_len := len(notes_b)
	buf.Write(iendecode.IntToBytes(notes_b_len))
	buf.Write(notes_b)

	// 定长部分
	file_b := iendecode.IntToBytes(c.file)
	buf.Write(file_b) // 8

	index_b := iendecode.IntToBytes(c.index)
	buf.Write(index_b) //8

	new_b := iendecode.BoolToBytes(c.new)
	buf.Write(new_b) //1

	del_b := iendecode.BoolToBytes(c.del)
	buf.Write(del_b) //1

	data = buf.Bytes()

	return
}
func (c *Config) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)

	key_b_len := iendecode.BytesToInt(buf.Next(8))
	c.key = string(buf.Next(key_b_len))

	value_b_len := iendecode.BytesToInt(buf.Next(8))
	c.value = string(buf.Next(value_b_len))

	enum_b_len := iendecode.BytesToInt(buf.Next(8))
	c.enum = iendecode.BytesToSliceString(buf.Next(enum_b_len))

	notes_b_len := iendecode.BytesToInt(buf.Next(8))
	c.notes = string(buf.Next(notes_b_len))

	c.file = iendecode.BytesToInt(buf.Next(8))

	c.index = iendecode.BytesToInt(buf.Next(8))

	c.new = iendecode.BytesToBool(buf.Next(1))

	c.del = iendecode.BytesToBool(buf.Next(1))

	return
}

// The Config's encode mode
type ConfigEncode struct {
	Key   string
	Value string
	Enum  []string
	Notes string
}
