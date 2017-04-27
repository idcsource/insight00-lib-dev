// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// drule2的远程控制者控制台终端
package operator_t

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
)

// 创建
func NewOperatorT(o *operator.Operator) (ot *OperatorT) {
	ot = &OperatorT{
		operator: o,
		regexp:   make(map[string]*regexp.Regexp),
	}
	ot.regexp["space"], _ = regexp.Compile(`[^ ]+`)
	ot.regexp["'b"], _ = regexp.Compile(`^'`)
	ot.regexp["b'"], _ = regexp.Compile(`'$`)
	return
}

func (o *OperatorT) CommandSplit(command string) (split []string, err error) {
	return o.commandSplit(command)
}

// 将命令按照规则进行拆解
func (o *OperatorT) commandSplit(command string) (split []string, err error) {
	// 第一遍用空格拆分
	tmp1 := strings.Split(command, " ")

	split = make([]string, 0)
	tmp1_l := len(tmp1)
	for i := 0; i < tmp1_l; i++ {
		t := tmp1[i]
		if o.regexp["'b"].MatchString(tmp1[i]) || tmp1[i] == "'" {
			// 合并引号
			if o.regexp["b'"].MatchString(tmp1[i]) && tmp1[i] != "'" {
				// 一节的情况不做处理
			} else {
				for j := i + 1; j <= tmp1_l; j++ {
					if j == tmp1_l {
						// 如果最后没有结尾
						err = fmt.Errorf("Command syntax error.")
						return make([]string, 0), err
					}
					if o.regexp["b'"].MatchString(tmp1[j]) {
						t = t + " " + tmp1[j]
						i = j
						break
					} else {
						t = t + " " + tmp1[j]
					}

				}
			}
		} else {
			// 处理空格
			if o.regexp["space"].MatchString(t) == false {
				continue
			}
		}
		t = strings.Trim(t, "'")
		t = strings.TrimSpace(t)
		split = append(split, t)
	}
	return
}
