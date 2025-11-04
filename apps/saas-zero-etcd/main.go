package main

import (
	"context"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// main 连接到etcd并打印所有键值对
// 该函数会连接到本地的etcd服务（localhost:2379），获取所有键值对并打印出来
// 如果连接或查询过程中出现错误，程序会终止并输出错误信息
func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"}, //改成对应注册中心地址
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// 获取所有键值对
	resp, err := cli.Get(context.Background(), "", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Number of keys: %d\n", resp.Count)

	// 打印所有键值对
	for _, kv := range resp.Kvs {
		fmt.Printf("Key: %s, Value: %s\n", string(kv.Key), string(kv.Value))
	}
}
