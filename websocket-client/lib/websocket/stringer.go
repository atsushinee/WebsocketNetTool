

package websocket

import "strconv"

func _() {
	
	
	var x [1]struct{}
	_ = x[opContinuation-0]
	_ = x[opText-1]
	_ = x[opBinary-2]
	_ = x[opClose-8]
	_ = x[opPing-9]
	_ = x[opPong-10]
}

const (
	_opcode_name_0 = "opContinuationopTextopBinary"
	_opcode_name_1 = "opCloseopPingopPong"
)

var (
	_opcode_index_0 = [...]uint8{0, 14, 20, 28}
	_opcode_index_1 = [...]uint8{0, 7, 13, 19}
)

func (i opcode) String() string {
	switch {
	case 0 <= i && i <= 2:
		return _opcode_name_0[_opcode_index_0[i]:_opcode_index_0[i+1]]
	case 8 <= i && i <= 10:
		i -= 8
		return _opcode_name_1[_opcode_index_1[i]:_opcode_index_1[i+1]]
	default:
		return "opcode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
func _() {
	
	
	var x [1]struct{}
	_ = x[MessageText-1]
	_ = x[MessageBinary-2]
}

const _MessageType_name = "MessageTextMessageBinary"

var _MessageType_index = [...]uint8{0, 11, 24}

func (i MessageType) String() string {
	i -= 1
	if i < 0 || i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
func _() {
	
	
	var x [1]struct{}
	_ = x[StatusNormalClosure-1000]
	_ = x[StatusGoingAway-1001]
	_ = x[StatusProtocolError-1002]
	_ = x[StatusUnsupportedData-1003]
	_ = x[statusReserved-1004]
	_ = x[StatusNoStatusRcvd-1005]
	_ = x[StatusAbnormalClosure-1006]
	_ = x[StatusInvalidFramePayloadData-1007]
	_ = x[StatusPolicyViolation-1008]
	_ = x[StatusMessageTooBig-1009]
	_ = x[StatusMandatoryExtension-1010]
	_ = x[StatusInternalError-1011]
	_ = x[StatusServiceRestart-1012]
	_ = x[StatusTryAgainLater-1013]
	_ = x[StatusBadGateway-1014]
	_ = x[StatusTLSHandshake-1015]
}

const _StatusCode_name = "StatusNormalClosureStatusGoingAwayStatusProtocolErrorStatusUnsupportedDatastatusReservedStatusNoStatusRcvdStatusAbnormalClosureStatusInvalidFramePayloadDataStatusPolicyViolationStatusMessageTooBigStatusMandatoryExtensionStatusInternalErrorStatusServiceRestartStatusTryAgainLaterStatusBadGatewayStatusTLSHandshake"

var _StatusCode_index = [...]uint16{0, 19, 34, 53, 74, 88, 106, 127, 156, 177, 196, 220, 239, 259, 278, 294, 312}

func (i StatusCode) String() string {
	i -= 1000
	if i < 0 || i >= StatusCode(len(_StatusCode_index)-1) {
		return "StatusCode(" + strconv.FormatInt(int64(i+1000), 10) + ")"
	}
	return _StatusCode_name[_StatusCode_index[i]:_StatusCode_index[i+1]]
}
