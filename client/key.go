package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

type APIKey struct {
	CallPerDay    int    `json:"call_per_day"`
	CallPerSecond int    `json:"call_per_second"`
	Key           string `json:"key"`

	leftForDay    int
	leftForSecond int

	lock *sync.Mutex

	secondTicker *time.Ticker
	dayTicker    *time.Ticker
	die          chan interface{}
}

func LoadAPIKeysFromFile(filename string) ([]APIKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file, %v", err)
	}
	var keys []APIKey
	err = json.Unmarshal(data, &keys)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling failed, %v, payload `%s`", err, string(data))
	}
	for i := range keys {
		keys[i].init()
	}
	return keys, nil
}

func (a *APIKey) init() {
	a.leftForDay = a.CallPerDay
	a.leftForSecond = a.CallPerSecond
	a.lock = new(sync.Mutex)
	a.die = make(chan interface{})

	perDay := time.Hour * time.Duration(a.CallPerDay) * 24
	perSecond := time.Second * time.Duration(a.CallPerSecond)

	a.dayTicker = time.NewTicker(perDay)
	a.secondTicker = time.NewTicker(perSecond)

	go func() {
		for {
			select {
			case <-a.dayTicker.C:
				a.replenishDay()
			case <-a.secondTicker.C:
				a.replenishSecond()
			case <-a.die:
				return
			}
		}
	}()
}

func BuildKey(key string, callPerDay, callPerSecond int) *APIKey {
	a := APIKey{
		Key:           key,
		CallPerDay:    callPerDay,
		CallPerSecond: callPerSecond,
	}
	a.init()

	return &a
}

func (a *APIKey) Use() (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.leftForSecond <= 0 {
		return "", errors.New("throttled for second usage")
	}

	if a.leftForDay <= 0 {
		return "", errors.New("throttled for day usage")
	}

	a.leftForDay--
	a.leftForSecond--

	return a.Key, nil
}

func (a *APIKey) LeftForDay() int {
	return a.leftForDay
}

func (a *APIKey) LeftForSecond() int {
	return a.leftForSecond
}

func (a *APIKey) HasLeft() bool {
	return a.LeftForDay() > 0 && a.LeftForSecond() > 0
}

func (a *APIKey) replenishDay() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.leftForDay = a.CallPerDay
}

func (a *APIKey) replenishSecond() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.leftForSecond = a.CallPerSecond
}

func (a *APIKey) Close() {
	a.dayTicker.Stop()
	a.secondTicker.Stop()
	a.die <- true
}
