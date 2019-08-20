package directory

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"../timedEvent"
	"github.com/arthurcgc/GoTree/printer"
)

type dir struct {
	filepath string
	level    int
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

func getBuildPath(path, filename string) string {
	var strs []string
	strs = append(strs, path)
	strs = append(strs, filename)
	fp := strings.Join(strs, "/")
	return fp
}

func isHidden(filename string) (bool, error) {
	matched, err := regexp.MatchString(`^\.`, filename)
	return matched, err
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

func (d *dir) getFileList(tEvent *timedEvent.TimedEvent) ([]os.FileInfo, error) {
	d.level++
	files, err := ioutil.ReadDir(d.filepath)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (d *dir) ReadFiles(argMap map[string]int, tEvent *timedEvent.TimedEvent) (bool, error) {
	if _, exists := argMap["max"]; exists && d.level >= argMap["max"] {
		return true, nil
	}
	files, err := d.getFileList(tEvent)
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
		printer.PrintTokens(d.level, '\t')
		printer.PrintFileInfo(file, argMap)
		shouldFollow, realpath, err := shouldFollowSymlink(file, d.filepath)
		if checkError(tEvent, err) != nil {
			return false, err
		}
		if shouldFollow {
			newroot := NewDirectory(realpath, d.level)
			var breakRecursion bool
			breakRecursion, err = newroot.ReadFiles(argMap, tEvent)
			if checkError(tEvent, err) != nil {
				return false, err
			}
			if breakRecursion {
				return true, nil
			}
			continue
		}
		if file.IsDir() {
			fp := getBuildPath(d.filepath, file.Name())
			newroot := NewDirectory(fp, d.level)
			var breakRecursion bool
			breakRecursion, err = newroot.ReadFiles(argMap, tEvent)
			if checkError(tEvent, err) != nil {
				return false, err
			}
			if breakRecursion {
				return true, nil
			}
		}
	}

	if d.level == 1 && tEvent != nil {
		tEvent.ChanSwitch()
		return true, nil
	}
	return false, nil
}

func NewDirectory(path string, level int) dir {
	newdir := dir{filepath: path, level: level}
	return newdir
}
