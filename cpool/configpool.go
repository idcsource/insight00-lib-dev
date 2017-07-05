// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package cpool

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/pubfunc"
)

// 新建一个配置池，并将提供的配置文件读出解析
func NewConfigPool(fname ...string) (*ConfigPool, error) {
	c := &ConfigPool{
		block: make(map[string]*Block),
		rege:  make(map[string]*regexp.Regexp),
		files: make([]string, 0),
		lines: make([][]string, 0),
	}
	c.rege["block"], _ = regexp.Compile(`^{(\s*)(.+)(\s*)}$`)   //匹配{xxxx}
	c.rege["config"], _ = regexp.Compile(`(.+)(\s*)=(\s*)(.+)`) //匹配xxx=xxx
	c.rege["section"], _ = regexp.Compile(`\[(\s*)(.+)(\s*)\]`) //匹配[xxx]

	for _, f := range fname {
		c.files = append(c.files, pubfunc.LocalFile(f))
	}

	for i, onef := range c.files {
		file, err1 := os.Open(onef)
		if err1 != nil {
			return nil, fmt.Errorf("cpool: %v", err1)
		}
		if err2 := c.read(i, bufio.NewReader(file)); err2 != nil {
			// 上面是读入每个文件的行
			return nil, fmt.Errorf("cpool: %v", err2)
		}
		if err3 := file.Close(); err3 != nil {
			return nil, fmt.Errorf("cpool: %v", err3)
		}
	}
	return c, nil
}

// 新建一个没有文件的空配置池
// 你可以随后用RegFile()、Reload()等方法动态设置自己的配置池
func NewConfigPoolNoFile() *ConfigPool {
	c := &ConfigPool{
		block: make(map[string]*Block),
		rege:  make(map[string]*regexp.Regexp),
		files: make([]string, 0),
		lines: make([][]string, 0),
	}
	c.rege["block"], _ = regexp.Compile(`^{(.+)}$`)             //匹配{xxxx}
	c.rege["config"], _ = regexp.Compile(`(.+)(\s*)=(\s*)(.+)`) //匹配xxx=xxx
	c.rege["section"], _ = regexp.Compile(`\[(\s*)(.+)(\s*)\]`) //匹配[xxx]
	return c
}

// 注册一个文件到配置池，你需要用Reload()方法对载入的文件进行处理
func (c *ConfigPool) RegFile(fname ...string) error {
	for _, f := range fname {
		fs := pubfunc.LocalFile(f)
		if pubfunc.FileExist(fs) == false {
			return errors.New("cpool: [ConfigPool]RegFile: Can't find the file " + f)
		}
		c.files = append(c.files, fs)
	}
	return nil
}

// Reload 重载配置文件
// 一定小心使用这个功能，它将会把当前的配置全部清空，如果你自己修改添加了配置也将一并消失
func (c *ConfigPool) Reload() error {
	if len(c.files) == 0 {
		return errors.New("There have no file be registered for configure.")
	}
	c.block = make(map[string]*Block)
	for i, onef := range c.files {
		file, err1 := os.Open(onef)
		if err1 != nil {
			return fmt.Errorf("cpool: %v", err1)
		}
		if err2 := c.read(i, bufio.NewReader(file)); err2 != nil {
			// 上面是读入每个文件的行
			return fmt.Errorf("cpool: %v", err2)
		}
		if err3 := file.Close(); err3 != nil {
			return fmt.Errorf("cpool: %v", err3)
		}
	}
	return nil
}

// 将配置文件存储到某个文件中
func (c *ConfigPool) WriteTo(filename string) error {
	filename = pubfunc.LocalFile(filename)
	var file string
	file += "// Insight 0+0 cpool writer configure file.\n"
	for block_name, one_block := range c.block {
		the_new_line := "{" + block_name + "}\n"
		file += "\n"
		file += the_new_line
		for section_name, one_section := range one_block.section {
			the_new_line := "[" + section_name + "]\n"
			file += "\n"
			file += the_new_line
			for config_name, one_config := range one_section.config {
				var thenode string
				if len(one_config.notes) != 0 {
					thenode = "    " + one_config.notes
				}
				the_new_line := "\t" + config_name + " = " + one_config.value + thenode + "\n"
				file += the_new_line
			}
		}
	}
	e := ioutil.WriteFile(filename, []byte(file), 0666)
	return e
}

