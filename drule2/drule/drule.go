// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drule

import (
	"fmt"
	"time"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/trule"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/random"
)

// 新建一个分布式统治者
//
// 自身的名字，工作模式，trule，日志
func NewDRule(selfname string, mode operator.DRuleOperateMode, t *trule.TRule, log *ilogs.Logs) (d *DRule, err error) {
	d = &DRule{
		selfname:        selfname,
		dmode:           mode,
		trule:           t,
		closed:          true,
		operators:       make(map[string]*operator.Operator),
		areas:           make(map[string]*AreasRouter),
		loginuser:       make(map[string]*loginUser),
		transaction_map: make(map[string]*transactionMap),
		logs:            log,
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
		area_root := &AreasRouterRoot{}
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
	if d.closed == false {
		err = fmt.Errorf("drule[DRule]Start: DRule already started.")
		return
	}
	d.trule.Start()
	// 准备远端operator数据
	d.operators = make(map[string]*operator.Operator)
	if d.dmode == operator.DRULE_OPERATE_MODE_MASTER {
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
	// 准备区域路由数据
	d.areas = make(map[string]*AreasRouter)
	areas, err := d.trule.ReadChildren(INSIDE_DMZ, AREA_DRULE_ROOT)
	if err != nil {
		err = fmt.Errorf("drule[DRule]Start: %v", err)
		return
	}
	for _, areaid := range areas {
		arearole := &AreasRouter{}
		err = d.trule.ReadRole(INSIDE_DMZ, areaid, arearole)
		if err != nil {
			err = fmt.Errorf("drule[DRule]Start: %v", err)
			return
		}
		// 检查配置，查看路由项目中是否使用了不存在远端Operator
		if arearole.Mirror == true {
			for _, m := range arearole.Mirrors {
				if _, find := d.operators[m]; find == false {
					err = fmt.Errorf("drule[DRule]Start: Area '%v' can not find the remote DRule server operator '%v' set.", arearole.AreaName, m)
					return
				}
			}
		} else {
			for _, ms := range arearole.Chars {
				for _, m := range ms {
					if _, find := d.operators[m]; find == false {
						err = fmt.Errorf("drule[DRule]Start: Area '%v' can not find the remote DRule server operator '%v' set.", arearole.AreaName, m)
						return
					}
				}
			}
		}
		d.areas[arearole.AreaName] = arearole
	}
	d.closed = false
	return
}

// 暂停，管理级别的操作还可以进行，这个模式主要是用来调整比如路由策略
func (d *DRule) Pause() {
	d.closed = true
	//d.trule.Pause()
}

// 停止，真正的停止，没有远程命令可以恢复
func (d *DRule) Close() {
	d.closed = true
	d.trule.Pause()
}

// 查看用户是否登陆，如果登陆了就续期
func (d *DRule) checkUserLogin(username, unid string) (yes bool) {
	login, find := d.loginuser[username]
	// 找不到用户登陆信息
	if find == false {
		yes = false
		return
	}
	login_time, find := login.unid[unid]
	// 找不到unid
	if find == false {
		yes = false
		return
	}
	// 时间超时
	if login_time.Unix()+operator.USER_ALIVE_TIME > time.Now().Unix() {
		delete(login.unid, unid)
		// 如果没有活跃的登陆信息了，就把用户的这一条删掉
		if len(login.unid) == 0 {
			delete(d.loginuser, username)
		}
		yes = false
		return
	}
	// 续期
	d.loginuser[username].unid[unid] = time.Now()
	yes = true
	return
}

// 查看用户的管理权限
func (d *DRule) getUserAuthority(username, unid string) (authoriy operator.UserAuthority, login bool) {
	// 查看是否登录
	have := d.checkUserLogin(username, unid)
	if have == false {
		login = false
		return
	}
	authoriy = d.loginuser[username].authority
	login = true
	return
}

// 查看用户一般权限，也就是针对area的权限,wr为true则检查是否可写，否则为是否可读
func (d *DRule) checkUserNormalPower(username, areaname string, wr bool) (have bool) {
	// 找到loginuser中的权限项目
	wrable, find := d.loginuser[username].wrable[areaname]
	if find == false {
		// 找不到
		have = false
		return
	}
	// 只有读权限，想要写权限
	if wr == true || wrable == false {
		have = false
		return
	}
	have = true
	return
}

// 找到一个角色应该被保存在哪里，不是本地还返回所有需要发送的operator名称
func (d *DRule) getRolePosition(areaname, roleid string) (position RolePosition, os []string) {
	// slave模式下没有远程连接，所以肯定是本地
	if d.dmode == operator.DRULE_OPERATE_MODE_SLAVE {
		position = ROLE_POSITION_IN_LOCAL
		return
	}
	area_set, have := d.areas[areaname]
	// 找不到配置也是本地
	if have == false {
		position = ROLE_POSITION_IN_LOCAL
		return
	}
	// 镜像的话，远程
	if area_set.Mirror == true && len(area_set.Mirrors) > 0 {
		os = area_set.Mirrors
		position = ROLE_POSITION_IN_REMOTE
		return
	}
	// 非镜像的话，并且找到
	if area_set.Mirror == false {
		theChar := string(roleid[0])
		os, have = area_set.Chars[theChar]
		if have == true && len(os) > 0 {
			position = ROLE_POSITION_IN_REMOTE
			return
		}
	}
	// 其他任何情况
	position = ROLE_POSITION_IN_LOCAL
	return
}

// 全面获取是否有区域的执行权限，以及存在哪里
func (d *DRule) getAreaPowerAndRolePosition(username, areaname, roleid string, wr bool) (have bool, position RolePosition, os []string) {
	have = d.checkUserNormalPower(username, areaname, wr)
	if have == true {
		position, os = d.getRolePosition(areaname, roleid)
	}
	return
}
