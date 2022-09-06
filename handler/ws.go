package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"strings"
	"websocket-splice/models"
)

var (
	upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	C        map[int]*websocket.Conn
	state    bool
	k        int
	Token32  = []byte("mysecretpassword")
	cookie   *models.Cookie
	login    string
	password string
)

const (
	https = "https://"

	splice     = "dev.service.com"
	spliceHost = https + splice

	api             = "api."
	spliceAPI       = api + splice
	httpsSpliceAPI  = https + spliceAPI + "/www"
	httpsSpliceSign = https + spliceAPI + "/www/sign_in"

	spliceCookie = "_splice_web_session"

	host       = "https://dev.mylocalsite.com"
	service    = "https://api.splice.com"
	preset     = "scnova22-"
	authCookie = "X-Authorization"
	lengthCode = 4
)

// JsonApi websocket returns json format
func JsonApi(c *gin.Context) {
	//Upgrade get request to webSocket protocol
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	defer ws.Close()
	if err != nil {
		LogError("error get connection" + err.Error())
		return
	}
	agent := c.Request.Header.Get("User-Agent")
	defer LogCommon(fmt.Sprintf("client disconnected: %s", agent))
	defer DeleteMap()

	u := http.Client{}
	C[len(C)] = ws
	session := c.Param("sess")
	if session == "" {
		return
	}

	LogCommon(fmt.Sprintf("client connected: %s", agent))
	login = c.Param("login")
	password = c.Param("pass")
	if login == "undefined" || login == "" || password == "undefined" || password == "" {
		m := fmt.Sprintf("no splice auth")
		fmt.Println(m)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"message":"%s"}`, m)))
		LogError(m)
		return
	}
	cookie, err = GetCookie(&u, login, password, agent)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"message":"%v"}`, err)))
		LogError(fmt.Sprintf("GetCookie %v", err))
		return
	}
	fmt.Println(cookie.Cookie)

	var account models.InputAccount
	auth, err := ExtractBody(&u, "GET", host+"/api/session/", session, nil)
	if err != nil {
		LogError(err.Error())
		return
	}
	if err := json.Unmarshal(auth, &account); err != nil {
		LogError(err.Error())
		return
	}
	LogCommon(fmt.Sprintf("authorized: %s", shorter(session)))

	//Write and Read ws
	for {
		_, text, err := ws.ReadMessage()

		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				return
			} else {
				LogError(err.Error())
				return
			}
		}

		switch string(text) {
		case "start":
			if state {
				ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"started already"}`))
			} else {
				state = true
				client := &http.Client{}
				go func() {
					for {
						if !state {
							ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"stopped"}`))
							return
						}

						k++
						startCommand(client, k, session, agent)
					}
				}()
			}
		case "stop":
			if state {
				state = false

				//stopCommand(ws)
				ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"shutdown"}`))
			} else {
				ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"not started. use start"}`))
			}
		case "ping":
			ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"pong"}`))
		default:
			ws.WriteMessage(websocket.TextMessage, []byte(`{"message":"invalid command. try start or stop"}`))
		}
	}
}

func SendMessage(msg string) {
	for _, e := range C {
		_ = e.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"loop":%d,"message":"%s"}`, k, msg)))
	}
}

func DeleteMap() {
	i := len(C)
	if i < 1 {
		LogError("websocket: amount of connection less zero")
		return
	}
	delete(C, i)
}

func GetCookie(u *http.Client, login, password, agent string) (*models.Cookie, error) {
	var cookie models.Cookie
	values := map[string]io.Reader{
		"login":    strings.NewReader(login),
		"password": strings.NewReader(password),
	}
	res, err := LoginSplice(u, httpsSpliceSign, agent, values)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%d: %v", res.StatusCode, err))
	}
	if res.StatusCode == 429 {
		return nil, errors.New("too many requests")
	}
	for _, e := range res.Cookies() {
		if e.Name == spliceCookie {
			cookie.Cookie = e.Value
			break
		}
	}
	if cookie.Cookie == "" {
		return nil, errors.New("no such cookie")
	}
	return &cookie, nil
}
