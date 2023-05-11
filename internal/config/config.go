package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type c struct {
	Port             int
	EtcdHosts        []string
	EtcdWorkerPrefix string
	Database         struct {
		MySQLHost     string
		MySQLPort     int
		MySQLUser     string
		MySQLPassword string
		MySQLDB       string
	}
	// 控制不同优先级的发送密度，如果队列打满
	// PriorityFactor = 2 , high : mid : low = 4 : 2 : 1
	// PriorityFactor = 10 , high : mid : low = 100 : 10 : 1
	// 如果队列不满，则继续优先高优先级
	PriorityFactor int
	Kinds          []string
	Priorities     []string // []string{"high", "mid", "low"}
}

func HasKind(x string) bool {
	return validKind[x]
}
func HasPriority(x string) bool {
	return validPriority[x]
}

var (
	Config        c
	doOnce        sync.Once
	validKind     = make(map[string]bool)
	validPriority = make(map[string]bool)
)

func Init(filePath string) {
	doOnce.Do(func() {
		content, err := os.ReadFile(filePath)
		if err != nil {
			panic(fmt.Sprintf("failed to load config file: %s", err.Error()))
		}

		if err = yaml.Unmarshal(content, &Config); err != nil {
			panic(fmt.Sprintf("failed to generate config: %s", err.Error()))
		}

		Config.Priorities = []string{"high", "mid", "low"}
		for _, x := range Config.Priorities {
			validPriority[x] = true
		}
		for _, x := range Config.Kinds {
			validKind[x] = true
		}
		runningConfig, _ := yaml.Marshal(&Config)
		fmt.Printf("--- configuration ---\n\n%s\n\n--- configuration ---", string(runningConfig))
		time.Sleep(time.Second)
	})
}
