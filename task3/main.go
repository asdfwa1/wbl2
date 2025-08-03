package main

import (
	"fmt"
	"os"
)

func Foo() error {
	var err *os.PathError = nil
	return err
}

func main() {
	err := Foo()
	fmt.Println(err)
	fmt.Println(err == nil)
}

// программа выведет nil \n false, так как созданная переменная err имеет тип *os.PathError, а значение nil. Чтобы вывело true, нужно чтобы тип и значние были равны nil
// Интерфейс это структура из двух полей tab (структура о типе и списке методов) и data (указатель на хранимые данные), а вот пустой интерфейс не определяет никаких методов, поэтому вместо interface table использует сразу type (указатель на тип)
