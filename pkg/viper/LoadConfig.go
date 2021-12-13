package viper

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DB       string `mapstructure:"DB_DB"`
}

// 讀專案中的config檔
func LoadConfig(mypath string) (config Config) {
	// 若有同名環境變量則使用環境變量
	viper.AutomaticEnv()
	viper.AddConfigPath(mypath)
	// 為了讓執行test也能讀到config添加索引路徑
	wd, err := os.Getwd()
	parent := filepath.Dir(wd)
	viper.AddConfigPath(path.Join(parent, mypath))
	viper.SetConfigName("db")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("can not load config: " + err.Error())
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("can not load config: " + err.Error())
	}
	return
}
