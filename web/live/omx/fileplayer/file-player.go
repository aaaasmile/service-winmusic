package fileplayer

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aaaasmile/service-winmusic/conf"
	"github.com/aaaasmile/service-winmusic/db"
	"github.com/aaaasmile/service-winmusic/web/idl"
	"github.com/aaaasmile/service-winmusic/web/live/omx/dbus"
	"github.com/aaaasmile/service-winmusic/web/live/omx/omxstate"
)

type infoFile struct {
	Title         string
	Description   string
	DurationInSec int
	TrackDuration string
	TrackPosition string
	TrackStatus   string
}

type FilePlayer struct {
	URI     string
	Info    *infoFile
	Dbus    *dbus.OmxDbus
	chClose chan struct{}
}

func (fp *FilePlayer) IsUriForMe(uri string) bool {
	if strings.Contains(uri, "/home") &&
		(strings.Contains(uri, ".mp4") || strings.Contains(uri, ".avi") ||
			strings.Contains(uri, ".mp3") || strings.Contains(uri, ".ogg") || strings.Contains(uri, ".wav")) {
		log.Println("this is a file for omx ", uri)
		fp.URI = uri
		return true
	}
	return false
}

func (fp *FilePlayer) GetStatusSleepTime() int {
	return 300
}

func (fp *FilePlayer) GetURI() string {
	return fp.URI
}

func (rp *FilePlayer) SetURI(uri string) {
	rp.URI = uri
}

func (fp *FilePlayer) GetTitle() string {
	if fp.Info != nil {
		return fp.Info.Title
	}
	return ""
}

func (rp *FilePlayer) GetPropValue(propname string) string {
	return ""
}

func (fp *FilePlayer) GetDescription() string {
	if fp.Info != nil {
		return fp.Info.Description
	}
	return ""
}
func (fp *FilePlayer) Name() string {
	return "file"
}
func (fp *FilePlayer) GetStreamerCmd(cmdLineArr []string) (string, string, []string) {
	moreargs := []string{}
	moreargs = append(moreargs, fmt.Sprintf("\"%s\"", fp.URI))
	args := strings.Join(cmdLineArr, " ")
	//cmd := fmt.Sprintf(`omxplayer %s "%s"`, args, fp.URI)
	//params := fmt.Sprintf(`%s `, args)
	return conf.Current.Player.Path, args, moreargs
}
func (fp *FilePlayer) CheckStatus(chDbOperation chan *idl.DbOperation) error {
	st := &omxstate.StateOmx{}
	if err := fp.Dbus.CheckTrackStatus(st); err != nil {
		return err
	}

	if fp.Info == nil {
		titles := strings.Split(fp.URI, "/")
		title := ""
		if len(titles) > 0 {
			title = titles[len(titles)-1]
		}
		info := infoFile{
			// TODO read from db using URI -> see player radio
			Title: title,
		}
		info.DurationInSec, _ = strconv.Atoi(st.TrackDuration)
		if info.DurationInSec > 0 {
			info.TrackDuration = time.Duration(int64(info.DurationInSec) * int64(time.Second)).String()
		}

		hi := db.ResUriItem{
			URI:           fp.URI,
			Title:         info.Title,
			Description:   info.Description,
			DurationInSec: info.DurationInSec,
			Type:          fp.Name(),
			Duration:      info.TrackDuration,
		}
		dop := idl.DbOperation{
			DbOpType: idl.DbOpHistoryInsert,
			Payload:  hi,
		}
		chDbOperation <- &dop
		fp.Info = &info
		log.Println("file-player info status set")
	}

	fp.Info.TrackPosition = st.TrackPosition
	fp.Info.TrackStatus = st.TrackStatus
	log.Println("Status set to ", fp.Info)
	return nil
}

func (fp *FilePlayer) CreateStopChannel() chan struct{} {
	if fp.chClose == nil {
		fp.chClose = make(chan struct{})
	}
	return fp.chClose
}

func (fp *FilePlayer) GetCmdStopChannel() chan struct{} {
	return fp.chClose
}

func (fp *FilePlayer) CloseStopChannel() {
	if fp.chClose != nil {
		close(fp.chClose)
		fp.chClose = nil
	}
}

func (fp *FilePlayer) GetTrackDuration() (string, bool) {
	if fp.Info != nil {
		return fp.Info.TrackDuration, true
	}
	return "", false

}
func (fp *FilePlayer) GetTrackPosition() (string, bool) {
	if fp.Info != nil {
		return fp.Info.TrackPosition, true
	}
	return "", false

}
func (fp *FilePlayer) GetTrackStatus() (string, bool) {
	if fp.Info != nil {
		return fp.Info.TrackStatus, true
	}
	return "", false
}
