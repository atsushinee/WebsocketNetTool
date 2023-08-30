package websocket



type MessageType int


const (
	
	MessageText MessageType = iota + 1
	
	MessageBinary
)
