package response

const (
	SuccessCode = 0
)

type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result    interface{} `json:"result,omitempty"`
}

func NewSuccessResponse(result interface{}) *SuccessResponse {
	return &SuccessResponse{
		Code:    SuccessCode,
		Message: "success",
		Result:    result,
	}
}
