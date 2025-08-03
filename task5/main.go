package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
}

func main() {
	var err error // interface error (type=nil,value=nil)
	err = test()  // присвоение err (type=*customError,value=nil)
	//fmt.Printf("%T, %v\n", err, err)
	if err != nil {
		println("error")
		return
	}
	println("ok")
}

// программа выведет error, так как интерфейс error содержит в поле tab информацию о типе *customError
