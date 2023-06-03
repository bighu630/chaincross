package p2p

const (
	UNLOCK int = iota // 第一次连接时，向连接方打招呼
	LOCK
	UNLOCKED
	MSG
)

type Msg struct {
	Type int
	Data []byte
}

type RemoteLock struct {
	ChainID  string
	AproveID string
	Hs       string
}

type RemoteUnlock struct {
	ChainID string
	LockID  string
	S       string
}
