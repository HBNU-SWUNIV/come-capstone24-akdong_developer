package client

import (
    "encoding/json"
    "net/http"
)

type RunResponse struct {
    Status string `json:"status"`
}

func ParseResponse(resp *http.Response) (*RunResponse, error) {
    var response RunResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    return &response, nil
}
