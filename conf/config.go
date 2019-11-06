package conf

import (
	"gopkg.in/ini.v1"
	"strings"
	"sync"
)

type ConfigFile string

var (
	Configuration *Config
	configLock    sync.Mutex
)

func NewConfig(configFile ConfigFile) (*Config, error) {

	if Configuration != nil {
		return Configuration, nil
	}

	configLock.Lock()
	defer configLock.Unlock()

	//初始化配置相关的内容
	localConfig := &Config{}
	cfg, err := ini.Load(string(configFile))

	if err != nil {
		return nil, err
	}

	localConfig.cfg = cfg
	Configuration = localConfig

	return Configuration, nil
}

//配置
type Config struct {
	cfg *ini.File
}

//解析获取配置的key   key1:key2:key3
func parseKey(key string) (section string, option string) {
	des := strings.Split(key, ":")
	if len(des) < 2 {
		return "", ""
	}
	section = des[0]
	option = strings.Join(des[1:], "")
	return
}

//获取字符串
func (s *Config) GetString(key string) string {
	section, option := parseKey(key)
	return s.cfg.Section(section).Key(option).String()
}

//获取整形
func (s *Config) GetInt(key string) int {
	section, option := parseKey(key)
	v, err := s.cfg.Section(section).Key(option).Int()
	if err != nil {
		return 0
	}
	return v
}

//获取布尔值
func (s *Config) GetBool(key string) bool {
	section, option := parseKey(key)
	v, err := s.cfg.Section(section).Key(option).Bool()
	if err != nil {
		return false
	}
	return v
}
