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
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个分布式统治者
//
// 自身的名字，工作模式，trule，日志
func NewDRule(selfname string, mode OperateMode, t *trule.TRule, log *ilogs.Logs) (d *DRule, err error) {
	d = &DRule{
		selfname:  selfname,
		dmode:     mode,
		trule:     t,
		closed:    true,
		operators: make(map[string]*operator.Operator),
		areas:     make(map[string]*AreasDRule),
		loginuser: make(map[string]*loginUser),
		logs:      log,
	}
	var have bool
	// 查看又无自己的区域，没有就建立
	have = t.AreaExist(INSIDE_DMZ)
	if have == false {
		err = t.AreaInit(INSIDE_DMZ)
		if err != nil {
			err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
			return
		}
	}
	// 查看有无自己的根管理员，没有就建立
	root_user_id := USER_PREFIX + ROOT_USER
	have = t.ExistRole(INSIDE_DMZ, root_user_id)
	if have == false {
		root_user := &DRuleUser{
			UserName:  ROOT_USER,
			Password:  random.GetSha1Sum(ROOT_USER_PASSWORD),
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
	// 查看有无自己的其他两个根，没有就建立
	have = t.ExistRole(INSIDE_DMZ, OPERATOR_ROOT)
	if have == false {
		operator_root := &DRuleOperatorRoot{}
		operator_root.New(OPERATOR_ROOT)
		err = t.StoreRole(INSIDE_DMZ, operator_root)
		if err != nil {
			err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
			return
		}
	}
	have = t.ExistRole(INSIDE_DMZ, AREA_DRULE_ROOT)
	if have == false {
		area_root := &AreasDRuleRoot{}
		area_root.New(AREA_DRULE_ROOT)
		err = t.StoreRole(INSIDE_DMZ, area_root)
		if err != nil {
			err = fmt.Errorf("drule[DRule]NewDRule: %v", err)
			return
		}
	}
	// 此时的工作还是在停止中
	return
}

// 启动服务，对工作模式等进行管理
func (d *DRule) Start() (err error) {
	// 准备区域数据
	areas, err := d.trule.ReadChildren(INSIDE_DMZ, AREA_DRULE_ROOT)
	if err != nil {
		err = fmt.Errorf("drule[DRule]Start: %v", err)
		return
	}
	for _, areaid := range areas {
		arearole := &AreasDRule{}
		err = d.trule.ReadRole(INSIDE_DMZ, areaid, arearole)
		if err != nil {
			err = fmt.Errorf("drule[DRule]Start: %v", err)
			return
		}
		d.areas[arearole.AreaName] = arearole
	}
	// 准备路由数据
	if d.dmode == OPERATE_MODE_MASTER {
		op_sets, err := d.trule.ReadChildren(INSIDE_DMZ, OPERATOR_ROOT)
		if err != nil {
			err = fmt.Errorf("drule[DRule]Start: %v", err)
			return err
		}
		for _, op_set_id := range op_sets {
			op_set_r := &DRuleOperator{}
			err = d.trule.ReadRole(INSIDE_DMZ, op_set_id, op_set_r)
			if err != nil {
				err = fmt.Errorf("drule[DRule]Start: %v", err)
				return err
			}
			var op *operator.Operator
			if op_set_r.TLS == false {
				op, err = operator.NewOperator(op_set_r.Name, op_set_r.Address, op_set_r.ConnNum, op_set_r.Username, op_set_r.Password, d.logs)
				if err != nil {
					err = fmt.Errorf("drule[DRule]Start: %v", err)
					return err
				}
			} else {
				op, err = operator.NewOperatorTLS(op_set_r.Name, op_set_r.Address, op_set_r.ConnNum, op_set_r.Username, op_set_r.Password, d.logs)
				if err != nil {
					err = fmt.Errorf("drule[DRule]Start: %v", err)
					return err
				}
			}
			d.operators[op_set_r.Name] = op
		}
	}
	d.closed = false
	return
}

// 关闭
func (d *DRule) Close() {
	d.closed = true
}
