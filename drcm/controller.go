// Copyright 2016
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Normal Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

package drcm

import (
	"sync"
	"fmt"
	"errors"
	
	"github.com/idcsource/Insight-0-0-lib/roles"
	"github.com/idcsource/Insight-0-0-lib/rolesio"
	"github.com/idcsource/Insight-0-0-lib/ilogs"
	"github.com/idcsource/Insight-0-0-lib/nst"
)

// 新建Controller
func NewController (ro rolesio.RolesInOutManager, max int64, logs *ilogs.Logs) *Controller {
	c := &Controller{
		readWrite: ro,
		rolesCache : make(map[string]roles.Roleer),
		deleteCache: make([]string,0),
		cacheMax : max,
		rolesCount : 0,
		cacheIsFull : make(chan bool),
		logs : logs,
		lock: new(sync.RWMutex),
		checkCacheNumOn: false,
	};
	if c.cacheMax > 0 {
		go c.runtimeStore();
	}
	return c;
}

// 注册一个角色
func (c *Controller) RegRole (role roles.Roleer) err {
	id := role.ReturnId();
	if len(id) == 0 {
		return errors.New("drcm: [Controller]RegRole: The role have no ID be set.");
	}
	rc.lock.RLock();
	defer rc.lock.RUnlock();
	go rc.checkCacheNum();
	if _, find := rc.rolesCache[id]; find == true {
		return errors.New("drcm: [Controller]RegRole: The role " + id + " already have .");
	}
	if c.checkDel(id) == true {
		c.deleteCache = c.sliceDel(c.deleteCache, id);
	}
	c.rolesCache[id] = role;
	c.rolesCount++;
	return nil;
}

// 获取一个角色
func (c *Controller) GetRole (id string) (roles.Roleer, error) {
	c.lock.RLock();
	defer c.lock.RUnlock();
	go c.checkCacheNum();
	var find bool;
	if find := rc.checkDelById(id); find == true {
		err := errors.New("drcm: [Controller]GetRole: The role " + id + " can't be found, maybe it be deleted !");
		return nil, err;
	}
	role, find := rc.rolesCache[id];
	if find == true {
		return role, nil;
	} else {
		role, err := c.readWrite.ReadRole(id);
		if err != nil {
			return nil , err;
		}
		c.rolesCache[id] = role;
		c.rolesCount++;
		return role, nil;
	}
}

// 删除一个角色，但这里只是从缓存中删除并标记为删除，真正的删除需要等到tostore时才会进行。
// 此方法永远都会返回nil，也就是只要你给予了，它都将在tostore时提交给永久存储去删除。
func (c *Controller) DeleteRole (id string) (err error) {
	rc.lock.Lock();
	if rc.checkDel(id) == true {
		return;
	}
	role, find := c.rolesCache[id];
	if find == true {
		delete(c.rolesCache, id);
	}else{
		return;
	}
	role.SetDelete(true);
	rc.deleteCache = append(c.deleteCache, id);
	role = nil;
	return;
}

// 向role中注册father关系，如果role之前有father关系，则返回错误。
func (c *Controller) SetFather (id, father string) (err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	
	if rc.checkDel(id) == true {
		return errors.New("drcm: [Controller]SetFather: Cant set selationship for a deleted role .");
	}
	
	go rc.checkCacheNum();
	err = c.checkCache(id);
	if err != nil {
		return;
	}
	
	c.rolesCache[id].SetFather(father);
	return nil;
}

// 从角色中获取父角色的id
func (c *Controller) GetFather (id string) (father string, err error) {
	rc.lock.Lock();
	defer rc.lock.Unlock();
	
	if rc.checkDel(id) == true {
		err = errors.New("drcm: [Controller]GetFather: Cant set selationship for a deleted role .");
		return;
	}
	
	go rc.checkCacheNum();
	err = c.checkCache(id);
	if err != nil {
		return;
	}
	father = c.rolesCache[id].GetFather();
	return nil;
}

// 将目前载入的角色关系全部保存起来，但不会删除缓存
func (c *Controller) ToStore ()  {
	c.toStore();
}

// 将目前载入的角色关系全部保存起来，内部使用，无锁
func (c *Controller) toStore ()  {
	c.checkCacheNumOn = true;
	c.lock.Lock();
	defer c.deferCheckCacheNumOn();
	defer c.lock.Unlock();
	// 检查在children和friends关系里面是否有被delete的成分，如果有则在保存前剔除
	for _, role := range c.rolesCache {
		children := role.GetChildren();
		delchild := make([]string,0);
		for _, name := range children {
			bedel := c.checkDelById(name);
			if bedel == true {
				delchild = append(delchild, name);
			}
		}
		for _, name := range delchild {
			role.DeleteChild(name);
		}
		friends := role.GetFriends();
		delfriend := make([]string,0);
		for name, _ := range friends {
			bedel := c.checkDelById(name);
			if bedel == true {
				delfriend = append(delfriend, name);
			}
		}
		for _, name := range delfriend {
			role.DeleteFriend(name);
		}
	}
	
	// 开始调用永久存储方法
	// 首先就是删除要删除的东西
	for _, id := range c.deleteCache {
		err := c.readWrite.DeleteRole(id);
		if err != nil && c.logs != nil {
			c.logs.ErrLog(err);
		}
	}
	c.deleteCache = make([]string, 0);		// 重置删除缓存
	// 然后写入每一个Role的本体，以及各种关系
	for _, role := range c.rolesCache {
		err := c.readWrite.StoreRole(role);		// 写入本体
		if err != nil && c.logs != nil {
			c.logs.ErrLog(err);
		}
	}
	return;
}

// 运行时存储，当缓存超过时执行
func (c *Controller) runtimeStore () {
	for {
		full := <-c.cacheIsFull;
		if full == true {
			c.toStore();
			c.rolesCache = make(map[string]roles.Roleer);
			c.rolesCount = 0;
		}
	}
}

// 检查缓存数，如果超出则执行运行时保存
func (c *Controller) checkCacheNum () {
	if c.cacheMax > 0 && c.rolesCount >= c.cacheMax && c.checkCacheNumOn == false {
		c.cacheIsFull <- true;
	}
}

func (c *Controller) deferCheckCacheNumOn() {
	c.checkCacheNumOn = false;
}

// 检查缓存，如果缓存里没有，则纳入到缓存里
func (c *Controller) checkCache (id string) error {
	_, find := c.rolesCache[id];
	if find == false {
		role, err := c.GetRole(id);
		if err != nil {
			return err;
		}
		c.rolesCache[id] = role;
		c.rolesCount++;
	}
}

// 查看是否被标记删除，标记删除则返回true。
func (c *Controller) checkDel (id string) bool {
	del := c.checkDelById(id);
	if del == true {
		return del;
	}
	del = role.ReturnDelete();
	return del;
}

func (c *Controller) checkDelById (id string) bool {
	for _, v := range c.deleteCache {
		if v == id {
			return true;
		}
	}
	return false;
}

// 切片中项目删除
func (c *Controller) sliceDel (sl []string, text string) []string {
	for k, v := range sl {
    if v == "ab" {
        kk := k + 1
        sl = append(s[:k], s[kk:]...)
    }
    return sl;
}
}
