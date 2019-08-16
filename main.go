package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/arthurcgc/GoTree/cli"
	"github.com/arthurcgc/GoTree/printer"
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

func shouldFollowSymlink(file os.FileInfo, filepath string) (bool, string, error) {
	if file.Mode()&os.ModeSymlink == 0 {
		return false, "", nil
	}
	fp := getBuildPath(filepath, file.Name())
	realpath, err := os.Readlink(fp)
	if err != nil {
		return false, "", err
	}
	_, err = ioutil.ReadDir(realpath)
	if err != nil {
		return false, realpath, nil
	}
	return true, realpath, nil
}

func hasPermission(mode uint32) bool {
	others := uint32(1 << 2)
	group := uint32(1 << 5)
	user := uint32(1 << 8)

	if (mode&others) != 0 || (mode&group) != 0 || (mode&user) != 0 {
		return true
	}
	return false
}

var timeout bool

func readFiles(filepath string, level int, argMap map[string]int, light *sync.Mutex) error {
	if timeout {
		return nil
	}
	if _, exists := argMap["max"]; exists && level >= argMap["max"] {
		return nil
	}
	level++
	files, err := ioutil.ReadDir(filepath)
	if err != nil {
		return err
	}
	for _, file := range files {
		ignore, err := isHidden(file.Name())
		if timeout {
			return nil
		}
		if err != nil {
			return err
		}
		if !hasPermission(uint32(file.Mode())) {
			ignore = true
			fmt.Println("User does not have permission to read:", file.Name())
		}
		if ignore {
			continue
		}
		printer.PrintTokens(level, '\t')
		printer.PrintFileInfo(file, argMap)
		shouldFollow, realpath, err := shouldFollowSymlink(file, filepath)
		if err != nil {
			return err
		}
		if shouldFollow {
			readFiles(realpath, level, argMap, light)
		}
		if file.IsDir() && !shouldFollow {
			fp := getBuildPath(filepath, file.Name())
			readFiles(fp, level, argMap, light)
		}
	}

	if level == 1 && light != nil {
		light.Lock()
		defer func() {
			light.Unlock()
		}()
		timeout = true
		return nil
	}
	return nil
}

func sleeping(dur int, light *sync.Mutex) {
	seconds := time.Duration(dur) * time.Second
	expire := time.Now().Add(seconds)
	for {
		if timeout {
			fmt.Printf("Read Files took: %v seconds to parse files", expire.Sub(time.Now()))
			return
		}
		if time.Now().After(expire) {
			light.Lock()
			defer func() {
				light.Unlock()
				fmt.Printf("Program timed out!\n")
			}()
			timeout = true
			return
		}
	}
}

func main() {
	args := cli.NewArgs()
	if args.ExistsTime() {
		var light = sync.Mutex{}
		go sleeping(args.ArgMap["time"], &light)
		readFiles(args.Root, 0, args.ArgMap, &light)
	} else {
		readFiles(args.Root, 0, args.ArgMap, nil)
	}

}
