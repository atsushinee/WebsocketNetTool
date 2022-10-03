package websocket

import (
	"errors"
	"fmt"
)



type StatusCode int








const (
	StatusNormalClosure   StatusCode = 1000
	StatusGoingAway       StatusCode = 1001
	StatusProtocolError   StatusCode = 1002
	StatusUnsupportedData StatusCode = 1003

	
	statusReserved StatusCode = 1004

	
	
	
	StatusNoStatusRcvd StatusCode = 1005

	
	
	
	StatusAbnormalClosure StatusCode = 1006

	StatusInvalidFramePayloadData StatusCode = 1007
	StatusPolicyViolation         StatusCode = 1008
	StatusMessageTooBig           StatusCode = 1009
	StatusMandatoryExtension      StatusCode = 1010
	StatusInternalError           StatusCode = 1011
	StatusServiceRestart          StatusCode = 1012
	StatusTryAgainLater           StatusCode = 1013
	StatusBadGateway              StatusCode = 1014

	
	
	
	StatusTLSHandshake StatusCode = 1015
)





type CloseError struct {
	Code   StatusCode
	Reason string
}

func (ce CloseError) Error() string {
	return fmt.Sprintf("status = %v and reason = %q", ce.Code, ce.Reason)
}





func CloseStatus(err error) StatusCode {
	var ce CloseError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return -1
}
