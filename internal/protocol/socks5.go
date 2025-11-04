package protocol

const (
	SOCKS5Version byte = 0x05
)

const (
	AuthNone     byte = 0x00
	AuthNoAccept byte = 0xFF
)

const (
	CmdConnect byte = 0x01
)

const (
	AddrTypeIPv4   byte = 0x01
	AddrTypeDomain byte = 0x03
)

const (
	RepSuccess              byte = 0x00
	RepGeneralFailure       byte = 0x01
	RepConnectionNotAllowed byte = 0x02 // no use
	RepNetworkUnreachable   byte = 0x03
	RepHostUnreachable      byte = 0x04
	RepConnectionRefused    byte = 0x05
	RepTTLExpired           byte = 0x06
	RepCommandNotSupported  byte = 0x07
	RepAddrTypeNotSupported byte = 0x08
)

const (
	Reserved byte = 0x00
)

type Request struct {
	Version  byte
	Command  byte
	Reserved byte
	AddrType byte
	DstAddr  string
	DstPort  uint16
}

type Reply struct {
	Version  byte
	Reply    byte
	Reserved byte
	AddrType byte
	BndAddr  []byte
	BndPort  uint16
}
