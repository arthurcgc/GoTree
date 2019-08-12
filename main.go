package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
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
	_, exists := argMap["-max"]
	if exists {
		if level > argMap["-max"] {
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
		if !hidden {
			printTokens(level+1, '\t')
			fmt.Printf("%s", file.Name())
			_, exists := argMap["-conv"]
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

func getArgs(args []string) []string {
	if len(args) > 2 {
		args = args[2:]
		return args
	}

	return nil
}

func byteConv(bytes int) (string, float64) {
	check := float64(bytes) * math.Pow10(-3)
	if check < 1 {
		return "bytes", float64(bytes)
	}
	check = float64(bytes) * math.Pow10(-6)
	if check < 1 {
		return "kb", check * math.Pow10(3)
	}
	check = float64(bytes) * math.Pow10(-9)
	if check < 1 {
		return "mb", check * math.Pow10(3)
	}
	return "gb", check
}

func getRoot() string {
	var root string
	n := len(os.Args)
	if n < 2 {
		root = "."
	} else {
		root = os.Args[1]
	}

	return root
}

func validateArgs(args []string) (map[string]int, error) {
	// validArgs := []string{"-conv", "-max", "-time"}
	argMap := make(map[string]int)
	for _, arg := range args {
		matched1, _ := regexp.MatchString(`^-conv`, arg)
		if matched1 {
			argMap["-conv"] = 1
		}
		matched2, _ := regexp.MatchString(`-max=[0-9]+`, arg)
		if matched2 {
			re := regexp.MustCompile(`[0-9]+`)
			val := re.FindString(arg)
			// fmt.Println("passed max value is: ", val)
			argMap["-max"], _ = strconv.Atoi(val)
		}
		matched3, _ := regexp.MatchString(`-time=[0-9]+`, arg)
		if matched3 {
			re := regexp.MustCompile(`[0-9]+`)
			val := re.FindString(arg)
			// fmt.Println("passed time limit is: ", val)
			argMap["-time"], _ = strconv.Atoi(val)
		}

		if !matched1 && !matched2 && !matched3 {
			return nil, errors.New("Invalid Argument(s) passed")
		}
	}

	return argMap, nil
}

func sleeping(timeout chan bool, dur int) {
	time.Sleep(time.Second * time.Duration(dur))
	timeout <- true
}

func main() {
	root := getRoot()
	args := getArgs(os.Args)
	argMap, err := validateArgs(args)
	if err != nil {
		panic(err)
	}
	if args == nil {
		readFiles(root, 0, nil, nil)
	} else {
		_, existTime := argMap["-time"]
		if existTime {
			doneChannel := make(chan bool, 1)
			timeout := make(chan bool, 1)
			go readFiles(root, 0, argMap, doneChannel)
			go sleeping(timeout, argMap["-time"])
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
