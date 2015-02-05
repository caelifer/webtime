package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	port = flag.String("port", "8888", "HTTP port number")
)

// Parse our templates
var homeTemplate = template.Must(template.New("time").Parse(htmlTemplate))

func main() {
	flag.Parse()
	host := "localhost"

	// Run HTTP server
	http.HandleFunc("/time/", homeHandler)
	http.HandleFunc("/ws/", wsHandler)
	http.HandleFunc("/", notFoundHandler)

	log.Printf("Running webserver on http://%s:%s/\n", host, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func handleError(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "custom 404")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := homeTemplate.Execute(w, r.Host); err != nil {
		handleError(err, w, r)
	}
}

type timeUpdate struct {
	Time string `json:"time"`
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 256, WriteBufferSize: 256}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &connection{
		send: make(chan []byte, 256),
		ws:   ws,
	}

	go func() {
		for {
			// On every one-second tick
			<-time.Tick(1 * time.Second)

			go func(ts string) {
				msg, err := json.Marshal(&timeUpdate{ts})
				if err != nil {
					log.Println("Bad encoding for", ts, err)
					return
				}
				c.send <- []byte(msg)
			}(time.Now().Local().Format("15:04:05 MST"))
		}
	}()

	go c.writer()
	c.reader()
}

type connection struct {
	send chan []byte
	ws   *websocket.Conn
}

func (c *connection) reader() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		log.Println("[From WS]", msg)
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for msg := range c.send {
		log.Println("[To WS]", string(msg))
		err := c.ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			break
		}
	}
	c.ws.Close()
}

const htmlTemplate = `
<html>
<head>
<title>WebTime via Websocket Example</title>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
<script type="text/javascript">
    (window.onload = function() {

    var conn;
    var timeDiv = $("#webtime");
        var cons = $("#cons");
        var err = $("#error");

        function mylog(msg) {
                // var txt = cons.text() + msg;
                // cons.text(txt);
        }

    function updateTime(msg) {
                // var ts = jQuery.parseJSON(msg);
                var ts = eval( '(' + msg + ')');
                timeDiv.html(ts["time"]);
    }

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://{{$}}/ws/");
        conn.onclose = function(evt) {
            mylog("Connection closed.");
        }
        conn.onmessage = function(evt) {
            mylog(evt.data);
                        updateTime(evt.data);
        }
    } else {
        err.update("Your browser does not support WebSockets.");
    }
    });
</script>
<style type="text/css">
html {
    overflow: hidden;
}

body {
    overflow: hidden;
    padding: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    background: gray;
}

#error {
    background: red;
        color: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    position: absolute;
    top: 0.5em;
    left: 0.5em;
    right: 0.5em;
    // bottom: 3em;
    // overflow: auto;
    visibility: hidden;
}
#cons {
    background: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    position: absolute;
    top: 0.5em;
    left: 0.5em;
    right: 0.5em;
    bottom: 0.5em;
    overflow: auto;
        visibility: hidden;
}

#webtime {
        background: white;
    padding: 0 0.5em 0 0.5em;
    margin: 0;
    position: absolute;
    top: 48%;
    left: 0%;
    width: 100%;
    overflow: hidden;
        font-size: 3em;
        text-align: center;
}

</style>
</head>
        <body>
                <div id="error"></div>
                <div id="webtime"></div>
                <div id="cons"></div>
        </body>
</html>
`