// 将配置池导出成编码模式
func (c *ConfigPool) EncodePool() PoolEncode {
	encode := PoolEncode{
		Blocks: make(map[string]BlockEncode),
	}
	for blockkey, oneblock := range c.block {
		encode.Blocks[blockkey] = BlockEncode{
			Key:      oneblock.key,
			Notes:    oneblock.notes,
			Sections: make(map[string]SectionEncode),
		}
		for sectionkey, onesection := range oneblock.section {
			encode.Blocks[blockkey].Sections[sectionkey] = SectionEncode{
				Key:     onesection.key,
				Notes:   onesection.notes,
				Configs: make(map[string]ConfigEncode),
			}
			for configkey, oneconfig := range onesection.config {
				encode.Blocks[blockkey].Sections[sectionkey].Configs[configkey] = ConfigEncode{
					Key:   oneconfig.key,
					Value: oneconfig.value,
					Enum:  oneconfig.enum,
					Notes: oneconfig.notes,
				}
			}
		}
	}
	return encode
}

// 解码配置池并装入自身
func (c *ConfigPool) DecodePool(pool PoolEncode) {
	for key, oneblock := range pool.Blocks {
		_, find := c.block[key]
		if find == false {
			c.block[key] = NewBlock(key, oneblock.Notes)
		}
		c.block[key].DecodeBlock(oneblock)
		if find == true {
			c.block[key].new = false
		}
	}
}

// 解码一个块并装入相应位置，如果有存在的配置项则会被覆盖
func (c *ConfigPool) DecodeBlock(bl BlockEncode) {
	bname := bl.Key
	_, find := c.block[bname]
	if find == false {
		c.block[bname] = NewBlock(bname, bl.Notes)
	}
	c.block[bname].DecodeBlock(bl)
	if find == true {
		c.block[bname].new = false
	}
}

// 解码一个片并装入相应位置，blockname为指定装入哪个块中，如果块不存在则新建，如果任何配置项已经存在则会覆盖
func (c *ConfigPool) DecodeSection(blockname string, sectioncode SectionEncode) {
	_, blockfind := c.block[blockname]
	if blockfind == false {
		c.block[blockname] = NewBlock(blockname, "")
	}
	c.block[blockname].DecodeSection(sectioncode)
}

// 将某个块导出为编码模式，如果块不存在，则返回错误
func (c *ConfigPool) EncodeBlock(b string) (encode BlockEncode, err error) {
	encode = BlockEncode{
		Sections: make(map[string]SectionEncode),
	}
	b = strings.TrimSpace(b)
	oneblock, find := c.block[b]
	if find == false {
		err = errors.New("cpool: [ConfigPool]EncodeBlock: Can't find the block : " + b)
		return
	}
	encode.Key = oneblock.key
	encode.Notes = oneblock.notes
	for sectionkey, onesection := range oneblock.section {
		encode.Sections[sectionkey] = SectionEncode{
			Key:     onesection.key,
			Notes:   onesection.notes,
			Configs: make(map[string]ConfigEncode),
		}
		for configkey, oneconfig := range onesection.config {
			encode.Sections[sectionkey].Configs[configkey] = ConfigEncode{
				Key:   oneconfig.key,
				Value: oneconfig.value,
				Enum:  oneconfig.enum,
				Notes: oneconfig.notes,
			}
		}
	}
	return
}

// 将某个片导出为编码模式，如果片不存在，则返回错误，s为block|section格式
func (c *ConfigPool) EncodeSection(s string) (encode SectionEncode, err error) {
	encode = SectionEncode{
		Configs: make(map[string]ConfigEncode),
	}
	s = strings.TrimSpace(s)
	sA := strings.Split(s, "|")
	if len(sA) != 2 {
		err = errors.New("cpool: [ConfigPool]EncodeSection: Request configuration node error : " + s)
		return
	}
	blockname := strings.TrimSpace(sA[0])
	sectionname := strings.TrimSpace(sA[1])
	block, err := c.GetBlock(blockname)
	if err != nil {
		err = errors.New("cpool: [ConfigPool]EncodeSection: Can't find the Section : " + s)
		return
	}
	onesection, ok := block.section[sectionname]
	if ok == false {
		err = errors.New("cpool: [ConfigPool]EncodeSection: Can't find the Section : " + s)
		return
	}
	encode.Key = onesection.key
	encode.Notes = onesection.notes
	for configkey, oneconfig := range onesection.config {
		encode.Configs[configkey] = ConfigEncode{
			Key:   oneconfig.key,
			Value: oneconfig.value,
			Enum:  oneconfig.enum,
			Notes: oneconfig.notes,
		}
	}
	return
}

