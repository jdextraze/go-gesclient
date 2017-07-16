package main

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type parent struct {
	abstractMethod func()
}

func (p *parent) test() {
	p.abstractMethod()
}

type Children struct {
	*parent
}

func (c *Children) method() {
	fmt.Println("Test")
}

func test(p *parent) {
	p.test()
}

func main() {
	c := &Children{}
	c.parent = &parent{c.method}
	c.test()

	test(c.parent)

	err := tasks.NewStarted(func() (interface{}, error) {
		fmt.Println("1")
		<-time.After(30 * time.Second)
		fmt.Println("2")
		result := true
		return &result, nil
	}).ContinueWith(func(t2 *tasks.Task) error {
		var result bool
		fmt.Println("3")
		err := t2.Result(&result)
		fmt.Println(result, err)
		return nil
	}).Wait()
	fmt.Println(err)
}
