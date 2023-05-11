package etcd

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/client/v3"
)

var (
	Client *clientv3.Client
)

const (
	ActionPUT    = "PUT"
	ActionDELETE = "DELETE"
)

func Init(address []string) {
	var err error
	Client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to connect to etcd: %s", err.Error()))
	}
}

func FakePut(ctx context.Context, key, val string) (interface{}, error) {
	return nil, nil
}
