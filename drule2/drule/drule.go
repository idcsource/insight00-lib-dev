// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
)

// 新建一个分布式统治者
//
// 自身的名字，工作模式，trule，日志
func NewDRule(selfname string, mode OperateMode, t *trule.TRule, log *ilogs.Logs) (d *DRule, err error) {
	d = &DRule{
		selfname:  selfname,
		dmode:     mode,
		trule:     t,
		closed:    false,
		operators: make(map[string]*operator.Operator),
		areas:     make(map[string]*AreasDRule),
		loginuser: make(map[string]*loginUser),
		logs:      log,
	}
	// 查看又无自己的区域，没有就建立
	have = t.AreaExist(INSIDE_DMZ)
	if have == false {
		err = t.AreaInit(INSIDE_DMZ)
		if err != nil {
			err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
			return
		}
	}
	// 查看又无自己的根管理员，没有就建立
	root_user_id := USER_PREFIX + ROOT_USER
	have = t.ExistRole(INSIDE_DMZ, root_user_id)
	if have == false {
		root_user := &DRuleUser{
			UserName:  ROOT_USER,
			Password:  ROOT_USER_PASSWORD,
			Email:     "",
			Authority: operator.USER_AUTHORITY_ROOT,
			WRable:    make(map[string]bool),
		}
		root_user.New(root_user_id)
		err = t.StoreRole(INSIDE_DMZ, root_user)
		if err != nil {
			err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
			return
		}
	}
	// 查看又无自己的其他两个根，没有就建立
	// 查看工作模式并作相应处理
	return
}
