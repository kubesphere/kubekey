package logger

import (
	"sync"
	"testing"
)

var log = NewLogger("", true)

func TestKubeKey_Print(t *testing.T) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		Log.Info("begin")

		log.Info("empty fields")
		l1 := *log
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func(x int, log1 KubeKeyLog) {
				log.Info("Congratulations!", "ssssss")
				wg.Done()
			}(j, l1)
		}
		wg.Wait()
	}

}
