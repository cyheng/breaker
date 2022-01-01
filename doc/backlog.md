### Breaker任务列表

~~1. 反向代理tcp协议~~
2. 断线重连
3. web端修改配置信息
4. 统计流量(流量就是字节数)
5. 可靠udp？
6. 支持http反向代理
~~7. 通过init函数注册factory~~
8. 协议
    * 消息类型,1byte
    * 长度,int64
    * 消息body
 
   

### TODO List
1. 重构为Master,Worker 流程(已完成)
2. 构建Working Pool
3. 优雅地判断conn是否关闭
4. 使用pkg/errors
   1. 应用程序中出现错误时，使用 errors.New  或者 errors.Errorf  返回错误
   2. 调用应用程序的其他函数出现错误，请直接返回，如果需要携带信息，请使用 errors.WithMessage
   3. 调用其他库（标准库、企业公共库、开源第三方库等）获取到错误时，请使用 errors.Wrap
   4. 禁止每个出错的地方都打日志，只需要在进程的最开始的地方使用 %+v  进行统一打印


