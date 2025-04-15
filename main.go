package main

import (
	_ "embed"
	"fmt"

	lineProtocol "github.com/s-r-engineer/library/lineProtocol"
	libraryLogging "github.com/s-r-engineer/library/logging"
	librarySync "github.com/s-r-engineer/library/sync"
)

func main() {
	nodes := parseTheEnv()
	add, done, wait := librarySync.GetWait()
	accumulator := lineProtocol.NewAccumulator()
	for _, node := range nodes {
		add()
		go func(node string) {
			defer done()
			m, err := newMikrotik(node, &accumulator)
			if err != nil {
				panic(err)
			}
			err = m.Run()
			if err != nil {
				libraryLogging.Error(err.Error())
			}
		}(node)
	}
	wait()
	fmt.Print(string(accumulator.GetBytes()))
}
