package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arthurcgc/GoTree/cli"
	"github.com/arthurcgc/GoTree/printer"
	"github.com/arthurcgc/GoTree/timedEvent"
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

func checkError(tEvent *timedEvent.TimedEvent, err error) error {
	if err != nil {
		if tEvent != nil {
			tEvent.Finished <- true
		}
		return err
	}
	return nil
}

type dir struct {
	filepath string
	level    int
	visited  bool
}

func getFileList(root *dir, tEvent *timedEvent.TimedEvent) ([]os.FileInfo, error) {
	root.level++
	files, err := ioutil.ReadDir(root.filepath)
	if checkError(tEvent, err) != nil {
		return nil, err
	}
	root.visited = true

	return files, nil
}

func readFiles(root dir, argMap map[string]int, tEvent *timedEvent.TimedEvent) (bool, error) {
	if _, exists := argMap["max"]; exists && root.level >= argMap["max"] {
		return true, nil
	}
	files, err := getFileList(&root, tEvent)
	if err != nil {
		return true, err
	}
	for _, file := range files {
		ignore, err := isHidden(file.Name())
		if tEvent != nil && tEvent.CheckReceiveSignalNoHang() {
			return true, nil
		}
		if checkError(tEvent, err) != nil {
			return false, err
		}
		if !hasPermission(uint32(file.Mode())) {
			ignore = true
			fmt.Println("User does not have permission to read:", file.Name())
		}
		if ignore {
			continue
		}
		printer.PrintTokens(root.level, '\t')
		printer.PrintFileInfo(file, argMap)
		shouldFollow, realpath, err := shouldFollowSymlink(file, root.filepath)
		if checkError(tEvent, err) != nil {
			return false, err
		}
		if shouldFollow {
			newroot := dir{filepath: realpath, level: root.level, visited: false}
			var breakLoop bool
			breakLoop, err = readFiles(newroot, argMap, tEvent)
			if checkError(tEvent, err) != nil {
				return false, err
			}
			if breakLoop {
				return true, nil
			}
			continue
		}
		if file.IsDir() {
			fp := getBuildPath(root.filepath, file.Name())
			newroot := dir{filepath: fp, level: root.level, visited: false}
			var breakLoop bool
			breakLoop, err = readFiles(newroot, argMap, tEvent)
			if checkError(tEvent, err) != nil {
				return false, err
			}
			if breakLoop {
				return true, nil
			}
		}
	}

	if root.level == 1 && tEvent != nil {
		tEvent.ChanSwitch(tEvent.Finished, tEvent.Light)
		return true, nil
	}
	return false, nil
}

func sleeping(dur int, tEvent *timedEvent.TimedEvent) {
	seconds := time.Duration(dur) * time.Second
	select {
	case <-tEvent.Finished:
		{
			tEvent.Wg.Done()
			return
		}
	case <-time.After(seconds):
		{
			fmt.Printf("\nProgram timed out!\n")
			tEvent.ChanSwitch(tEvent.Finished, tEvent.Light)
			tEvent.Wg.Done()
			return
		}
	}
}

func main() {
	args := cli.NewArgs()
	root := dir{filepath: args.Root, level: 0, visited: false}
	if args.ExistsTime() {
		tEvent := timedEvent.NewTimedEvent()
		tEvent.Wg.Add(1)
		go sleeping(args.ArgMap["time"], tEvent)
		readFiles(root, args.ArgMap, tEvent)
		tEvent.Wg.Wait()
	} else {
		readFiles(root, args.ArgMap, nil)
	}
}
