package egitlab

type config struct {
	Token   string // 用于 调用gitlab api的token
	BaseUrl string // gitlab api base url
}

func DefaultConfig() *config {
	return &config{
		BaseUrl: "http://127.0.0.1:9000/api/v4",
	}
}
