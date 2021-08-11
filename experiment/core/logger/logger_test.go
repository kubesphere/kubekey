package logger

import (
	"strconv"
	"sync"
	"testing"
)

var log = NewLogger()

func TestKubeKey_Print(t *testing.T) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		log.SetModule("CREATE")
		log.SetTask("task1")
		l1 := *log
		wg.Add(1)
		go func(x int, log1 KubeKeyLog) {
			log.SetNode("node" + strconv.Itoa(x))
			log.Info("Congratulations!", "ssssss")
			wg.Done()
		}(i, l1)
	}
	wg.Wait()
}
