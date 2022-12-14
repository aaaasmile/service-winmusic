package live

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aaaasmile/service-winmusic/conf"
	"github.com/aaaasmile/service-winmusic/db"
	"github.com/aaaasmile/service-winmusic/web/idl"
	"github.com/aaaasmile/service-winmusic/web/live/omx"
	"github.com/aaaasmile/service-winmusic/web/live/omx/omxstate"
)

var (
	player *omx.OmxPlayer
	liteDB *db.LiteDB
)

type PageCtx struct {
	RootUrl    string
	Buildnr    string
	VueLibName string
}

func getURLForRoute(uri string) string {
	arr := strings.Split(uri, "/")
	//fmt.Println("split: ", arr, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		ss := arr[i]
		if ss != "" {
			if !strings.HasPrefix(ss, "?") {
				//fmt.Printf("Url for route is %s\n", ss)
				return ss
			}
		}
	}
	return uri
}

func APiHandler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	log.Println("Request: ", req.RequestURI)
	var err error
	switch req.Method {
	case "GET":
		err = handleGet(w, req)
	case "POST":
		log.Println("POST on ", req.RequestURI)
		err = handlePost(w, req)
	}
	if err != nil {
		log.Println("Error exec: ", err)
		http.Error(w, fmt.Sprintf("Internal error on execute: %v", err), http.StatusInternalServerError)
	}

	t := time.Now()
	elapsed := t.Sub(start)
	log.Printf("Service %s total call duration: %v\n", idl.Appname, elapsed)
}

func handleGet(w http.ResponseWriter, req *http.Request) error {
	u, _ := url.Parse(req.RequestURI)
	log.Println("GET requested ", u)

	pagectx := PageCtx{
		RootUrl:    conf.Current.RootURLPattern,
		Buildnr:    idl.Buildnr,
		VueLibName: conf.Current.VueLibName,
	}
	templName := "templates/vue/index.html"

	tmplIndex := template.Must(template.New("AppIndex").ParseFiles(templName))

	err := tmplIndex.ExecuteTemplate(w, "base", pagectx)
	if err != nil {
		return err
	}
	return nil
}

func writeResponse(w http.ResponseWriter, resp interface{}) error {
	blobresp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.Write(blobresp)
	return nil
}

func writeResponseNoWsBroadcast(w http.ResponseWriter, resp interface{}) error {
	blobresp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.Write(blobresp)
	return nil
}

func writeErrorResponse(w http.ResponseWriter, errorcode int, resp interface{}) error {
	blobresp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	http.Error(w, string(blobresp), errorcode)
	return nil
}

func listenDbOperations(dbCh chan *idl.DbOperation) {
	log.Println("Waiting for history item")
	for {
		item := <-dbCh
		proc := false
		log.Println("Db operation rec ", item.DbOpType)
		if item.DbOpType == idl.DbOpHistoryInsert {
			if vv, ok := item.Payload.(db.ResUriItem); ok {
				proc = true
				if err := liteDB.InsertHistoryItem(&vv); err != nil {
					log.Println("Error on insert history: ", err)
				}
			}
		}

		if !proc {
			log.Println("Db operation not recognized ", item)
		}
	}
}

func InitFromConfig(cmdPlayer *conf.Player, debug bool, dbPath string) error {
	player.SetCommandLine(cmdPlayer)
	liteDB.DebugSQL = debug
	liteDB.SqliteDBPath = dbPath
	if err := liteDB.OpenSqliteDatabase(); err != nil {
		return err
	}
	log.Println("Handler initialized", cmdPlayer, debug, dbPath)
	return nil
}

func HandlerShutdown() {
	chstop := make(chan struct{})
	chTimeout := make(chan struct{})
	timeout := 3 * time.Second
	time.AfterFunc(timeout, func() {
		chTimeout <- struct{}{}
	})

	log.Println("Force poweroff player")
	go func(chst1 chan struct{}) {
		player.PowerOff()
		chst1 <- struct{}{}
	}(chstop)

	count := 2
	select {
	case <-chstop:
		log.Println("Poweroff terminated ok")
		count--
		if count <= 0 {
			log.Println("Shutdown in player ok")
			break
		}
	case <-chTimeout:
		log.Println("Timeout on shutdown, something was blockd")
		break
	}
	log.Println("Exit from HandlerShutdown")
}

func init() {
	dbOpCh := make(chan *idl.DbOperation)
	workers := make([]omxstate.WorkerState, 0)

	chStatus1 := make(chan *omxstate.StateOmx)
	w1 := omxstate.WorkerState{ChStatus: chStatus1}
	workers = append(workers, w1)

	chStatus2 := make(chan *omxstate.StateOmx)
	player = omx.NewOmxPlayer(dbOpCh)
	w2 := omxstate.WorkerState{ChStatus: chStatus2}
	workers = append(workers, w2)
	go player.ListenOmxState(chStatus2)

	liteDB = &db.LiteDB{}

	go listenDbOperations(dbOpCh)
	go omxstate.ListenStateAction(player.ChAction, workers)
}
