package service_launch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/coreos/etcd/clientv3"
)

const cfgContent = `
a=b
c=1
`

func TestConfigCenter(t *testing.T) {
	var err error
	key := "test_k1"

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 10 * time.Second,
	})
	assert.Equal(t, nil, err)
	defer cli.Close()
	cc := NewConfigCenter(cli, "/test_config")

	assert.Equal(t, nil, err)
	err = cc.SetConfig(key, cfgContent)
	assert.Equal(t, nil, err)
	actual, err := cc.GetConfig(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, cfgContent, actual)
	err = cc.RemoveConfig(key)
	assert.Equal(t, nil, err)
}
