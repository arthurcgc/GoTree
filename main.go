package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
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

func readFiles(filepath string, level int, argMap map[string]bool) {
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
		}
		if file.IsDir() && !hidden {
			var strs []string
			strs = append(strs, filepath)
			strs = append(strs, file.Name())
			fp := strings.Join(strs, "/")
			readFiles(fp, level+1, argMap)
		}
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

func validateArgs(args []string) (map[string]bool, error) {
	validArgs := []string{"-conv", "-max", "-time"}
	argMap := make(map[string]bool)
	for _, arg := range args {
		validFlag := false
		for _, flag := range validArgs {
			if arg == flag {
				validFlag = true
				argMap[arg] = true
			}
		}
		if !validFlag {
			return nil, errors.New("Invalid Argument(s) passed")
		}
	}

	return argMap, nil
}

func main() {
	// start := time.Now()
	root := getRoot()
	args := getArgs(os.Args)
	argMap, err := validateArgs(args)
	if err != nil {
		panic(err)
	}

	if args == nil {
		readFiles(root, 0, nil)
	} else {
		readFiles(root, 0, argMap)
	}
}