// read 读出所有配置文件的行，并放入line中，消灭掉所有空白行和只有注释的行，然后解析所有的行
func (c *ConfigPool) read(i int, buf *bufio.Reader) error {
	a := make([]string, 0)
	c.lines = append(c.lines, a)
	for {
		l, err1 := buf.ReadString('\n')
		if err1 == io.EOF {
			if len(l) == 0 {
				break
			}
		} else if err1 != nil {
			return err1
		}
		c.lines[i] = append(c.lines[i], l)
	}
	err2 := c.doLines(i) // c.doLines 处理所有的行
	if err2 != nil {
		return err2
	}
	return nil
}

// stripComments 去除掉每一行里的注释
func (c *ConfigPool) stripComments(l string) (string, string) {
	var no string
	for _, c := range []string{" ;", "\t;", " #", "\t#", " //", "\t//", " --", "\t--"} {
		if i := strings.Index(l, c); i != -1 {
			no = l[i+1:]
			l = l[0:i]
		}
	}
	return l, no
}

// doLines 处理 c.line中的每一行
func (c *ConfigPool) doLines(filei int) error {
	ij := 0
	for i, l := range c.lines[filei] {
		if i < ij {
			continue
		}
		lnoc := strings.TrimSpace(l)
		if len(lnoc) == 0 || lnoc[0] == '#' || lnoc[0] == '/' || lnoc[0] == '-' || lnoc[0] == ';' {
			continue
		}
		thetype, thename, thenote, err := c.checkLine(lnoc)
		if err != nil {
			return err
		}
		if thetype == 1 {
			continue
		}

		//		var notes string
		//		lnoc, notes = c.stripComments(lnoc)
		//		lnoc = strings.TrimSpace(lnoc)
		//		notes = strings.TrimSpace(notes)
		//		if len(lnoc) == 0 {
		//			continue
		//		}
		//if c.rege["block"].MatchString(lnoc) == true {
		if thetype == 2 {
			//theR := c.rege["block"].FindStringSubmatch(lnoc)
			//theR1 := strings.TrimSpace(theR[1])
			if len(c.lines[filei]) >= i+1 {
				r, j, err1 := c.doSection(filei, i+1) // c.doSection 将所有后续的行放到doSection里处理一个config文件
				if err1 != nil {
					return err1
				}
				//r.key = theR1
				r.key = thename
				//r.notes = notes
				r.notes = thenote
				ij = j
				c.block[thename] = r
			}
		}
	}
	return nil
}

// doSection 处理所有的[xxxx]以及之内的东西，直到碰到block标记
func (c *ConfigPool) doSection(filei int, linei int) (*Block, int, error) {
	ca := &Block{
		section: make(map[string]*Section),
		file:    filei,
		index:   linei - 1,
		new:     false,
		del:     false,
	}
	var ij int
	var ri int
	ri = linei
	for i, l := range c.lines[filei][linei:] {
		if i+linei < ij {
			continue
		}
		lnoc := strings.TrimSpace(l)
		if len(lnoc) == 0 || lnoc[0] == '#' || lnoc[0] == '/' || lnoc[0] == '-' || lnoc[0] == ';' {
			continue
		}
		thetype, thename, thenote, err := c.checkLine(lnoc)
		if err != nil {
			return nil, 0, err
		}
		if thetype == 1 {
			continue
		}
		//		var notes string
		//		lnoc, notes = c.stripComments(lnoc)
		//		lnoc = strings.TrimSpace(lnoc)
		//		notes = strings.TrimSpace(notes)
		//		if len(lnoc) == 0 {
		//			continue
		//		}
		//if c.rege["section"].MatchString(lnoc) == true {
		if thetype == 3 {
			//theR := c.rege["section"].FindStringSubmatch(lnoc)
			//theR1 := strings.TrimSpace(theR[2])
			if len(c.lines[filei][linei:]) >= i+1 {
				r, j, err1 := c.doConfig(filei, linei+i+1) // 处理每一个 xxx = xxx
				if err1 != nil {
					return nil, 0, err1
				}
				//r.key = theR1
				r.key = thename
				//r.notes = notes
				r.notes = thenote
				//ca.section[theR1] = r
				ca.section[thename] = r
				ij = j
			}
			//} else if c.rege["block"].MatchString(lnoc) == true {
		} else if thetype == 2 {
			ri = i
			break
		}
	}
	return ca, ri, nil
}

