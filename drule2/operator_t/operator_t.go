// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

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

// 命令输入，并返回结构字符集，一行一个string
func (o *OperatorT) Command(command string) (r [][]string, err error) {
	// 拆分命令
	split, err := o.commandSplit(command)
	if err != nil {
		return
	}
	// 交给执行器
	r, err = o.exec(split)
	return
}

// 命令执行器
func (o *OperatorT) exec(split []string) (r [][]string, err error) {
	split_l := len(split)
	if split_l == 1 {
		if split[0] == "help" {
			r, err = o.execHelp()
		} else {
			err = fmt.Errorf("Command syntax error.")
		}
		return
	}
	if split_l < 2 {
		err = fmt.Errorf("Command syntax error.")
		return
	}
	switch split[0] {
	case "area":
		r, err = o.execArea(split, split_l)
	case "user":
		r, err = o.execUser(split, split_l)
	default:
		err = fmt.Errorf("Command syntax error.")
	}
	return
}

// 处理帮助
func (o *OperatorT) execHelp() (r [][]string, err error) {
	r = make([][]string, 0)
	r = append(r, []string{"获取帮助", "help"})
	r = append(r, []string{"列出所有区域", "area list"})
	r = append(r, []string{"添加区域", "area add 'area_name'"})
	r = append(r, []string{"删除区域", "area delete 'area_name'"})
	r = append(r, []string{"区域改名", "area rename 'old_name' to 'new_name'"})
	r = append(r, []string{"列出所有用户", "user list"})
	r = append(r, []string{"新建用户", "user add 'user_name' for {drule|admin|normal} with password 'password' [and email 'email']"})
	r = append(r, []string{"删除用户", "user delete 'user_name'"})
	r = append(r, []string{"修改用户密码", "alert user 'user_name' password 'password'"})
	r = append(r, []string{"修改用户邮箱", "alert user 'user_name' email 'email'"})
	r = append(r, []string{"添加用户的区域权限", "alert user 'user_name' with 'area_name' {writable|readonly}"})
	r = append(r, []string{"删除用户的区域权限", "alert user 'user_name' without 'area_name'"})
	return
}

// 主要是用来检查命令的
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
			} else {
				t = strings.ToLower(t)
			}
		}
		t = strings.Trim(t, "'")
		t = strings.TrimSpace(t)
		split = append(split, t)
	}
	return
}
