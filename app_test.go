package main

import (
	"reflect"
	"testing"
	"time"
)

func Test(t *testing.T) {
	ans := map[string]int{
		"https://pingocean.com/ru": 168,
		"https://hh.kz":            683,
	}
	var (
		maxWorkers = 8
		needle     = "<div"
		urls       = []string{"https://pingocean.com/ru", "https://hh.kz"}
		timeout    = time.Duration(time.Second * 10)
	)
	output := Run(urls, needle, maxWorkers, timeout)
	if !reflect.DeepEqual(ans, output) {
		t.Errorf("\nExpected: %+v\n Got: %+v\n", ans, output)
	}

}
