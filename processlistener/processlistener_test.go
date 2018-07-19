package processlistener

import (
	"sync"
	"testing"
)

const (
	concurrencyLevel = 5
)

func BenchmarkListener(b *testing.B) {
	var notification addNotification

	var swg sync.WaitGroup
	swg.Add(b.N)
	b.SetParallelism(concurrencyLevel)

	pl := NewProcessListener(&ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			swg.Done()
		},
	})
	var wg WaitGroup
	defer wg.Wait()
	defer close(pl.addCh)
	wg.Start(pl.run)
	wg.Start(pl.pop)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pl.add(notification)
		}
	})
	swg.Wait()
	b.StopTimer()
}
