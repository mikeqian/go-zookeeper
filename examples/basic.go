package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func main() {
	c, _, err := zk.Connect([]string{"10.0.2.214:2181", "10.0.2.215:2181", "10.0.2.216:2181"}, time.Second*10)
	if err != nil {
		panic(err)
	}

	defer c.Close()
	acls := zk.WorldACL(zk.PermAll)

	l := zk.NewLock(c, "/test", acls)
	if err := l.Lock(); err != nil {
		fmt.Println(err)
	}

	val := make(chan int, 3)
	l2 := zk.NewLock(c, "/test", acls)
	go func() {
		if err := l2.Lock(); err != nil {
			fmt.Println(err)
		}
		val <- 2
		if err := l2.Unlock(); err != nil {
			fmt.Println(err)
		}
		val <- 3
	}()
	time.Sleep(time.Millisecond * 100)
	val <- 1
	if err := l.Unlock(); err != nil {
		fmt.Println(err)
	}
	if x := <-val; x != 1 {
		fmt.Printf("Expected 1 instead of %d\n", x)
	}
	if x := <-val; x != 2 {
		fmt.Printf("Expected 2 instead of %d\n", x)
	}
	if x := <-val; x != 3 {
		fmt.Printf("Expected 3 instead of %d\n", x)
	}
}
