package helper

import (
	"github.com/go-redis/redis/v8"
	"github.com/neccoys/go-driver/mysqlx"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	COPO_DB     *gorm.DB
	REDIS       *redis.Client
	RpcServices sync.Map
	err         error
)

func init() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("etc")

	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(0)
	}
}

func init() {
	COPO_DB, err = mysqlx.New(
		viper.GetString("DB_HOST"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_USERNAME"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_DATABASE"),
	).
		SetCharset("utf8").
		SetLoc("UTC").
		SetLogger(logger.Default.LogMode(logger.Info)).
		Connect(mysqlx.Pool(1, 2, 180))

	if err != nil {
		log.Panicln("COPO_DB Error:", err)
	}
}

func init() {
	REDIS = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    viper.GetString("REDIS_MASTER_NAME"),
		SentinelAddrs: strings.Split(viper.GetString("REDIS_SENTINEL_NODE"), ";"),
		DB:            viper.GetInt("REDIS_DB"),
	})

	if err != nil {
		log.Panicln("REDIS Error:", err)
	}
}
