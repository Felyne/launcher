package service_launch

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var ErrNoConfig = errors.New("no config")

const (
	PathSeparator         = "/"
	DefaultConfigBasePath = "/config_center"
	contextTimeout        = 15 * time.Second
)

type ConfigCenter struct {
	etcdClient     *clientv3.Client
	configBasePath string
}

func NewConfigCenter(client *clientv3.Client, configBasePath string) *ConfigCenter {
	if configBasePath == "" {
		configBasePath = DefaultConfigBasePath
	}
	return &ConfigCenter{
		etcdClient:     client,
		configBasePath: configBasePath,
	}
}

func (cc *ConfigCenter) GetConfig(cfgName string) (string, error) {
	cfgPath := cc.genPath(cfgName)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		contextTimeout)
	resp, err := cc.etcdClient.Get(ctx, cfgPath)
	cancel()
	if nil != err {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", ErrNoConfig
	}
	return string(resp.Kvs[0].Value), err
}

func (cc *ConfigCenter) SetConfig(cfgName string, content string) error {
	cfgPath := cc.genPath(cfgName)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		contextTimeout)
	_, err := cc.etcdClient.Put(ctx, cfgPath, content)
	cancel()
	return err
}

func (cc *ConfigCenter) RemoveConfig(cfgName string) error {
	cfgPath := cc.genPath(cfgName)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		contextTimeout)
	_, err := cc.etcdClient.Delete(ctx, cfgPath)
	cancel()
	return err
}

func (cc *ConfigCenter) genPath(cfgName string) string {
	return cc.configBasePath + PathSeparator + cfgName
}
