package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	BASE_URL = "http://192.168.59.4:9000/prometheus/api/v1/"
	NODE_CPU = "node_cpu"
)

type Res struct {
	Status  interface{} `json:"status"`
	Data *Data `json:"data"`
}
type Data struct {
	ResultType interface{} `json:"resultType"`
	Result []interface{} `json:"result"`
}
type Result struct {
	Metric *Metric
	Value []interface{}
}
type Metric struct {
	Name string
	Cpu string
	Instance string
	Job string
	Mode string
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func main(){
	timeUnix := time.Now().Unix()
	url := BASE_URL + "query?query=" + NODE_CPU + "&time=" + strconv.FormatInt(timeUnix, 10)
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
	if err != nil{
		fmt.Println("json err")
		return
	}
	for _, val := range res.Data.Result{
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
		if(result.Metric.Instance == "localhost.localdomain"){
			fmt.Printf("mode:%s value:%s\n", result.Metric.Mode, result.Value[1])
		}
	}


}