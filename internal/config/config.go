package config

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Logging            bool
	LogPath            string
	Network            string
	Index              string
	Debug              bool
	Reindex            bool
	ReindexSize        uint64
	BulkIndexSize      uint64
	SoftForkBlockCycle int
	SoftForkQuorum     int
	Navcoind           NavcoindConfig
	ElasticSearch      ElasticSearchConfig
	ZeroMq             ZeroMqConfig
	Sentry             SentryConfig
	RabbitMq           RabbitMqConfig
}

type NavcoindConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Ssl      bool
}

type ElasticSearchConfig struct {
	Hosts       []string
	Sniff       bool
	HealthCheck bool
	Debug       bool
	Username    string
	Password    string
	MappingDir  string
}

type ZeroMqConfig struct {
	Address string
}

type RabbitMqConfig struct {
	User     string
	Password string
	Host     string
	Port     int
}

type SentryConfig struct {
	Active bool
	DSN    string
}

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Fatal("Unable to init config")
	}

	if Get().Debug {
		log.SetLevel(log.DebugLevel)
	}

	if Get().Sentry.Active {
		_ = raven.SetDSN(Get().Sentry.DSN)
	}

	initLogger()
}

func initLogger() {
	var logger *lumberjack.Logger
	if Get().Logging {
		filename := fmt.Sprintf("%s/indexer.log", Get().LogPath)
		log.Infof("Logging to %s", filename)

		log.SetFormatter(&log.JSONFormatter{})
		logger = &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		}
		log.SetOutput(io.MultiWriter(os.Stdout, logger))

		log.RegisterExitHandler(func() {
			if logger == nil {
				return
			}
			log.Info("Indexer is exiting")

			_ = logger.Close()
		})
	}
}

func Get() *Config {
	return &Config{
		Logging:            getBool("LOGGING", false),
		LogPath:            getString("LOG_PATH", "/app/logs"),
		Network:            getString("NAVCOIND_NETWORK", "mainnet"),
		Index:              getString("INDEX_NAME", "xxx"),
		SoftForkBlockCycle: getInt("SOFTFORK_BLOCKCYCLE", 20160),
		SoftForkQuorum:     getInt("SOFTFORK_QUORUM", 75),
		Debug:              getBool("DEBUG", false),
		Reindex:            getBool("REINDEX", false),
		ReindexSize:        getUint64("REINDEX_SIZE", 200),
		BulkIndexSize:      getUint64("BULK_INDEX_SIZE", 200),
		Navcoind: NavcoindConfig{
			Host:     getString("NAVCOIND_HOST", ""),
			Port:     getInt("NAVCOIND_PORT", 8332),
			User:     getString("NAVCOIND_USER", "user"),
			Password: getString("NAVCOIND_PASSWORD", "password"),
			Ssl:      getBool("NAVCOIND_SSL", false),
		},
		ElasticSearch: ElasticSearchConfig{
			Hosts:       getSlice("ELASTIC_SEARCH_HOSTS", make([]string, 0), ","),
			Sniff:       getBool("ELASTIC_SEARCH_SNIFF", true),
			HealthCheck: getBool("ELASTIC_SEARCH_HEALTH_CHECK", true),
			Debug:       getBool("ELASTIC_SEARCH_DEBUG", false),
			Username:    getString("ELASTIC_SEARCH_USERNAME", ""),
			Password:    getString("ELASTIC_SEARCH_PASSWORD", ""),
			MappingDir:  getString("ELASTIC_SEARCH_MAPPING_DIR", "/data/mappings"),
		},
		ZeroMq: ZeroMqConfig{
			Address: getString("ZEROMQ_ADDRESS", "tcp://navcoind:28332"),
		},
		RabbitMq: RabbitMqConfig{
			User:     getString("RABBITMQ_USER", "user"),
			Password: getString("RABBITMQ_PASSWORD", "user"),
			Host:     getString("RABBITMQ_HOST", "localhost"),
			Port:     getInt("RABBITMQ_PORT", 5672),
		},
		Sentry: SentryConfig{
			Active: getBool("SENTRY_ACTIVE", false),
			DSN:    getString("SENTRY_DSN", ""),
		},
	}
}

func getString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func getInt(key string, defaultValue int) int {
	valStr := getString(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}

	return defaultValue
}

func getUint(key string, defaultValue uint) uint {
	return uint(getInt(key, int(defaultValue)))
}

func getUint64(key string, defaultValue uint) uint64 {
	return uint64(getInt(key, int(defaultValue)))
}

func getBool(key string, defaultValue bool) bool {
	valStr := getString(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultValue
}

func getSlice(key string, defaultVal []string, sep string) []string {
	valStr := getString(key, "")
	if valStr == "" {
		return defaultVal
	}

	return strings.Split(valStr, sep)
}
