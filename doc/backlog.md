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
9. 交互模型，参考了frp和nps，最终选择了frp的参考模型
   * frp模型,创建了proxy之后，服务端就发了几个指令要求连接(client dial server),加入到worker pool中
   * nps模型，创建了proxy之后，服务端dial client



### TODO List
- [x] 重构为Master,Worker 流程
- [x] 构建Working Pool
- [x] 莫名其妙断掉的问题->proxy conn 阻塞
- [ ] 支持http
- [ ] 支持断线重连(心跳机制)
- [ ] hot reload 配置文件,reload 指令
- [ ] 配置校验指令,check 指令
- [ ] 统计信息->通过prometheus
- [ ] 负载均衡(frps)
- [ ] 加密与压缩
- [ ] TLS 协议加密(与加密与压缩不同，为了防止中间人攻击)
- [ ] TCP 多路复用(减少文件占用符的使用)
- [ ] KCP增强(弱网环境下传输效率提升明显，但是会有一些额外的流量消耗)
- [ ] http_proxy,静态代理
- [ ] static_file ,HTTP 服务查看指定的目录下的文件
- [ ] 客户端的代理限速
- [x] 代码重构，目前不优雅
  * 添加session对象，让working连接添加到master session中
  * 实现middleware
  * 实现router
