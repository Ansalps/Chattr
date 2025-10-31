package response

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

func ClientResponse(statusCode int, message string, data interface{}) Response {
	
	return  Response{
		StatusCode: statusCode,
		Message: message,
		Data: data,
	}
}