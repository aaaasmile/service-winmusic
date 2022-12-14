package idl

var (
	Appname = "win-music"
	Buildnr = "00.000.01.20221214-00"
)

type StreamProvider interface {
	IsUriForMe(uri string) bool
	GetStatusSleepTime() int
	GetURI() string
	SetURI(string)
	GetTitle() string
	GetDescription() string
	GetPropValue(string) string
	Name() string
	GetStreamerCmd(cmdLineArr []string) (string, string, []string)
	CheckStatus(chDbOperation chan *DbOperation) error
	CreateStopChannel() chan struct{}
	GetCmdStopChannel() chan struct{}
	CloseStopChannel()
	GetTrackDuration() (string, bool)
	GetTrackPosition() (string, bool)
	GetTrackStatus() (string, bool)
}

type DbOpType int

const (
	DbOpHistoryInsert DbOpType = iota
)

type DbOperation struct {
	DbOpType DbOpType
	Payload  interface{}
}
