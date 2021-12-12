# frps源码阅读笔记

### 思考的问题

1.它是怎么使用kcp的

2.xtcp的流程

3.如何统计流量
  其实就是流量的字节数量统计
4.如何进行流量压缩和加密解密

5.xtcp的流程

6.如何实现日志自动分日期的
  用了beego的日志库

7.context是如何使用的

### 启动流程

通过cobra读取命令行参数，其参数定义在var中（private）

定义一个ServerCommonConf,用于记录当前参数（无论来源命令行还是ini文件）

基于config，RunServer，是一个普通的function

日志用的是BeeLogger,目的暂时不知道

RunServer函数里面会新建一个server(在server package 下)

调用它的run方法



错误处理：一直return error,直到root中的runE(仅在new server之前)



server.Run,藏了几个协程

- NatHoleController
- HandleListener（kcp）
- HandleListener(websocket)
- HandleListener(tls)
- HandleListener(普通请求)

### HandleListener流程

1. accept一个连接，传xlog和context(xlog和beego log的区别？答：xlog通过context保证了线程安全,为什么context线程安全？不可变,context在main中通过context.Background()初始化，不可变，而且都是同一个引用)

   context.TODO(重构的时候用到，还没传context),context.Background()

   反模式->context的参数理应作为函数参数的时候

   类似ThreadLocal

2. 检查TLS(why?)

3. 新建一个goroutine,处理这个连接

4. 重复第一步

### goroutine处理流程

1. 检查是否启用tcp 多路复用

2. 没有，直接处理（也是我平常编程的时候的用法）

3. 有，转成stream 启用一个协程处理（它是怎么做到的？）

   

### handleconnection流程

1. conn设置read超时时间,读数据
2. conn不设置超时时间（为什么）
3. 判断消息类型
   1. login: Login-》RegisterControl
   2. newWorkConn：RegisterWorkConn
   3. newVisitorConn：RegisterVisitorConn


### 监听用户请求
在新建service的时候，通过bind port监听
然后，通过监听的数据转到chan 中


### 如何优雅的监听请求
