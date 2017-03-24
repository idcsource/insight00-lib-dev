// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ]
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// Use of this source code is governed by GNU LGPL v3 license

// DRule可以理解为Distributed Rule，也就是分布式统治者。
//
// DRule是在drcm之后出现的东西，它可能会复制粘贴许多drcm的代码，不过它可能不再拘泥一些事情。
// 比如它虽然实现了rolesio.RolesInOutManager的接口，但却不建议使用这个接口。而更重要的是它将增加对事务（Transaction）的支持。
// 我深知，DRule代表了这个项目越走越远，越走越背离初衷。但既然它不是一个严肃的项目，也没有人知道初衷是什么，那就这样吧！
//
// TRule
//
// “事务统治者”为这个drule包的最底层，因为它的存在才有了DRule。它提供了对事务的支持。
//
// 目前TRule支持简单的事务操作，包括事务的创建（Begin）、执行（Commit）、回滚（Rollback）。
// TRule彻底改变了之前其他与“角色”存储有关的实现方式，放弃了全局的内存缓存，不再依靠计算缓存个数或调用ToStore()来实现将缓存保存进硬盘的机制。
// 在多种不同粒度的锁的支持下，在“角色”的缓存的事务调度与排队机制下，“角色”的缓存与硬盘存储间达成了完全的动态按需机制。
// 不再需要操作的“角色”将积极保存回硬盘，以更好的保证数据安全并释放内存空间。
//
// TRule的初始化只需要提供*hardstore.HardStore实例作为面向硬盘的底层存储，NewTRule函数将返回*TRule的实例。之后则可以按照rolesio.RolesInOutManager接口定义的方法对“角色”进行操作（使用ToStore()方法则会弹出错误）。
// 但并不建议直接使用*TRule来操作“角色”，这么做并不高效也不安全，因为所有操作在底层都是走创建事务、对“角色”操作、提交事务的流程，无法利用缓存等特性。
//
// 建议使用*TRule.Begin()方法返回*Transaction实例在事务模式内执行对“角色”的各种操作。*Transaction依然实现了rolesio.RolesInOutManager接口定义的方法，同样如果使用ToStore()方法会提示错误。
// 在事务模式内，任何对“角色”的操作都是在缓存下完成。目前写状态是独占锁，读状态默认为“脏读”。通过*Transaction.Commit()对事务进行提交，通过*Transaction.Rollback()对所有更改进行回滚撤销。
// 如果使用*TRule.Prepare(roleids ...string)方法，将在创建事务时对所有请求的“角色”（roleids）首先进行独占，在这种状态下一定要注意避免死锁情况的发生。
// *Transaction.LockRoles()方法的作用一样，都是对输入的“角色”请求独占。
//
// DRule
//
// “分布式统治者”提供了构建分布式“角色”控制服务器的功能。不同于drcm的ZrStorage，DRule只提供服务端功能，也就是slave和master，而不提供own模式。
//
// DRule需要提供一个*cpool.Block类型的配置信息，示例如下。
//	{drule}
//
//	[main]
//	# mode可以是own、master
//	mode = master
//	# 自己的标识名称
//	selfname = Master0001
//	# 作为服务监听的端口
//	port = 9999
//	# 自己的身份验证码，访问者需要提供这串代码来获得操作权限
//	code = dfadfa3gd3567
//	# slave配置的名字，用逗号分割，master必须
//	slave = s001,s002
//
//	# local是本地的存储设置，包含缓存和底层HardStore的设置
//	[local]
//	# 本地的存储位置
//	path = /pathname/
//	# 本地存储的路径层深
//	path_deep = 2
//
//	[s001]
//	# 路由方案，角色名的首字母的列表，设置表明这些字母开头的角色将交由这台slave处理
//	control = a,b,c,d,e,f,1,2,3
//	# master与slave的连接数
//	conn_num = 5
//	# slave的身份验证码
//	code = 2g2gh9t40hg2h2g
//	# slave的地址和端口
//	address = 192.168.1.101:11111
//
//	[s002]
//	# 与s001重复的字母将规整为镜像，执行读操作的时候随机选取，进行写操作的时候同时执行
//	control = 0,1,2,3,4,5,6,7,8,9
//	conn_num = 5
//	code = 9ghfg290ghg
//	address = 192.168.1.102:11111
//
// Operator
//
// “操作者”是DRule的客户端。
//
// Operator在新建时需要提供一台DRule的相关信息，包括地址、连接数、身份验证码。NewOperator()函数中的selfname为自身标识名。之后如果需要进行镜像管理则可以通过AddServer()方法添加。
//
// Operator对所有镜像同等对待，没有路由规则，读操作的时候随即选取，写操作的时候同时执行。
//
// Operator提供与TRule几乎完全一样的方法。但在写代码的时候并没有再分事务类型，故如果你在事务内执行Begin()或在事务外执行Commit()都会弹出错误。
package drule
