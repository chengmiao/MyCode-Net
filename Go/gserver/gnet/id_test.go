package gnet

import (
	"testing"
)

func TestGenID(t *testing.T) {
	goCount := 10

	ids := make(map[int64]struct{})
	idsChan := make(chan int64, 1000000)
	ntfChan := make(chan bool, goCount)
	flags := make([]bool, goCount)
	for i := 0; i < goCount; i++ {
		flags[i] = false
		go func(index int) {
			for n := 0; n < 16384; n++ {
				id := createSessionID()
				idsChan <- id
			}
			flags[index] = true
			ntfChan <- true
		}(i)
	}
	for {
		select {
		case <-ntfChan:
			for i := 0; i < goCount; i++ {
				if !flags[i] {
					goto SelectEnd
				}
			}
			goto ForEnd
		case id := <-idsChan:
			_, ok := ids[id]
			if ok {
				t.Errorf("Found exited id: %v", id)
				t.FailNow()
			}
			ids[id] = struct{}{}
		}
	SelectEnd:
	}
ForEnd:
	close(idsChan)
	close(ntfChan)
}

// func Benchmark_GenID(b *testing.B) {
// 	gen := sdk.NewSequenceGenerator(2)
// 	var wg sync.WaitGroup
// 	for n := 0; n < 10000; n++ {
// 		wg.Add(1)
// 		go func() {
// 			for i := 0; i < b.N; i++ {
// 				gen.NextID()
// 			}
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// }
