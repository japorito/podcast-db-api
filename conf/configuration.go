package conf

import (
	"fmt"
	"github.com/fsnotify/fsnotify" // TODO: Get rid of this if not needed for OnConfigChange
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var Machine DefangedViper

var configurations map[string]*viper.Viper = make(map[string]*viper.Viper)
var configurationLock sync.RWMutex

type DefangedViper interface {
	AllKeys() []string
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	IsSet(key string) bool
}

func GetConfiguration(filename string) (DefangedViper, error) {
	configurationLock.RLock()
	conf, found := configurations[filename]
	configurationLock.RUnlock()

	if !found {
		conf, err := loadConfiguration(filename)
		if err != nil {
			return nil, err
		}

		configurationLock.Lock()
		configurations[filename] = conf
		configurationLock.Unlock()
	}

	return conf, nil
}

func SetPrimaryConfiguration(path string) (DefangedViper, error) {
	machineConf := viper.New()

	// treat .conf files like java .properties
	if filepath.Ext(path) == ".conf" {
		machineConf.SetConfigType("properties")
	}

	machineConf.SetConfigFile(path)

	if err := machineConf.ReadInConfig(); err != nil {
		return nil, err
	}

	Machine = machineConf

	configurationLock.Lock()
	configurations["machine"] = machineConf
	configurationLock.Unlock()

	return Machine, nil
}

func loadConfiguration(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)

	confPath := Machine.GetString("configuration.path")
	for _, directory := range strings.Split(confPath, ",") {
		directory = strings.TrimSpace(directory)

		if directory != "" {
			v.AddConfigPath(directory)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if !Machine.GetBool("production") {
		v.WatchConfig()

		// TODO: Can I get rid of this? I get a segfault on conf change without it.
		v.OnConfigChange(func(in fsnotify.Event) {
			fmt.Println("Config changed!")
		})
	}

	return v, nil
}
