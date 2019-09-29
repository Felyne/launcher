package service_launch

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/micro/go-micro"

	"github.com/coreos/etcd/clientv3"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"
	"github.com/micro/go-plugins/registry/etcdv3"
)

//每个服务端都要提供
type SetupFunc func(s server.Server, cfgContent string) error

// example: ./server 0 localhost:2379
func Start(serviceName, version, buildTime string, setup SetupFunc) {
	if len(os.Args) < 3 {
		if len(os.Args) == 2 && os.Args[1] == "-v" {
			fmt.Printf("version: %s\nbuildTime: %s\nserviceName: %s\n",
				version, buildTime, serviceName)
		} else {
			help()
		}
		os.Exit(1)
	}
	portStr := os.Args[1] //监听端口，0表示自动分配
	etcdAddrs := os.Args[2:]

	err := run(serviceName, version, portStr, etcdAddrs, setup)
	if err != nil {
		log.Fatal(err)
	}
}

func run(serviceName, version, portStr string, etcdAddrs []string, setup SetupFunc) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		return err
	}

	cc := NewConfigCenter(cli, "")
	cfgContent, err := cc.GetConfig(serviceName)
	if err != nil {
		return err
	}
	cli.Close()

	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = etcdAddrs
	})

	options := []micro.Option{
		micro.Name(serviceName),
		micro.Registry(reg),
	}
	listenAddr := getAddr(portStr)
	if listenAddr != "" {
		options = append(options, micro.Address(listenAddr))
	}
	if version != "" {
		options = append(options, micro.Version(version))
	}

	service := micro.NewService(options...)
	service.Init()
	err = setup(service.Server(), cfgContent)
	if err != nil {
		return err
	}

	return service.Run()
}

func help() {
	info := `
Usage:%s [port] [etcdAddr...]
  port     port for listen.if value is 0,listen on a random port
`
	fmt.Printf(info, os.Args[0])
}

func getAddr(port string) (addr string) {
	if port == "0" || port == ":0" {
		return
	}
	if false == strings.HasPrefix(port, ":") {
		addr = ":" + port
	}
	return
}
