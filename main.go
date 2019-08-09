package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// regex function to check if given file is hidden
func isHidden(filename string) (bool, error) {
	matched, err := regexp.MatchString(`^\.`, filename)
	// if matched {
	//	fmt.Println("File", filename, "is hidden")
	// }
	return matched, err
}

func readFiles(filepath string, level int) {
	files, err := ioutil.ReadDir(filepath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {

		hidden, _ := isHidden(file.Name())
		if !hidden {
			printTokens(level+1, '-')
			fmt.Println(file.Name())
		}
		if file.IsDir() && !hidden {
			var strs []string
			strs = append(strs, filepath)
			strs = append(strs, file.Name())
			fp := strings.Join(strs, "/")
			readFiles(fp, level+1)
		}
	}
}

func printTokens(level int, token rune) {
	for i := 0; i < level; i++ {
		fmt.Printf("-")
	}
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	readFiles(root, 0)

}