// 处理每一个xxx = xxxx
func (c *ConfigPool) doConfig(filei int, linei int) (*Section, int, error) {
	ca := &Section{
		config: make(map[string]*Config),
		file:   filei,
		index:  linei - 1,
		new:    false,
		del:    false,
	}
	var ri int
	for i, l := range c.lines[filei][linei:] {
		lnoc := strings.TrimSpace(l)
		// if this line is comment
		if len(lnoc) == 0 || lnoc[0] == '#' || lnoc[0] == '/' || lnoc[0] == '-' || lnoc[0] == ';' {
			continue
		}
		thetype, _, _, err := c.checkLine(lnoc)
		if err != nil {
			return nil, 0, err
		}
		if thetype == 1 {
			continue
		}
		//if c.rege["config"].MatchString(lnoc) == true {
		if thetype == 4 {
			key, value, notes, err := c.splitConfig(lnoc)
			if err != nil {
				return nil, 0, err
			}
			ca.config[key] = &Config{
				key:   key,
				value: value[0],
				enum:  value,
				file:  filei,
				index: linei + i,
				notes: notes,
				new:   false,
				del:   false,
			}
			//} else if c.rege["section"].MatchString(lnoc) == true || c.rege["block"].MatchString(lnoc) == true {
		} else if thetype == 2 || thetype == 3 {
			ri = i
			break
		}

		//		var notes string
		//		lnoc, notes = c.stripComments(lnoc)
		//		lnoc = strings.TrimSpace(lnoc)
		//		notes = strings.TrimSpace(notes)
		//		if len(lnoc) == 0 {
		//			continue
		//		}
		//		if c.rege["config"].MatchString(lnoc) == true {
		//			theR := c.rege["config"].FindStringSubmatch(lnoc)
		//			key := strings.TrimSpace(theR[1])
		//			value := strings.TrimSpace(theR[4])
		//			ca.config[key] = &Config{
		//				key:   key,
		//				value: value,
		//				file:  filei,
		//				index: linei + i,
		//				notes: notes,
		//				new:   false,
		//				del:   false,
		//			}
		//		} else if c.rege["section"].MatchString(lnoc) == true || c.rege["block"].MatchString(lnoc) == true {
		//			ri = i
		//			break
		//		}
	}
	return ca, ri, nil
}

func (c *ConfigPool) checkLine(line string) (thetype uint8, name, note string, err error) {
	thesplit, err := pubfunc.CommandSplit(line, false)
	if err != nil {
		return
	}
	now_type := 0
	for _, onesplit := range thesplit {
		onesplit2 := strings.TrimSpace(onesplit)
		if len(onesplit2) == 0 && now_type != 2 {
			continue
		}
		switch now_type {
		case 0:
			if onesplit == "#" || onesplit == "//" || onesplit == "--" {
				thetype = 1 // 1 is comment
				return
			} else if c.rege["block"].MatchString(onesplit) == true {
				thetype = 2 // 2 is block
				name = strings.TrimSpace(c.rege["block"].FindStringSubmatch(onesplit)[2])
				now_type = 1
			} else if c.rege["section"].MatchString(onesplit) == true {
				thetype = 3 // 3 is section
				name = strings.TrimSpace(c.rege["section"].FindStringSubmatch(onesplit)[2])
				now_type = 1
			} else {
				thetype = 4 // 4 is config
				return
			}
		case 1:
			if onesplit == "#" || onesplit == "//" || onesplit == "--" {
				now_type = 2
			}
		case 2:
			if len(onesplit) == 0 {
				note += " "
			} else {
				note += onesplit
			}
		}
	}
	return
}

func (c *ConfigPool) splitConfig(line string) (key string, value []string, note string, err error) {
	thesplit, err := pubfunc.CommandSplit(line, false)
	if err != nil {
		return
	}
	now_type := 0
	value = make([]string, 0)
	for _, onesplit := range thesplit {
		onesplit2 := strings.TrimSpace(onesplit)
		if len(onesplit2) == 0 && now_type != 3 {
			continue
		}
		switch now_type {
		case 0:
			// 0 is not have the key
			key = onesplit
			now_type = 1
		case 1:
			// 1 is not have =
			if onesplit == "=" {
				now_type = 2
				continue
			}
		case 2:
			// 2 is not have enough value
			if onesplit == "#" || onesplit == "//" || onesplit == "--" {
				now_type = 3
				continue
			}
			value = append(value, onesplit)
		case 3:
			// 3 is the note
			if len(onesplit) == 0 {
				note += " "
			} else {
				note += onesplit
			}

		}
	}
	note = strings.TrimSpace(note)
	if len(key) == 0 || len(value) == 0 {
		err = fmt.Errorf("Command syntax error.")
	}
	return
}

