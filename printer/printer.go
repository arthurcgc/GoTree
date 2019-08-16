package printer

import (
	"fmt"
	"math"
	"os"
)

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

func PrintTokens(level int, token rune) {
	for i := 0; i < level; i++ {
		fmt.Printf("%c", token)
	}
}

func PrintFileInfo(file os.FileInfo, argMap map[string]int) {
	fmt.Printf("%s", file.Name())

	_, exists := argMap["human"]
	if !exists {
		fmt.Println(" [", file.Size(), "bytes]")
	} else {
		str, size := byteConv(int(file.Size()))
		fmt.Println(" [", size, str, "]")
	}
}
