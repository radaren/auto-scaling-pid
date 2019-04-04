package main

import (
	"time"
	"net/http"
	"math"
	"math/rand"
)

// 模拟真实请求，设置timeout=2s
func timeoutRequest(){
	client := http.Client{
		Timeout: time.Duration(2 * time.Second),
	}
	client.Get("http://10.68.197.164:80")
}

func genBatchRequest(requestNum int64){
	for i:=int64(0); i < requestNum; i++{
		go timeoutRequest()
	}
}

// Recomment: min=2,max=8
func rectGenerator(min, max int64)func(int64)int64{
	lastRequestRate := int64(0)
	randomRange     := max - min
	randomBottom    := min
	rand.Seed(1) // 多次试验要求同一输出
	return func (timeStamp int64)int64{
		if timeStamp % 100 == 0{
			lastRequestRate = rand.Int63() % randomRange + randomBottom
		}
		genBatchRequest(lastRequestRate)
		return lastRequestRate
	}
}

// Recommend: A=5, w=0.1
func sinGenerator(A, w float64)func(int64)int64{
	return func (timeStamp int64) int64{
		requestNum := int64(math.Ceil(A * (math.Sin(w * float64(timeStamp)) + 1)))
		genBatchRequest(requestNum)
		return requestNum
	}
}

// Recommend: A=0.8, w=0.001
func expGenerator(A, w float64)func(int64)int64{
	return func (timeStamp int64) int64{
		requestNum := int64(math.Ceil(A * math.Exp(w * float64(timeStamp))))
		genBatchRequest(requestNum)
		return requestNum
	}
}

// Recommend: A=2, w=2
func logGenerator(A, w float64)func(int64)int64{
	return func (timeStamp int64) int64{
		requestNum := int64(math.Ceil(A * math.Log(w * float64(timeStamp))))
		genBatchRequest(requestNum)
		return requestNum
	}
}
