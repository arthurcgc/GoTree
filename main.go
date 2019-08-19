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
}

func readFiles(root dir, argMap map[string]int, retCond *bool, tEvent *timedEvent.TimedEvent) error {
	filepath := root.filepath
	level := root.level
	if *retCond {
		return nil
	}
	if _, exists := argMap["max"]; exists && level >= argMap["max"] {
		*retCond = true
		return nil
	}
	level++
	files, err := ioutil.ReadDir(filepath)
	if checkError(tEvent, err) != nil {
		return err
	}
	for _, file := range files {
		ignore, err := isHidden(file.Name())
		if tEvent != nil && tEvent.CheckReceiveSignalNoHang() && !*retCond {
			*retCond = true
			return nil
		}
		if checkError(tEvent, err) != nil {
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
		if checkError(tEvent, err) != nil {
			return err
		}
		if shouldFollow {
			newroot := dir{filepath: realpath, level: level}
			err = readFiles(newroot, argMap, retCond, tEvent)
			if *retCond {
				return nil
			}
			if checkError(tEvent, err) != nil {
				return err
			}
			continue
		}
		if file.IsDir() {
			fp := getBuildPath(filepath, file.Name())
			newroot := dir{filepath: fp, level: level}
			err = readFiles(newroot, argMap, retCond, tEvent)
			if *retCond {
				return nil
			}
			if checkError(tEvent, err) != nil {
				return err
			}
		}
	}

	if level == 1 && tEvent != nil {
		tEvent.TurnSwitch(tEvent.Finished, tEvent.Light)
		*retCond = true
	}
	return nil
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
			tEvent.TurnSwitch(tEvent.Finished, tEvent.Light)
			tEvent.Wg.Done()
			return
		}
	}
}

func main() {
	args := cli.NewArgs()
	root := dir{filepath: args.Root, level: 0}
	var retCond bool
	if args.ExistsTime() {
		tEvent := timedEvent.NewTimedEvent()
		tEvent.Wg.Add(1)
		go sleeping(args.ArgMap["time"], tEvent)
		readFiles(root, args.ArgMap, &retCond, tEvent)
		tEvent.Wg.Wait()
	} else {
		readFiles(root, args.ArgMap, &retCond, nil)
	}
}