// 获取所有块（Block）的名字
func (c *ConfigPool) GetAllBlockName() []string {
	re := make([]string, 0)
	for name, _ := range c.block {
		re = append(re, name)
	}
	return re
}

// 返回一个块（Block）的配置内容
func (c *ConfigPool) GetBlock(ns string) (*Block, error) {
	ns = strings.TrimSpace(ns)
	nils, find := c.block[ns]
	if find == false {
		return nil, errors.New("cpool: [ConfigPool]GetBlock: Can't find the config : " + ns)
	} else {
		return nils, nil
	}
}

// 获取一个片(Section)内的配置内容，格式是：block|section
func (c *ConfigPool) GetSection(s string) (*Section, error) {
	s = strings.TrimSpace(s)
	sA := strings.Split(s, "|")
	if len(sA) != 2 {
		return nil, errors.New("cpool: [ConfigPool]GetSection: Request configuration node error : " + s)
	}
	ns := strings.TrimSpace(sA[0])
	cs := strings.TrimSpace(sA[1])
	rsA, err := c.GetBlock(ns)
	if err != nil {
		return nil, errors.New("cpool: [ConfigPool]GetSection: Can't find the config : " + s)
	}
	rsAs, ok := rsA.section[cs]
	if ok == false {
		return nil, errors.New("cpool: [ConfigPool]GetSection: Can't find the config : " + s)
	}
	return rsAs, nil
}

