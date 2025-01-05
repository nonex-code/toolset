package gpool

import (
	"testing"
)

func TestGpool(t *testing.T) {
	pool := NewTaskPool(10)
	for i := 0; i < 100; i++ {
		v := i
		task := func() {
			t.Log(v)
			//time.Sleep(time.Second * 1)
		}
		err := pool.Submit(task)
		if err != nil {
			t.Error(err)
			return
		}

	}
	pool.Close()
}
