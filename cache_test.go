package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	c := New(DefaultConfig())
	value, err := c.Get(func() (interface{}, error) {
		return 1, errors.New("error")
	})
	assert.Equal(t, err, errors.New("error"))
	assert.Nil(t, value)

	value, err = c.Get(func() (interface{}, error) {
		return 1, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, value)

	value, err = c.Get(func() (interface{}, error) {
		return 2, errors.New("error")
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, value)

	value, err = c.Get(func() (interface{}, error) {
		return 3, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, value)
}

func TestLock(t *testing.T) {
	ch := make(chan interface{})
	c := New(DefaultConfig())
	go func() {
		c.Get(func() (interface{}, error) {
			time.Sleep(time.Millisecond * 100)
			return 1, nil
		})
	}()
	go func() {
		time.Sleep(time.Millisecond)
		v, _ := c.Get(func() (interface{}, error) {
			return 2, nil
		})
		ch <- v
	}()
	v := <-ch
	assert.Equal(t, 1, v)
}

func TestLockWithError(t *testing.T) {
	ch := make(chan interface{})
	c := New(DefaultConfig())
	go func() {
		c.Get(func() (interface{}, error) {
			time.Sleep(time.Millisecond * 100)
			return 1, errors.New("error")
		})
	}()
	go func() {
		time.Sleep(time.Millisecond)
		v, _ := c.Get(func() (interface{}, error) {
			return 2, nil
		})
		ch <- v
	}()
	v := <-ch
	assert.Equal(t, 2, v)
}

func TestExpires(t *testing.T) {
	c := New(NewConfig(time.Millisecond*100, 0))
	v, _ := c.Get(func() (interface{}, error) {
		time.Sleep(time.Millisecond * 100)
		return 1, nil
	})
	assert.Equal(t, 1, v)

	time.Sleep(time.Millisecond * 50)
	v, _ = c.Get(func() (interface{}, error) {
		return 2, nil
	})
	assert.Equal(t, 1, v)

	time.Sleep(time.Millisecond * 50)
	v, _ = c.Get(func() (interface{}, error) {
		return 3, nil
	})
	assert.Equal(t, 3, v)

	time.Sleep(time.Millisecond * 50)
	v, _ = c.Get(func() (interface{}, error) {
		return 4, errors.New("error")
	})
	assert.Equal(t, 3, v)

	time.Sleep(time.Millisecond * 50)
	v, err := c.Get(func() (interface{}, error) {
		return 5, errors.New("error")
	})
	assert.Equal(t, errors.New("error"), err)
	assert.Nil(t, v)

	v, _ = c.Get(func() (interface{}, error) {
		return 6, nil
	})
	assert.Equal(t, 6, v)
}

// testing that the values are randomly distributed, so there's a low probability of failure.
func TestJitter(t *testing.T) {
	results := map[int64]int{}
	for i := 0; i < 1000; i++ {
		e := NewConfig(time.Second, time.Second*10).expiresAt()
		sub := int64(e.Sub(time.Now()) / time.Second)
		results[sub]++
	}
	assert.Equal(t, 10, len(results))
	assert.Greater(t, results[1], 20)
	assert.Greater(t, results[2], 20)
	assert.Greater(t, results[3], 20)
	assert.Greater(t, results[4], 20)
	assert.Greater(t, results[5], 20)
	assert.Greater(t, results[6], 20)
	assert.Greater(t, results[7], 20)
	assert.Greater(t, results[8], 20)
	assert.Greater(t, results[9], 20)
	assert.Greater(t, results[10], 20)
}

func TestFixedJitter(t *testing.T) {
	e := NewConfig(time.Second, time.Second*10).expiresAt()
	results := map[time.Duration]int{}
	for i := 0; i < 100; i++ {
		sub := e.Sub(time.Now()) / time.Millisecond
		results[sub]++
	}
	assert.Equal(t, 1, len(results))
}

func TestInvalidate(t *testing.T) {
	c := New(DefaultConfig())
	value, _ := c.Get(func() (interface{}, error) {
		return 1, nil
	})
	assert.Equal(t, 1, value)

	c.Invalidate()

	value, _ = c.Get(func() (interface{}, error) {
		return 2, nil
	})
	assert.Equal(t, 2, value)
}

func TestNoExpiration(t *testing.T) {
	c := New(NewConfig(NoExpiration, 0))
	value, _ := c.Get(func() (interface{}, error) {
		return 1, nil
	})
	assert.Equal(t, 1, value)
	value, _ = c.Get(func() (interface{}, error) {
		return 2, nil
	})
	assert.Equal(t, 1, value)

	c.Invalidate()

	value, _ = c.Get(func() (interface{}, error) {
		return 3, nil
	})
	assert.Equal(t, 3, value)
}
