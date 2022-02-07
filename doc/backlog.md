### Breaker任务列表


### TODO List
- [x] 构建Working Pool
- [x] 莫名其妙断掉的问题->proxy conn 阻塞
- [x] 支持断线重连(心跳机制)
- [ ] KCP增强(弱网环境下传输效率提升明显，但是会有一些额外的流量消耗)
- [ ] 负载均衡(frps)
- [ ] 加密与压缩(加密算法采用 aes-128-cfb，压缩算法采用 snappy)
- [ ] TLS 协议加密(与加密与压缩不同，为了防止中间人攻击)
- [ ] TCP 多路复用(减少文件占用符的使用)
- [x] 配置校验指令,check 指令
- [ ] http_proxy,静态代理
- [ ] static_file ,HTTP 服务查看指定的目录下的文件
- [ ] 客户端的代理限速
- [x] 代码重构，目前不优雅
  * 添加session对象，让working连接添加到master session中
  * 实现middleware
  * 实现router
