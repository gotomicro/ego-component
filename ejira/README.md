# Jira API组件使用指南
Jira Server REST API 
https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0

# 安装
私有化部署下载地址：https://www.atlassian.com/software/jira/download
安装过程：https://confluence.atlassian.com/adminjiraserver071/installing-jira-applications-on-linux-from-archive-file-855475657.html

# 常见问题
1. BasicAuth认证失败，提示Basic Authentication Failure - Reason : AUTHENTICATION_DENIED
答：去系统管理界面将 Maximum Authentication Attempts Allowed 改大一点

# 组件
* [ x ] [User](./service_user.go)
* [ x ] [Project](./service_project.go)
* [ x ] [Status](./service_status.go)
* [ x ] [Version](./service_version.go)
* [ x ] [Priority](./service_priority.go)
* [ x ] [Resolution](./service_resolution.go)
* [ x ] [IssueLinkType](./service_issuelinktype.go)
* [ x ] [Issue](./service_issue.go)

## 快速上手
使用样例可参考 [example](examples/main.go)