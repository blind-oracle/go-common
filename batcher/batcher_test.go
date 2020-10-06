package batcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Batcher(t *testing.T) {
	cnt := 0

	f := func(b []interface{}) error {
		for range b {
			cnt++
		}

		return nil
	}

	c := Config{
		Flush: f,
	}

	b, err := New(c)
	assert.Nil(t, err)

	for i := 0; i < 10000; i++ {
		b.Queue(i)
	}

	b.Close()
	assert.Equal(t, 10000, cnt)
}
