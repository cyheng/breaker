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
- [x] 重构为Master,Worker 流程
- [x] 构建Working Pool
- [ ] 支持http
- [ ] 支持断线重连(心跳机制)
- [ ] 统计信息->通过prometheus
- [ ] KCP增强
- [ ] Websocket 
- [ ] 测试用例
- [ ] 代码重构，目前不优雅
