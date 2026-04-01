package transport

type ApiError struct {
	StatusCode int    `json:"-"`
	Error      string `json:"error"`
	Details    string `json:"details,omitempty"`
}

func (e *ApiError) AddDetails(details string) *ApiError {
	e.Details = details
	return e
}

var ApiBadRequest = ApiError{Error: "bad request params", StatusCode: 400}
var ApiUnauthorized = ApiError{Error: "authorization error", StatusCode: 401}
var InternalError = ApiError{Error: "internal server error", StatusCode: 500}
