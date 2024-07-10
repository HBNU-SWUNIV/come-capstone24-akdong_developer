package controllerhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var MapServerEndpoint = "http://10.0.16.70:8070"

func HttpGet(url string) []byte {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// GET 요청 생성
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %s\n", err)
		return nil
	}

	// 필요 시 헤더 추가
	// req.Header.Set("Key", "Value")

	// 요청 보내기
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request failed: %s\n", err)
		return nil
	}
	defer response.Body.Close()

	// 응답 본문 읽기
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err)
		return nil
	}

	// 응답 상태 코드 및 본문 출력
	// fmt.Printf("Response status code: %d\n", response.StatusCode)
	// fmt.Printf("Response body: %s\n", body)

	return body
}

// TODO : 유정이가 MapSavedTime Endpoint 해주면 그거 request 해서
// 새로운 사람이 들어올 때 마다 재현해줘야됨

type MapTimeResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetMapTime(mapid string) MapTimeResponseData {
	body := HttpGet(MapServerEndpoint + "/maptime?mapId=" + mapid)
	var result MapTimeResponseData
	err := json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON response : %s\n", err)
		return MapTimeResponseData{}
	}
	return result
}
