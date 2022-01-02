package conf

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

// App config
type Config struct {
	Server struct {
		Port                     string `yaml:"port"`
		Host                     string `yaml:"host"`
		Prefix                   string `yaml:"prefix"`
		JWTAccessSecret          string `yaml:"jwt_access_secret"`
		JWTRefreshSecret         string `yaml:"jwt_refresh_secret"`
		TransactionsPipeCapacity uint16 `yaml:"transactions_pipe_cap"`
	} `yaml:"server"`
	WalletDatabase struct {
		Host              string `yaml:"host"`
		Port              string `yaml:"port"`
		DBName            string `yaml:"db_name"`
		WalletsTable      string `yaml:"wallets_table"`
		TransactionsTable string `yaml:"transactions_table"`
		Username          string `yaml:"user"`
		Password          string `yaml:"password"`
	} `yaml:"wallet_database"`
	Kafka struct {
		NewOrdersTopic      string   `yaml:"new_orders_topic"`
		RejectedOrdersTopic string   `yaml:"rejected_orders_topic"`
		GroupID             string   `yaml:"group_id"`
		Brokers             []string `yaml:"brokers"`
		ExternalClientsPort uint16   `yaml:"external_clients_port"`
		InternalClientsPort uint16   `yaml:"internal_clients_port"`
		SendMsgTimeout      uint8    `default:"5" yaml:"send_msg_timeout"`
		ConsumeLoopTick     uint16   `default:"500" yaml:"consume_loop_tick"`
	} `yaml:"kafka"`
	Logger struct {
		LogLevel string `yaml:"log_level"`
	} `yaml:"logger"`
}

func New() *Config {
	f, err := os.Open("./config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg Config

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	if err := defaults.Set(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func (c *Config) RegistryDatabaseURI() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.WalletDatabase.Username,
		c.WalletDatabase.Password,
		c.WalletDatabase.Host,
		c.WalletDatabase.Port,
		c.RegistryDatabaseDBName(),
	)
}

func (c *Config) RegistryDatabaseDBName() string {
	return c.WalletDatabase.DBName
}
