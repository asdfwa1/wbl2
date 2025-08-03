package main

import "fmt"

func test() (x int) {
	defer func() {
		x++
	}()
	x = 1
	return
}

func anotherTest() int {
	var x int
	defer func() {
		x++
	}()
	x = 1
	return x
}

func main() {
	fmt.Println(test())
	fmt.Println(anotherTest())
}

// программа выведет 2 \n 1, в функции test отложенное выполнение повлияет на именнованное возращаемое значение, а в функции anotherTest отложенное выполнение изменит созданную локальную переменную x, а не скопированное возращаемое значение
