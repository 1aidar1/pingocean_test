package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Item struct {
	url    string
	needle string
}

const DEFAULT_TIMEOUT = 4000
const DEFAULT_NEEDLE = ""

func main() {
	timer := time.Now()
	var (
		needle     string
		urls       []string
		timeoutInt int
		maxWorkers int
	)
	flag.StringVar(&needle, "needle", DEFAULT_NEEDLE, "searchable string")
	flag.IntVar(&timeoutInt, "time", DEFAULT_TIMEOUT, "timeout for each worker in milliseconds")
	flag.IntVar(&maxWorkers, "workers", runtime.NumCPU(), "max parrallel routines")

	flag.Parse()
	urls = flag.Args()

	timeout := time.Millisecond * time.Duration(timeoutInt)

	output := Run(urls, needle, maxWorkers, timeout)
	for k, v := range output {
		fmt.Printf("%s - %d\n", k, v)
	}

	fmt.Println("finished in", time.Since(timer))

}
func Run(urls []string, needle string, maxWorkers int, timeout time.Duration) map[string]int {
	done := make(chan bool)
	defer close(done)

	itemChan := Prepare(done, urls, needle)
	workers := make([]<-chan map[string]int, maxWorkers)
	errs := make(chan error)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = Work(done, itemChan, timeout, errs)
	}

	go func() {
		for e := range errs {
			//some error handling
			fmt.Println(e)
		}
	}()
	ans := make(map[string]int)
	for j := range MergeWorks(done, workers...) {
		for k, v := range j {
			ans[k] = v
		}
	}
	return ans
}

func Work(done <-chan bool, items <-chan Item, timeout time.Duration, errChan chan error) <-chan map[string]int {

	counts := make(chan map[string]int)

	transfer := make(chan interface{})

	count := func(needle, url string) (int, error) {
		c := http.Client{}
		res, err := c.Get(url)
		if err != nil {
			return -1, err
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return -1, err
		}
		count := strings.Count(string(body), needle)
		return count, nil
	}
	go func() {
		for item := range items {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// start := time.Now()

			// start counting in routine
			go func(itemV Item) {
				occurences, err := count(itemV.needle, itemV.url)
				if err != nil {
					transfer <- err
					return
				}
				transfer <- occurences
			}(item)

			//w8 for signal
			select {
			case <-done:
				cancel()
				return
			case mp := <-transfer:
				switch val := mp.(type) {
				case int:
					counts <- map[string]int{item.url: val}
					// fmt.Println(item.url, time.Since(start))
				case error:
					errChan <- errors.New(fmt.Sprint(item.url, " [ERROR] ", val))
				}
			case <-ctx.Done():
				errChan <- errors.New(fmt.Sprint(item.url, "[TIMEOUT]"))
			}
			cancel()
		}
		close(counts)
	}()
	return counts
}

func MergeWorks(done <-chan bool, channels ...<-chan map[string]int) <-chan map[string]int {
	var wg sync.WaitGroup
	wg.Add(len(channels))
	out := make(chan map[string]int)
	muliplex := func(c <-chan map[string]int) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case out <- i:
			}
		}
	}
	for _, c := range channels {
		go muliplex(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func Prepare(done <-chan bool, urls []string, needle string) <-chan Item {
	items := make(chan Item)
	go func() {
		for _, url := range urls {
			item := Item{
				url:    url,
				needle: needle,
			}
			select {
			case <-done:
				return
			case items <- item:
			}
		}
		close(items)
	}()
	return items
}
