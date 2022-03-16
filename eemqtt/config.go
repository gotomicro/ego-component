package eemqtt

import (
	"time"
)

/**
  待完善：
    支持 TLS 方式连接
    WebSocket 自定义连接方式
*/
type config struct {
	ServerURL         string                    `json:"serverURL" toml:"serverURL"`                 //连接地址  tcp://host:1883  ws://mosquitto:80
	Username          string                    `json:"username" toml:"username"`                   //用户名
	Password          string                    `json:"password" toml:"password"`                   //密码
	ClientID          string                    `json:"clientID" toml:"clientID"`                   //客户端标识
	KeepAlive         uint16                    `json:"keepAlive" toml:"keepAlive"`                 //默认值 30
	ConnectRetryDelay time.Duration             `json:"connectRetryDelay" toml:"connectRetryDelay"` //default 10s
	ConnectTimeout    time.Duration             `json:"connectTimeout" toml:"connectTimeout"`       //default 10s
	SubscribeTopics   map[string]subscribeTopic `json:"subscribeTopics" toml:"subscribeTopics"`     //连接后自动订阅主题
	Debug             bool                      `json:"debug" toml:"debug"`                         // Debug 是否开启debug模式
}

//订阅主题
type subscribeTopic struct {
	Topic string `json:"topic" toml:"topic"`
	Qos   byte   `json:"qos" toml:"qos"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Debug:             false,
		KeepAlive:         30,
		ConnectRetryDelay: time.Second * 10,
		ConnectTimeout:    time.Second * 10,
		ClientID:          "",
		SubscribeTopics:   make(map[string]subscribeTopic),
	}
}

//如果想使用 TLS 连接，可以如下设置：
//func NewTlsConfig() *tls.Config {
//	certpool := x509.NewCertPool()
//	ca, err := ioutil.ReadFile("ca.pem")
//	if err != nil {
//		log.Fatalln(err.Error())
//	}
//	certpool.AppendCertsFromPEM(ca)
//	// Import client certificate/key pair
//	clientKeyPair, err := tls.LoadX509KeyPair("client-crt.pem", "client-key.pem")
//	if err != nil {
//		panic(err)
//	}
//	return &tls.Config{
//		RootCAs:            certpool,
//		ClientAuth:         tls.NoClientCert,
//		ClientCAs:          nil,
//		InsecureSkipVerify: true,
//		Certificates:       []tls.Certificate{clientKeyPair},
//	}
//}

//如果不设置客户端证书，可以如下设置：
//func NewTlsConfigs() *tls.Config {
//	certpool := x509.NewCertPool()
//	ca, err := ioutil.ReadFile("ca.pem")
//	if err != nil {
//		log.Fatalln(err.Error())
//	}
//	certpool.AppendCertsFromPEM(ca)
//	return &tls.Config{
//		RootCAs:            certpool,
//		ClientAuth:         tls.NoClientCert,
//		ClientCAs:          nil,
//		InsecureSkipVerify: true,
//	}
//}
