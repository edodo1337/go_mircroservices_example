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
		Port                  string `default:"8000" yaml:"port"`
		Host                  string `default:"localhost" yaml:"host"`
		Prefix                string `yaml:"prefix"`
		JWTAccessSecret       string `yaml:"jwt_access_secret"`
		JWTRefreshSecret      string `yaml:"jwt_refresh_secret"`
		NewOrdersPipeCapacity uint16 `yaml:"new_orders_pipe_cap"`
	} `yaml:"server"`
	RegistryDatabase struct {
		Host            string `default:"localhost" yaml:"host"`
		Port            string `default:"5432" yaml:"port"`
		DBName          string `yaml:"db_name"`
		OrdersTable     string `yaml:"orders_table"`
		OrderItemsTable string `yaml:"order_items_table"`
		ProductsTable   string `yaml:"products_table"`
		Username        string `yaml:"user"`
		Password        string `yaml:"password"`
	} `yaml:"registry_database"`
	Kafka struct {
		NewOrdersTopic      string   `yaml:"new_orders_topic"`
		RejectedOrdersTopic string   `yaml:"rejected_orders_topic"`
		SuccessTopic        string   `yaml:"success_topic"`
		GroupID             string   `default:"registry" yaml:"group_id"`
		Brokers             []string `yaml:"brokers"`
		ExternalClientsPort uint16   `yaml:"external_clients_port"`
		InternalClientsPort uint16   `yaml:"internal_clients_port"`
		MaxWait             uint8    `default:"200" yaml:"max_wait"`
		SendMsgTimeout      uint8    `default:"5" yaml:"send_msg_timeout"`
		ConsumeLoopTick     uint16   `default:"500" yaml:"consume_loop_tick"`
	} `yaml:"kafka"`
	Logger struct {
		LogLevel string `default:"INFO" yaml:"log_level"`
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
		c.RegistryDatabase.Username,
		c.RegistryDatabase.Password,
		c.RegistryDatabase.Host,
		c.RegistryDatabase.Port,
		// Зачем нужна эта функция?
		c.RegistryDatabaseDBName(),
	)
}

func (c *Config) RegistryDatabaseDBName() string {
	return c.RegistryDatabase.DBName
}
