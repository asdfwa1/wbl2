package main

import (
	"fmt"
	"math/rand"
	"time"
)

func asChan(vs ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, v := range vs {
			c <- v
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
		close(c)
	}()
	return c
}

func merge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		for {
			select {
			case v, ok := <-a:
				if ok {
					c <- v
				} else {
					a = nil
				}
			case v, ok := <-b:
				if ok {
					c <- v
				} else {
					b = nil
				}
			}
			if a == nil && b == nil {
				close(c)
				return
			}
		}
	}()
	return c
}

func main() {
	rand.Seed(time.Now().Unix())
	a := asChan(1, 3, 5, 7)
	b := asChan(2, 4, 6, 8)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v)
	}
}

// программа выведет 12354768 или 21345678 или любую другую последовательность в случайном порядке из цифр [1,8] из-за случайной задержки в функции asChan.
// В этой программе используется конвейер из этапа генерации данных, этапа объединения и этапа использования (вывода в main). Select позволяет неблокирующим образом читать данные из нескольких каналов, обеспечивая эффективное объединение данных, то есть одна горутина слушает несколько каналов и реагирует на тот из которого пришло значение первым
// И ещё в каждой ветке select проверяется условие на открытый и закрытый канал, и если канал закрыт, то переменной канала присваивается nil значение, чтобы больше не попадать в эту ветку select и не заблокироваться
