package http

type ResponseFormatter struct {
	Data interface{} 	`json:"data,omitempty"`
	Status int			`json:"http_status,omitempty"`
	Errors interface{}	`json:"errors,omitempty"`
}
