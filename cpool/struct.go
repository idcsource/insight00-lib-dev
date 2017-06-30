// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// 配置池，负责载入和管理多个配置文件。
//
// 配置信息为三层结构，分别为块（Block）、片（Section）、配置项（Config）。
// 块配置使用大括号定义名称，片配置使用中括号定义名称。
//
// 块名、片名、配置项末尾均可以添加注释，注释可以采用#或/开头，且可以单独一行或在配置行的行尾。
// 单独的注释行在使用WriteTo()方法保存配置时将不被保存进文件。
// 在配置行尾的注释将会被记录进WriteTo()方法保存的配置文件中。
//
// 配置文件的风格如下：
//     #注释行，WriteTo()方法将不保存这样的注释
//     {block_name}
//     [section_name]
//         config_key_name = config
//         #注释行，WriteTo()方法将不保存这样的注释
//         ...
//     [other_section_name]    #行中的注释，WriteTo()方法将保存这样的注释
//         other_config_key_name = config
//         ...
//
//     {other_block_name}    #行中的注释，WriteTo()方法将保存这样的注释
//     [section_name]
//         config_key_name = config    #行中的注释，WriteTo()方法将保存这样的注释
//         ...
//     [other_section_name]
//         other_config_key_name = config
//         ...
//
// 使用NewConfigPool()或NewConfigPoolNoFile()创建配置池。
package cpool

import (
	"regexp"
)

// 配置池
type ConfigPool struct {
	block map[string]*Block
	rege  map[string]*regexp.Regexp
	files []string   // 文件名
	lines [][]string // 每个文件里的每一行
}

// 配置池的编码模式
type PoolEncode struct {
	Blocks map[string]BlockEncode
}

// 块配置
type Block struct {
	section map[string]*Section
	key     string // 键
	notes   string // 注释信息
	file    int    // 哪个文件
	index   int    // 文件行计数位置
	new     bool   // 如果为true就为新建
	del     bool   // 如果为true就为删除
}

// 块的编码模式
type BlockEncode struct {
	Key      string
	Notes    string
	Sections map[string]SectionEncode
}

// 片配置
type Section struct {
	config map[string]*Config
	key    string // 键
	notes  string // 注释信息
	file   int    // 哪个文件
	index  int    // 文件行计数位置
	new    bool   // 如果为true就为新建
	del    bool   // 如果为true就为删除
}

// 片的编码模式
type SectionEncode struct {
	Key     string
	Notes   string
	Configs map[string]ConfigEncode
}

// 单项配置
type Config struct {
	key   string   // 键
	value string   // 值
	enum  []string // 枚举
	file  int      // 哪个文件
	index int      // 文件行计数位置
	notes string   // 注释信息
	new   bool     // 如果为true就为新建
	del   bool     // 如果为true就为删除
}

// 单项配置的编码模式
type ConfigEncode struct {
	Key   string
	Value string
	Notes string
}
