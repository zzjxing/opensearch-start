package config

import (
	"fmt"
	"os"
	"sync"
)

var (
	OpenSearchADDR     string
	OpenSearchUser     string
	OpenSearchPassword string
)
var once sync.Once

func Init() {
	var checkEnvVar = func(name string) string {
		value := os.Getenv(name)
		if value == "" {
			fmt.Printf("%s 初始化失败，请设置环境变量: %s\n", name, name)
			os.Exit(1)
		}
		return value
	}
	once.Do(func() {
		OpenSearchADDR = checkEnvVar("OPENSEARCH_HOSTS")
		OpenSearchUser = checkEnvVar("OPENSEARCH_USERNAME")
		OpenSearchPassword = checkEnvVar("OPENSEARCH_PASSWORD")
	})

}
