package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type inputDataStruct struct {
	url           []string
	count         int
	timeOut       int
	dataResponses map[string][]float64
	noResponses   map[string]int
}

func getResponse(urlChan chan string, inputData inputDataStruct) {
	url := <-urlChan
	start := time.Now()
	client := http.Client{
		Timeout: time.Duration(inputData.timeOut) * time.Second,
	}
	result, err := client.Get(url)
	if err != nil {
		inputData.noResponses[url] = inputData.noResponses[url] + 1
		log.Fatal(err)
	}
	elapsed := time.Since(start).Seconds()
	defer result.Body.Close()
	s := fmt.Sprintf("%s %f", url, elapsed)
	log.Println(s)
	if result.StatusCode == http.StatusOK {
		appendResponse(url, elapsed, inputData)
	}
}

func appendResponse(url string, time float64, inputData inputDataStruct) {
	inputData.dataResponses[url] = append(inputData.dataResponses[url], time)
}

func httpRequest(inputData inputDataStruct) {
	urlChan := make(chan string)
	count := inputData.count
	j := 0
	var wg sync.WaitGroup
	for i := range inputData.url {
		for j < count {
			wg.Add(1)
			go func() {
				getResponse(urlChan, inputData)
				wg.Done()
			}()
			urlChan <- inputData.url[i]
			j++
		}
		j = 0
	}
	wg.Wait()

}

func parseArgument(item string) []string {
	return strings.Split(item, ",")
}
func findMinMaxAvg(values []float64) (min float64, max float64, avg float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	min = values[0]
	max = values[0]
	var sum float64 = 0
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum = sum + v
	}
	var count = float64(len(values))
	avg = sum / count
	return min, max, avg
}

func printMinMaxAvg(dataResponses map[string][]float64) {
	for key, _ := range dataResponses {
		min, max, avg := findMinMaxAvg(dataResponses[key])
		fmt.Printf("url: %s, min: %f, max: %f, avg: %f \n", key, min, max, avg)

	}
}

func main() {
	url := flag.String("url", "", "url.")
	count := flag.Int("count", 1, "count response.")
	timeout := flag.Int("timeout", 1, "timeout response.")
	flag.Parse()
	var inputData inputDataStruct
	inputData.count = *count
	inputData.timeOut = *timeout
	inputData.dataResponses = map[string][]float64{}
	inputData.noResponses = map[string]int{}
	inputData.url = parseArgument(*url)
	start := time.Now()
	httpRequest(inputData)
	fmt.Println("main")
	elapsed := time.Since(start).Seconds()
	fmt.Println(inputData.dataResponses)
	printMinMaxAvg(inputData.dataResponses)
	fmt.Printf("Total time: %f \n", elapsed)
	fmt.Println(inputData.noResponses)
	fmt.Scanln()
}
