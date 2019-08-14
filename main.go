package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
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
		}
		if !hidden {
			printTokens(level, '\t')
			fmt.Printf("%s", file.Name())
			_, exists := argMap["human"]
			if !exists {
				fmt.Println(" [", file.Size(), "bytes]")
			} else {
				str, size := byteConv(int(file.Size()))
				fmt.Println(" [", size, str, "]")
			}
			if file.Mode()&os.ModeSymlink != 0 {
				var strs []string
				strs = append(strs, filepath)
				strs = append(strs, file.Name())
				fp := strings.Join(strs, "/")
				realpath, _ := os.Readlink(fp)
				var err error
				_, err = ioutil.ReadDir(realpath)
				if err == nil {
					readFiles(realpath, level, argMap, doneChannel)
				}
			}
		}
		if file.IsDir() && !hidden {
			var strs []string
			strs = append(strs, filepath)
			strs = append(strs, file.Name())
			fp := strings.Join(strs, "/")
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

func validateArgs() (map[string]int, string, error) {
	argMap := make(map[string]int)
	human := flag.Bool("human", false, "Prints size of files in human form")
	max := flag.Int("max", -1, "maximum number of levels to go trough")
	time := flag.Int("time", -1, "Set maximum elapsed time of program in seconds")
	help := flag.Bool("help", false, "show help")
	flag.Parse()
	if *human != false {
		argMap["human"] = 1
	}
	if *help != false {
		argMap["help"] = 1
	}
	if *max != -1 {
		var err error
		argMap["max"] = *max
		if err != nil {
			panic(err)
		}
	}
	if *time != -1 {
		var err error
		argMap["time"] = *time
		if err != nil {
			panic(err)
		}
	}

	args := flag.Args()
	if len(args) > 1 {
		err := errors.New("Two or more arguments passed, need one")
		panic(err)
	}

	if len(args) == 0 {
		return argMap, ".", nil
	}
	return argMap, args[0], nil
}

func sleeping(timeout chan bool, dur int) {
	time.Sleep(time.Second * time.Duration(dur))
	timeout <- true
}

func main() {
	argMap, root, err := validateArgs()
	if err != nil {
		panic(err)
	}
	if len(argMap) == 0 {
		readFiles(root, 0, nil, nil)
	} else {
		_, existsTime := argMap["time"]
		_, help := argMap["help"]
		if help {
			printHelp()
			return
		}
		if existsTime {
			doneChannel := make(chan bool, 1)
			timeout := make(chan bool, 1)
			go readFiles(root, 0, argMap, doneChannel)
			go sleeping(timeout, argMap["time"])
			select {
			case <-timeout:
				return
			case <-doneChannel:
				return
			}
		} else {
			readFiles(root, 0, argMap, nil)
		}
	}
}

func printHelp() {
	fmt.Println("-human \t display file size in human form")
	fmt.Println("-time=[int] \t set maximun time in seconds")
	fmt.Println("-max=[int] \t set maximun level of directories")
}
