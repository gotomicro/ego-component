# egitlab
封装gitlab api 调用

```go
type config struct {
	Token   string // 用于 调用gitlab api的token
	BaseUrl string // gitlab api base url
}
token: 若token所属的gitlab用户被禁用，请使用其他用户创建token,并替换token配置
```