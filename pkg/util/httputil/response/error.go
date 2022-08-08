package response

const (
	errorMessageNotFound      = "not found"
	errorCodeNotFound         = 10004
	errorMessageInternalError = "internal error"
	errorInternalError        = 10005
)

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors"`
}

func ErrorNotFound(errors interface{}) *ErrorResponse {
	return &ErrorResponse{
		Code:    errorCodeNotFound,
		Message: errorMessageNotFound,
		Errors:  errors,
	}
}

func ErrorInternalError(errors interface{}) *ErrorResponse {
	return &ErrorResponse{
		Code:    errorInternalError,
		Message: errorMessageInternalError,
		Errors:  errors,
	}
}
