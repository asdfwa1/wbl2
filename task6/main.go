package main

import (
	"fmt"
)

func main() {
	var s = []string{"1", "2", "3"}
	modifySlice(s)
	fmt.Println(s)
}

func modifySlice(i []string) {
	i[0] = "3"
	i = append(i, "4")
	i[1] = "5"
	i = append(i, "6")
}

// программа выведет [3 2 3], так как. В функцию modifySlice передается копия среза указывающая на массив, и все изменения, которые произойдут до переаллокации памяти будут видны после выхода финкции.
