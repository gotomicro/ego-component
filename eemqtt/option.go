package eemqtt

import "time"

type Option func(c *Container)

func WithServerURL(serverURL string) Option {
	return func(c *Container) {
		c.config.ServerURL = serverURL
	}
}

func WithClientID(clientID string) Option {
	return func(c *Container) {
		c.config.ClientID = clientID
	}
}

func WithKeepAlive(keepAlive uint16) Option {
	return func(c *Container) {
		c.config.KeepAlive = keepAlive
	}
}

func WithConnectRetryDelay(connectRetryDelay time.Duration) Option {
	return func(c *Container) {
		c.config.ConnectRetryDelay = connectRetryDelay
	}
}

func WithConnectTimeout(connectTimeout time.Duration) Option {
	return func(c *Container) {
		c.config.ConnectTimeout = connectTimeout
	}
}

func WithSubscribeTopics(subscribeTopics map[string]subscribeTopic) Option {
	return func(c *Container) {
		c.config.SubscribeTopics = subscribeTopics
	}
}

func WithUsername(username string) Option {
	return func(c *Container) {
		c.config.Username = username
	}
}

func WithPassword(password string) Option {
	return func(c *Container) {
		c.config.Password = password
	}
}
