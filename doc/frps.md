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

## read
1. ReqWorkConn
2. NewProxyResponse
3. Pong(定时)

### on userconnect
4. ReqWorkConn

## write
1. newMaster
2. Ping(定时)
   ###on userconnect
不做任何操作
 

### server 流程:
## read
1. newMaster
### on userconnect
不做任何操作

## write
1. ReqWorkConn
2. NewProxyResponse
### on userconnect
1. ReqWorkConn