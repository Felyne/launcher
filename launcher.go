package launcher

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Felyne/configcenter"
	"github.com/micro/go-micro"

	"github.com/coreos/etcd/clientv3"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/etcdv3"
)

type SetupFunc func(service micro.Service, envName, cfgContent string) error

// example: ./server dev 0 localhost:2379
func Run(serviceName, version, buildTime string, setup SetupFunc) {
	if len(os.Args) < 4 {
		if len(os.Args) == 2 && os.Args[1] == "-v" {
			fmt.Printf("version: %s\nbuildTime: %s\nserviceName: %s\n",
				version, buildTime, serviceName)
		} else {
			help()
		}
		os.Exit(1)
	}
	envName := os.Args[1]    //环境名
	portStr := os.Args[2]    //监听端口，0表示自动分配
	etcdAddrs := os.Args[3:] //etcd地址

	err := run(serviceName, version, envName, portStr, etcdAddrs, setup)
	if err != nil {
		log.Fatal(err)
	}
}

//根据env和serviceName从etcd获取配置，服务启动后服务信息注册到etcd
func run(serviceName, version, envName, portStr string,
	etcdAddrs []string, setup SetupFunc) error {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		return err
	}

	cc := configcenter.New(cli, envName)
	cfgContent, err := cc.GetConfig(serviceName)
	if err != nil {
		return err
	}
	cli.Close()

	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = etcdAddrs
	})

	options := []micro.Option{
		micro.Name(GenServiceRegName(envName, serviceName)),
		micro.Registry(reg),
		micro.RegisterTTL(10*time.Second),
		micro.RegisterInterval(5*time.Second),
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
	if err := setup(service, envName, cfgContent); err != nil {
		return err
	}

	return service.Run()
}

func help() {
	info := `
Usage:%s [envName] [port] [etcdAddr...]
  envName  env namespace
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

//生成服务的注册名
func GenServiceRegName(envName, serviceName string) string {
	return envName + ".service." + serviceName
}
