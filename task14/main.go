package main

import (
	"fmt"
	"time"
)

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()

		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}

func or(channels ...<-chan interface{}) <-chan interface{} { // рекурсивный подход с обработкой 3 каналов напрямую уменьшает глубину рекурсии в отличие от обработки 2 канало. ПРИМЕР на вход 5 каналов = 3 явных считывания + рекрсия на оставшиеся 2, а если было бы 2 считывания, то получилось бы 2 явных считывания + рекурсия на 3 считывания далее ещё 2 считывания + рекурсия на оставшийся 1, в итоге глубина рекурсия бы ровнялась log2(N), что глубже чем log3(N).
	switch {
	case len(channels) == 0:
		c := make(chan interface{})
		close(c)
		return c
	case len(channels) == 1:
		return channels[0]
	}

	done := make(chan interface{})
	go func() {
		defer close(done)

		switch {
		case len(channels) == 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-or(append(channels[3:], done)...):
			}
		}
	}()

	return done
}
