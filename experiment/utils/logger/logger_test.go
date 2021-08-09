package logger

import (
	"strconv"
	"sync"
	"testing"
)

func TestKubeKey_Print(t *testing.T) {
	log := NewLogger()
	wg := &sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		log.SetModule("CREATE")
		log.SetTask("task1")
		log.SetNode("node" + strconv.Itoa(i))
		l1 := *log
		wg.Add(1)
		go func(x int, log KubeKeyLog) {

			log.Info("Congratulations!", "ssssss")
			wg.Done()
		}(i, l1)
	}
	wg.Wait()
}
