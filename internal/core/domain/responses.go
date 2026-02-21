package domain

type ResponseWrapper struct {
	*ResponseOk    `json:"responseOk,omitempty"`
	*ResponseError `json:"responseError,omitempty"`
}
type ResponseOk struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
}
type ResponseError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Data       any    `json:"data"`
}
