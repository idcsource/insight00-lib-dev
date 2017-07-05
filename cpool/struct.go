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
	"regexp"
)

// The Configure Pool
type ConfigPool struct {
	block map[string]*Block
	rege  map[string]*regexp.Regexp
	files []string   // config file's name
	lines [][]string // every files and every lines
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
	file  int      // which file
	index int      // the line count
	notes string   // if is new, it is true
	new   bool     // if is new, it is true
	del   bool     // if is true, it be deleted
}

// The Config's encode mode
type ConfigEncode struct {
	Key   string
	Value string
	Enum  []string
	Notes string
}
