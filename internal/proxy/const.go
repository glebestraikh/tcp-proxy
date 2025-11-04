package proxy

// Версия SOCKS5
const (
	SOCKS5Version byte = 0x05
)

// Команды
const (
	CmdConnect byte = 0x01
)

// Типы адресов
const (
	AddrTypeIPv4   byte = 0x01
	AddrTypeDomain byte = 0x03
)

// Методы аутентификации
const (
	AuthNone     byte = 0x00
	AuthNoAccept byte = 0xFF
)

// Ответы на команды
const (
	RepSuccess              byte = 0x00
	RepGeneralFailure       byte = 0x01
	RepConnectionNotAllowed byte = 0x02
	RepNetworkUnreachable   byte = 0x03
	RepHostUnreachable      byte = 0x04
	RepConnectionRefused    byte = 0x05
	RepTTLExpired           byte = 0x06
	RepCommandNotSupported  byte = 0x07
	RepAddrTypeNotSupported byte = 0x08
)

// Зарезервированый байт
const (
	Rsv byte = 0x00
)
