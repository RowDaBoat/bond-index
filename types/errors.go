package types

type ErrorCode int

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeNameRequired
	ErrorCodeNameNotFound
	ErrorCodeStoreFailure
	ErrorCodeRoutingNotFound
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

var (
	ErrNameRequired = DomainError{
		Code:    ErrorCodeNameRequired,
		Message: "name is required",
	}

	ErrNameNotFound = DomainError{
		Code:    ErrorCodeNameNotFound,
		Message: "name not found",
	}

	ErrRoutingNotFound = DomainError{
		Code:    ErrorCodeRoutingNotFound,
		Message: "routing not found",
	}
)
