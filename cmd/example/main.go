package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"goudeaux.com/img/cache"
)

func main() {
	var wg sync.WaitGroup
	c := cache.New(25)

	wg.Add(2)
	go func() {
		for i := 0; i < 1000; i++ {
			c.Set(strconv.Itoa(i%25), strconv.Itoa(i%25+1), 20*time.Millisecond)
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}()
	go func() {
		time.Sleep(25 * 5 * time.Millisecond)
		for i := 0; i < 1000; i++ {
			c.Get(strconv.Itoa(i % 25))
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}()
	wg.Wait()
	b, _ := json.MarshalIndent(c.Stats(), "", "  ")
	fmt.Println(string(b))
}
