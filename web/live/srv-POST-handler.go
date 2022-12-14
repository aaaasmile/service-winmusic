package live

import (
	"fmt"
	"log"
	"net/http"
)

func handlePost(w http.ResponseWriter, req *http.Request) error {
	var err error
	lastPath := getURLForRoute(req.RequestURI)
	log.Println("Check the last path ", lastPath)
	switch lastPath {
	case "Resume":
		err = handlePauseOrResume(w, req, player, "Resume")
	case "Pause":
		err = handlePauseOrResume(w, req, player, "Pause")
	case "ChangeVolume":
		err = handleChangeVolume(w, req, player)
	case "SetPowerState":
		err = handleSetPowerState(w, req, player)
	case "GetPlayerState":
		err = handlePlayerState(w, req, player)
	case "NextTitle":
		err = handleNextTitle(w, req, player)
	case "PreviousTitle":
		err = handlePreviousTitle(w, req, player)
	case "PlayUri":
		err = handlePlayUri(w, req, player)
	case "FetchHistory":
		err = handleHistoryRequest(w, req)
	case "HandleMusic":
		err = handleMusicRequest(w, req)
	case "HandleRadio":
		err = handleRadioRequest(w, req)
	default:
		return fmt.Errorf("%s method is not supported", lastPath)
	}

	return err
}
