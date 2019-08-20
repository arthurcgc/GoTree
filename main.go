package main

import (
	"github.com/arthurcgc/GoTree/cli"
	"github.com/arthurcgc/GoTree/directory"
	"github.com/arthurcgc/GoTree/timedEvent"
)

func main() {
	args := cli.NewArgs()
	root := directory.NewDirectory(args.Root, 0)
	if args.ExistsTime() {
		tEvent := timedEvent.NewTimedEvent()
		tEvent.Wg.Add(1)
		go tEvent.Sleeping(args.ArgMap["time"])
		root.ReadFiles(args.ArgMap, tEvent)
		tEvent.Wg.Wait()
	} else {
		root.ReadFiles(args.ArgMap, nil)
	}
}
