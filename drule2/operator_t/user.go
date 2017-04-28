// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package operator_t

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
)

// user的命令执行器，从split[1]开始
func (o *OperatorT) execUser(split []string, l int) (r [][]string, err error) {
	errs := operator.NewDRuleError()
	// area list
	if split[1] == "list" {
		rt, errs := o.operator.UserList()
		err = errs.IsError()
		r = make([][]string, 0)
		for _, rto := range rt {
			var auth string
			switch rto.Authority {
			case operator.USER_AUTHORITY_DRULE:
				auth = "drule"
			case operator.USER_AUTHORITY_NORMAL:
				auth = "normal"
			case operator.USER_AUTHORITY_ROOT:
				auth = "admin"
			}
			r = append(r, []string{rto.UserName, rto.Email, auth})
		}
		return
	}
	if l < 3 {
		err = fmt.Errorf("Command syntax error.")
		return
	}
	switch split[1] {
	case "delete":
		// user delete 'user_name'
		errs = o.operator.UserDel(split[2])
		err = errs.IsError()
	case "add":
		err = o.execUserAdd(split, l)
	default:
		err = fmt.Errorf("Command syntax error.")
	}
	return
}

// user的处理user add，从split[2]开始
func (o *OperatorT) execUserAdd(split []string, l int) (err error) {
	// user add 'user_name' for {drule|admin|root|normal} with password 'password' [and email 'email']
	if l < 8 {
		err = fmt.Errorf("Command syntax error.")
		return
	}
	username := split[2]
	password := split[7]

	var auth operator.UserAuthority
	switch split[4] {
	case "drule":
		auth = operator.USER_AUTHORITY_DRULE
	case "admin":
		auth = operator.USER_AUTHORITY_ROOT
	case "root":
		auth = operator.USER_AUTHORITY_ROOT
	case "normal":
		auth = operator.USER_AUTHORITY_NORMAL
	default:
		err = fmt.Errorf("Command syntax error.")
		return
	}
	var email string
	if l == 11 {
		email = split[10]
	}
	errs := o.operator.UserAdd(username, password, email, auth)
	err = errs.IsError()
	return
}
