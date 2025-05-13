package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type KafkaConfig struct {
	Brokers    []string         `mapstructure:"brokers"`
	Topics     KafkaTopics      `mapstructure:"topics"`
	Consumer   ConsumerConfig   `mapstructure:"consumer"`
	Producer   ProducerConfig   `mapstructure:"producer"`
	DeadLetter DeadLetterConfig `mapstructure:"dead_letter"`
}

type KafkaTopics struct {
	PromoCreated       string `mapstructure:"promo_created"`
	PromoUpdated       string `mapstructure:"promo_updated"`
	PromoDeleted       string `mapstructure:"promo_deleted"`
	PromoViewed        string `mapstructure:"promo_viewed"`
	PromoClicked       string `mapstructure:"promo_clicked"`
	PromoLiked         string `mapstructure:"promo_liked"`
	UserRegistered     string `mapstructure:"user_registered"`
	UserProfileUpdated string `mapstructure:"user_profile_updated"`
}

type ConsumerConfig struct {
	GroupID         string        `mapstructure:"group_id"`
	AutoOffsetReset string        `mapstructure:"auto_offset_reset"`
	MaxRetries      int           `mapstructure:"max_retries"`
	RetryBackoff    time.Duration `mapstructure:"retry_backoff"`
}

type ProducerConfig struct {
	Acks         string        `mapstructure:"acks"`
	Retries      int           `mapstructure:"retries"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
	MaxInFlight  int           `mapstructure:"max_in_flight"`
}

type DeadLetterConfig struct {
	Topic      string        `mapstructure:"topic"`
	MaxRetries int           `mapstructure:"max_retries"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Port      int    `mapstructure:"port"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
	Subsystem string `mapstructure:"subsystem"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
