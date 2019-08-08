package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

// regex function to check if given file is hidden
func isHidden(filename string) (bool, error) {
	matched, err := regexp.MatchString(`^\..*`, filename)
	// if matched {
	//	fmt.Println("File", filename, "is hidden")
	// }
	return matched, err
}

func printTokens(total int, token rune, name string) {
	for i := total; i < total; i++ {
		fmt.Printf("%c", token)
	}
	fmt.Printf("%s\n", name)
}

func printTree(root string, files []os.FileInfo, nTokens int) {
	token := '-'
	var dirs []os.FileInfo
	for i := 0; i < nTokens; i++ {
		fmt.Printf("%c", token)
	}
	fmt.Printf("%s\n", root)

	for i := 0; i < len(files); i++ {
		file := files[i]
		if file.IsDir() {
			dirs = append(dirs, file)
		}
		printTokens(nTokens+1, token, file.Name())
	}
	nTokens++

	for _, dir := range dirs {
		normalFiles, _ := getFiles(dir.Name())
		printTree(dir.Name(), normalFiles, nTokens)
	}
}

func getFiles(root string) ([]os.FileInfo, []os.FileInfo) {
	var files []os.FileInfo
	var err error
	files, err = ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}

	var normalFiles []os.FileInfo
	var hiddenFiles []os.FileInfo
	for _, file := range files {
		name := file.Name()
		// fmt.Println("hiddenVsNormal: file name = ", name)
		hidden, err := isHidden(name)
		if err != nil {
			panic(err)
		}
		if !hidden && name != root {
			normalFiles = append(normalFiles, file)
		} else {
			hiddenFiles = append(hiddenFiles, file)
		}
	}

	return normalFiles, hiddenFiles
}

func printFilesNames(files []os.FileInfo) {
	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func fillGraph(root string) Graph {
	var graph Graph
	var dirCount int

	files, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}

	fmt.Println(root)
	for _, file := range files {
		hidden, _ := isHidden(file.Name())
		if file.IsDir() {
			dirCount++
			node := CreateNode(file, 1, hidden)
			graph.PushBack(node)
		}
		if !hidden {
			token := '-'
			fmt.Printf("%c%s\n", token, file.Name())
		}
	}

	return graph
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	graph := fillGraph(root)

	fmt.Println(graph)

}
