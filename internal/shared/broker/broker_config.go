package broker

type BrokerConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

func NewBrokerConfig(host, port, user, password, vhost string) *BrokerConfig {
	return &BrokerConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		VHost:    vhost,
	}
}

func (rc *BrokerConfig) GetConnectionString() string {
	return "amqp://" + rc.User + ":" + rc.Password + "@" + rc.Host + ":" +
		rc.Port + "/%2f"
}

// "amqp://wheres-my-pizza:wheresmypizza@localhost:5672/%2F"
