package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/srfirouzi/webui"
)

const (
	windowWidth  = 480
	windowHeight = 320
)

var indexHTML = `
<!doctype html>
<html>
	<head>
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<script>
		function run(){
			var t=document.getElementById('type').value;
			var b=document.getElementById('btn').value;
			external.invoke(t+","+b);
		}
		function resp(data){
			document.getElementById('res').value=data;
		}
		</script>
	</head>
	<body>
	type:
	<select id="type">
		<option value="0">WEBUI_MSG_MSG</option>
		<option value="1">WEBUI_MSG_INFO</option>
		<option value="2">WEBUI_MSG_WARNING</option>
		<option value="3">WEBUI_MSG_ERROR</option>
	</select>
	<br/>
	button:
	<select id="btn">
		<option value="0">WEBUI_MSG_OK</option>
		<option value="4">WEBUI_MSG_OK_CANCEL</option>
		<option value="8">WEBUI_MSG_YES_NO</option>
		<option value="12">WEBUI_MSG_YES_NO_CANCEL</option>
	</select>
	<button onclick="run()">ok</button>
	<br/>
	respose<input id="res" value="closable" type="text" />
	</body>
</html>
`

func startServer() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer ln.Close()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(indexHTML))
		})
		log.Fatal(http.Serve(ln, nil))
	}()
	return "http://" + ln.Addr().String()
}

func handleRPC(w webui.WebUI, data string) {
	d := strings.Split(data, ",")
	icon, _ := strconv.Atoi(d[0])
	btn, _ := strconv.Atoi(d[1])
	out := w.Msg(icon, btn, "title", "this is message for test,in msg part")
	str := strconv.Itoa(out)
	w.Eval("resp(\"" + str + "\");")

}

func main() {
	url := startServer()
	w := webui.New(webui.Settings{
		Width:                  windowWidth,
		Height:                 windowHeight,
		Title:                  "message box",
		URL:                    url,
		ExternalInvokeCallback: handleRPC,
	})
	w.SetMinSize(300, 300)
	w.SetColor(255, 255, 255, 255)
	defer w.Exit()
	w.Run()
}
