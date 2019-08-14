package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arthurcgc/GoTree/cli"
)

// regex function to check if given file is hidden
func isHidden(filename string) (bool, error) {
	matched, err := regexp.MatchString(`^\.`, filename)
	return matched, err
}

func printTokens(level int, token rune) {
	for i := 0; i < level; i++ {
		fmt.Printf("%c", token)
	}
}

func printFileInfo(file os.FileInfo, argMap map[string]int) {
	fmt.Printf("%s", file.Name())
	_, exists := argMap["human"]
	if !exists {
		fmt.Println(" [", file.Size(), "bytes]")
	} else {
		str, size := byteConv(int(file.Size()))
		fmt.Println(" [", size, str, "]")
	}
}

func buildPath(path, filename string) string {
	var strs []string
	strs = append(strs, path)
	strs = append(strs, filename)
	fp := strings.Join(strs, "/")
	return fp
}

func shouldFollowSymlink(file os.FileInfo, filepath string) (bool, string) {
	if file.Mode()&os.ModeSymlink != 0 {
		fp := buildPath(filepath, file.Name())
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
			printTokens(level, '\t')
			printFileInfo(file, argMap)
			shouldFollow, realpath := shouldFollowSymlink(file, filepath)
			if shouldFollow {
				readFiles(realpath, level, argMap, doneChannel)
			}
		}
		if file.IsDir() && !hidden {
			fp := buildPath(filepath, file.Name())
			readFiles(fp, level, argMap, doneChannel)
		}
	}
	if level == 1 && doneChannel != nil {
		doneChannel <- true
	}
}

func byteConv(bytes int) (string, float64) {
	check := float64(bytes) * math.Pow10(-3)
	if check < 1 {
		return "BYTES", float64(bytes)
	}
	check = float64(bytes) * math.Pow10(-6)
	if check < 1 {
		return "KiB", check * math.Pow10(3)
	}
	check = float64(bytes) * math.Pow10(-9)
	if check < 1 {
		return "MiB", check * math.Pow10(3)
	}
	return "GiB", check
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
		_, existsTime := args.ArgMap["time"]
		_, help := args.ArgMap["help"]
		if help {
			printHelp()
			return
		}
		if existsTime {
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

func printHelp() {
	fmt.Println("-human \t display file size in human form")
	fmt.Println("-time=[int] \t set maximun time in seconds")
	fmt.Println("-max=[int] \t set maximun level of directories")
}
