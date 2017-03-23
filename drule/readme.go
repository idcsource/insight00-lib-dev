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
// *Transaction.Prepare()方法的作用一样，都是对输入的“角色”请求独占。
//
// DRule
//
// “分布式统治者”提供了构建分布式“角色”控制服务器的功能。
//
// Operator
//
// “操作者”是DRule的客户端。
package drule
