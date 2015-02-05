package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/caelifer/webtime/client"
	"github.com/caelifer/webtime/service"
)

var (
	port = flag.String("port", "8888", "HTTP port number")
)

// Parse our templates
var homeTemplate = template.Must(template.New("time").Parse(htmlTemplate))

// Start TimeTicker service
var timeTicker = service.NewTimeTicker()

func main() {
	flag.Parse()
	host := "localhost"

	// Setup routes
	http.HandleFunc("/time/", homeHandler)
	http.HandleFunc("/ws/", wsHandler)
	http.HandleFunc("/", notFoundHandler)

	// Run HTTP server
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

func wsHandler(w http.ResponseWriter, r *http.Request) {
	client.NewWSClient(timeTicker).ServeHTTP(w, r)
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
        var ts = jQuery.parseJSON(msg);
        timeDiv.html(ts["time"]);
    }

    if (window["WebSocket"]) {

        // Open Websocket back to the server
        conn = new WebSocket("ws://{{$}}/ws/");

        // Handle WS close
        conn.onclose = function(evt) {
            mylog("Connection closed.");
        }

        // Handle new message from WS
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
