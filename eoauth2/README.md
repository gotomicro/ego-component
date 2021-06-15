# eoauth2 组件使用指南
## 简介 
- 根据开源项目``github.com/RangelReale/osin``做了改造
- 支持http oauth2 server
- 支持grpc oauth2 server
- 方便改造为多客户端的sso服务


### 单点登录系统
* 客户端服务端写入state信息，并生成url
* 通过浏览器请求sso，sso返回给浏览器code信息
* code信息回传给客户端的服务端，请求sso服务，获得token
* 将token存入到浏览器的http only的cookie里
* 所有接口都可以通过该token，grpc获取用户信息

## 流程
### Authorize
* 写入authorize表，生成code码
* 写入authorize过期时间

### Token
* 第一次保存token
    * save access token
    * remove authorization token
* 刷新token
    * save access token
    * remove authorization token
    * remove previous access token
    
