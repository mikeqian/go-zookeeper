package main

import (
	"github.com/mikeqian/go-zookeeper/zk"
	"log"
	"time"
)

func main() {
	c, _, err := zk.Connect([]string{"10.0.2.214:2181", "10.0.2.215:2181", "10.0.2.216:2181"}, time.Second*10)
	if err != nil {
		panic(err)
	}

	acls := zk.WorldACL(zk.PermAll)

	l := zk.NewLock(c, "/test", acls)
	if err := l.Lock(); err != nil {
		log.Fatal(err)
	}
	if err := l.Unlock(); err != nil {
		log.Fatal(err)
	}

	val := make(chan int, 3)

	if err := l.Lock(); err != nil {
		log.Fatal(err)
	}

	l2 := zk.NewLock(c, "/test", acls)
	go func() {
		if err := l2.Lock(); err != nil {
			log.Fatal(err)
		}
		val <- 2
		if err := l2.Unlock(); err != nil {
			log.Fatal(err)
		}
		val <- 3
	}()
	time.Sleep(time.Millisecond * 100)

	val <- 1
	if err := l.Unlock(); err != nil {
		log.Fatal(err)
	}
	if x := <-val; x != 1 {
		log.Fatal("Expected 1 instead of %d", x)
	}
	if x := <-val; x != 2 {
		log.Fatal("Expected 2 instead of %d", x)
	}
	if x := <-val; x != 3 {
		log.Fatal("Expected 3 instead of %d", x)
	}
}
