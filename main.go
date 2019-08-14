package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"./cli"
	"./printer"
)

// regex function to check if given file is hidden
func isHidden(filename string) (bool, error) {
	matched, err := regexp.MatchString(`^\.`, filename)
	return matched, err
}

func getBuildPath(path, filename string) string {
	var strs []string
	strs = append(strs, path)
	strs = append(strs, filename)
	fp := strings.Join(strs, "/")
	return fp
}

func shouldFollowSymlink(file os.FileInfo, filepath string) (bool, string) {
	if file.Mode()&os.ModeSymlink != 0 {
		fp := getBuildPath(filepath, file.Name())
		realpath, _ := os.Readlink(fp)
		var err error
		_, err = ioutil.ReadDir(realpath)
		if err == nil {
			return true, realpath
		}
	}
	return false, ""
}

func readFiles(filepath string, level int, argMap map[string]int, doneChannel chan bool) {
	_, exists := argMap["max"]
	if exists {
		if level > argMap["max"] {
			return
		}
	}
	level++
	files, err := ioutil.ReadDir(filepath)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		hidden, _ := isHidden(file.Name())
		permission := file.Mode()
		if permission&(1<<2) == 0 {
			hidden = true
			fmt.Println("User does not have permission to read:", file.Name())
		}
		if !hidden {
			printer.PrintTokens(level, '\t')
			printer.PrintFileInfo(file, argMap)
			shouldFollow, realpath := shouldFollowSymlink(file, filepath)
			if shouldFollow {
				readFiles(realpath, level, argMap, doneChannel)
			}
		}
		if file.IsDir() && !hidden {
			fp := getBuildPath(filepath, file.Name())
			readFiles(fp, level, argMap, doneChannel)
		}
	}
	if level == 1 && doneChannel != nil {
		doneChannel <- true
	}
}

func sleeping(timeout chan bool, dur int) {
	time.Sleep(time.Second * time.Duration(dur))
	timeout <- true
}

func main() {
	args := cli.NewArgs()
	if len(args.ArgMap) == 0 {
		readFiles(args.Root, 0, nil, nil)
	} else {
		if args.HelpFlag() {
			printer.PrintHelp()
			return
		}
		if args.ExistsTime() {
			doneChannel := make(chan bool, 1)
			timeout := make(chan bool, 1)
			go readFiles(args.Root, 0, args.ArgMap, doneChannel)
			go sleeping(timeout, args.ArgMap["time"])
			select {
			case <-timeout:
				return
			case <-doneChannel:
				return
			}
		} else {
			readFiles(args.Root, 0, args.ArgMap, nil)
		}
	}
}