// 将自己后来添加的Block注册进配置池，如果配置池中已经有同名的则返回错误
func (c *ConfigPool) RegBlock(b *Block) error {
	bkey := b.key
	_, find := c.block[bkey]
	if find == true {
		return errors.New("cpool: [ConfigPool]RegBlock: The Block " + bkey + " is already exist.")
	}
	c.block[bkey] = b
	return nil
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (c *ConfigPool) SetConfig(s, v, n string) error {
	s = strings.TrimSpace(s)
	sA := strings.Split(s, "|")
	if len(sA) != 2 {
		return errors.New("cpool: [ConfigPool]SetConfig: Request configuration node error : " + s)
	}
	ns := strings.TrimSpace(sA[0]) //block
	cs := strings.TrimSpace(sA[1])
	csA := strings.Split(cs, ".")
	if len(csA) != 2 {
		return errors.New("cpool: [ConfigPool]SetConfig: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0]) //section
	csA2 := strings.TrimSpace(csA[1]) //config
	rsA, err := c.GetBlock(ns)
	if err != nil {
		rsA = NewBlock(ns, "")
		c.block[ns] = rsA
	}
	rsB, ok := rsA.section[csA1]
	if ok == false {
		rsB = NewSection(csA1, "")
		rsA.section[csA1] = rsB
	}
	rs, ok2 := rsB.config[csA2]
	if ok2 == false {
		rsB.config[csA2] = &Config{
			key:   csA2,
			notes: "#" + n,
			value: v,
			new:   true,
		}
	} else {
		rs.value = v
		if len(n) != 0 {
			rs.notes = "#" + n
		}
	}
	return nil
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (c *ConfigPool) SetInt64(s string, v int64, n string) error {
	var vs string
	vs = strconv.FormatInt(v, 10)
	return c.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (c *ConfigPool) SetFloat(s string, v float64, n string) error {
	var vs string
	vs = strconv.FormatFloat(v, 'f', -1, 64)
	return c.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
func (c *ConfigPool) SetBool(s string, v bool, n string) error {
	var vs string
	if v == true {
		vs = "true"
	} else if v == false {
		vs = "false"
	}
	return c.SetConfig(s, vs, n)
}

// 设置一个配置项，s的格式是：block|section.configkey，v设置的值，n为注释
// 如果配置项已经存在，将改写设置的值，否则新建项目
// 注意这里的注释为第二个参数
func (c *ConfigPool) SetEnum(s string, n string, v ...string) error {
	var vs string
	vs = strings.Join(v, ",")
	return c.SetConfig(s, vs, n)
}

// 获取一个配置，没有找到将返回错误，s的格式是：block|section.keyname
func (c *ConfigPool) GetConfig(s string) (string, error) {
	s = strings.TrimSpace(s)
	sA := strings.Split(s, "|")
	if len(sA) != 2 {
		return "", errors.New("cpool: [ConfigPool]GetConfig: Request configuration node error : " + s)
	}
	ns := strings.TrimSpace(sA[0]) //block
	cs := strings.TrimSpace(sA[1])
	csA := strings.Split(cs, ".")
	if len(csA) != 2 {
		return "", errors.New("cpool: [ConfigPool]GetConfig: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0]) //section
	csA2 := strings.TrimSpace(csA[1]) //config
	rsA, err := c.GetBlock(ns)
	if err != nil {
		return "", errors.New("cpool: [ConfigPool]GetConfig: Can't find the config : " + s)
	}
	rsB, ok := rsA.section[csA1]
	if ok == false {
		return "", errors.New("cpool: [ConfigPool]GetConfig: Can't find the config : " + s)
	}
	rs, ok2 := rsB.config[csA2]
	if ok2 == false {
		return "", errors.New("cpool: [ConfigPool]GetConfig: Can't find the config : " + s)
	}
	return rs.value, nil
}

// 获取一个配置，并转换为字符串切片，转换失败或没有找到将返回错误，s的格式是：block|section.keyname
func (c *ConfigPool) GetEnum(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	sA := strings.Split(s, "|")
	if len(sA) != 2 {
		return nil, errors.New("cpool: [ConfigPool]GetEnum: Request configuration node error : " + s)
	}
	ns := strings.TrimSpace(sA[0]) //block
	cs := strings.TrimSpace(sA[1])
	csA := strings.Split(cs, ".")
	if len(csA) != 2 {
		return nil, errors.New("cpool: [ConfigPool]GetEnum: Request configuration node error : " + s)
	}
	csA1 := strings.TrimSpace(csA[0]) //section
	csA2 := strings.TrimSpace(csA[1]) //config
	rsA, err := c.GetBlock(ns)
	if err != nil {
		return nil, errors.New("cpool: [ConfigPool]GetEnum: Can't find the config : " + s)
	}
	rsB, ok := rsA.section[csA1]
	if ok == false {
		return nil, errors.New("cpool: [ConfigPool]GetEnum: Can't find the config : " + s)
	}
	rs, ok2 := rsB.config[csA2]
	if ok2 == false {
		return nil, errors.New("cpool: [ConfigPool]GetEnum: Can't find the config : " + s)
	}
	return rs.enum, nil
	//	config, err := c.GetConfig(s)
	//	if err != nil {
	//		return nil, err
	//	}
	//	configa := strings.Split(config, ",")
	//	returna := make([]string, 0)
	//	for _, v := range configa {
	//		v = strings.TrimSpace(v)
	//		if len(v) > 0 {
	//			returna = append(returna, v)
	//		}
	//	}
	//	return returna, nil
}

// 获取一个配置，并转换为64位整数，转换失败或没有找到将返回错误，s的格式是：block|section.keyname
func (c *ConfigPool) TranInt64(s string) (int64, error) {
	cf, err := c.GetConfig(s)
	if err != nil {
		return 0, errors.New("pool: [ConfigPool]TranInt64: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseInt(cf, 10, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil

}

// 获取一个配置，并转换为64位浮点，转换失败或没有找到将返回错误，s的格式是：block|section.keyname
func (c *ConfigPool) TranFloat(s string) (float64, error) {
	cf, err := c.GetConfig(s)
	if err != nil {
		return 0, errors.New("pool: [ConfigPool]TranFloat: Request configuration node error : " + s)
	}
	i, err2 := strconv.ParseFloat(cf, 64)
	if err2 != nil {
		return 0, err2
	}
	return i, nil

}

// 获取一个配置，并转换为布尔值，转换失败或没有找到将返回错误，s的格式是：block|section.keyname
func (c *ConfigPool) TranBool(s string) (bool, error) {
	cf, err := c.GetConfig(s)
	if err != nil {
		return false, errors.New("pool: [ConfigPool]TranBool: Request configuration node error : " + s)
	}
	cf = strings.ToLower(cf)
	if cf == "true" || cf == "yes" || cf == "t" || cf == "y" {
		return true, nil
	} else if cf == "false" || cf == "no" || cf == "f" || cf == "n" {
		return false, nil
	} else {
		return false, errors.New("pool: [ConfigPool]TranBool: Not bool : " + s)
	}
}
