package main

import (
	"github.com/mikeqian/go-zookeeper/zk"
	"time"
)

func main() {
	c1 := register("http://www.baidu.com/v1")
	defer c1.Close()

	c2 := register("http://www.baidu.com/v2")
	defer c2.Close()

	for {
		time.Sleep(time.Second * 30)
	}
}

func register(url string) *zk.Conn {
	c, _, err := zk.Connect([]string{"10.0.2.214:2181", "10.0.2.215:2181", "10.0.2.216:2181"}, time.Second*10)
	if err != nil {
		panic(err)
	}
	acls := zk.WorldACL(zk.PermAll)
	s := zk.NewService(c, "/services/log", acls)
	s.Register(url)

	return c
}
