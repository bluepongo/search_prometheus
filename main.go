package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	BASE_URL   = "http://192.168.10.215:31626/api/v1/"
	SEARCH_RES = "tikv_thread_cpu_seconds%3A1m"
	//SEARCH_RES = "tikv_engine_memory_bytes"
	TIME_DUR  = 60
	LOG_FILE  = "./log/tikv.log"
	LIMIT     = 100.1
	INDICATOR = "CPU"
	//INDICATOR = "Memory"
)

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func WriteFile(path string, write string) {
	var f *os.File
	var err error
	if checkFileIsExist(path) { //如果文件存在
		f, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend) //打开文件
	} else {
		f, err = os.Create(path) //创建文件
		fmt.Println("文件不存在, 已创建文件" + path)
	}
	if err != nil {
		fmt.Println("create log file error", err)
		return
	}

	_, err = io.WriteString(f, write)
	if err != nil {
		fmt.Println("write log file error", err)
	}
	f.Close()
}

type Res struct {
	Status interface{} `json:"status"`
	Data   *Data       `json:"data"`
}
type Data struct {
	ResultType interface{}   `json:"resultType"`
	Result     []interface{} `json:"result"`
}
type Result struct {
	Metric *Metric         `json:"metric"`
	Values [][]interface{} `json:"values"`
}
type Metric struct {
	Instance string `json:"instance"`
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func main() {
	startUnix := time.Now().Add(-time.Duration(TIME_DUR) * time.Minute).Unix()
	endUnix := time.Now().Unix()
	//startUnix := time.Now().Add(-time.Duration(1455) * time.Minute).Unix()
	//endUnix := time.Now().Add(-time.Duration(180) * time.Minute).Unix()
	// time := time.Now().Add(- time.Duration(10) * time.Minute)
	//fmt.Println(strconv.FormatInt(startUnix, 10))
	url := BASE_URL + "query_range?query=" + SEARCH_RES + "&start=" + strconv.FormatInt(startUnix, 10) + "&end=" + strconv.FormatInt(endUnix, 10) + "&step=14&_=1618463290327"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	res := Res{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("json err")
		return
	}
	for _, val := range res.Data.Result {
		result := Result{}
		arr, err := json.Marshal(val)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(arr, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, r := range result.Values {
			cpuPercent, _ := strconv.ParseFloat(r[1].(string), 64)
			cpuPercent *= 100
			fmt.Printf("%.2f%% \t", cpuPercent)
			timeLayout := "2006-01-02 15:04:05"
			fmt.Println(time.Unix(int64(r[0].(float64)), 0).Format(timeLayout))
			if cpuPercent > LIMIT {
				time := time.Unix(int64(r[0].(float64)), 0).Format(timeLayout)
				altStr := "[ALERT][" + time + "][" + INDICATOR + "] exceeds the specified limit, and the specific value is: " + strconv.FormatFloat(cpuPercent, 'f', 2, 32) + "%.\n"
				WriteFile(LOG_FILE, altStr)
			}
		}
		//if result.Metric.Instance == "ltidb-cluster-01-tikv-0" {
		//	fmt.Printf("mode:%s value:%s\n", result.Metric.Mode, result.Value[1])
		//}
	}
}
