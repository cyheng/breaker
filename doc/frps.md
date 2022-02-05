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

8.HandleListener 和 control的区别
- HandleListener 是处理客户端的Dial
- 
 
### client 流程:
client会建立两个连接login(用户发送control)和ReqWorkConn(用于转发流量)
login-> dial server(能不能省略这步？)
ReqWorkConn-> dial server

## 断线重连

frpc是通过keepControllerWorking 来保持断线重连

##  hot reload 配置文件,reload 指令


##  配置校验指令,check 指令