package responses

type MapResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MapResponse_map struct {
	Code    int                    `json:"code"`
	Message map[string]interface{} `json:"message"`
}

type MapResponse_list struct {
    Code    int                      `json:"code"`
    Message []map[string]interface{} `json:"message"`
}
