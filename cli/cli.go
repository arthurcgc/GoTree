package cli

import (
	"errors"
	"flag"
)

type Args struct {
	ArgMap map[string]int
	Root   string
}

func NewArgs() Args {
	args := Args{}
	args.ArgMap = make(map[string]int)
	human := flag.Bool("human", false, "Prints size of files in human form")
	max := flag.Int("max", -1, "maximum number of levels to go trough")
	time := flag.Int("time", -1, "Set maximum elapsed time of program in seconds")
	help := flag.Bool("help", false, "show help")
	flag.Parse()
	if *human != false {
		args.ArgMap["human"] = 1
	}
	if *help != false {
		args.ArgMap["help"] = 1
	}
	if *max != -1 {
		var err error
		args.ArgMap["max"] = *max
		if err != nil {
			panic(err)
		}
	}
	if *time != -1 {
		var err error
		args.ArgMap["time"] = *time
		if err != nil {
			panic(err)
		}
	}

	listArgs := flag.Args()
	if len(listArgs) > 1 {
		err := errors.New("Two or more arguments passed, need one")
		panic(err)
	}
	if len(listArgs) == 0 {
		args.Root = "."
		return args
	}
	args.Root = listArgs[0]
	return args
}
