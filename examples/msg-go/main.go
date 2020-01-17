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
		function msg(){
			var t=document.getElementById('type').value;
			var b=document.getElementById('btn').value;
			external.invoke("m,"+t+","+b);
		}
		function fileopen(){
			external.invoke("open");
		}
		function filesave(){
			external.invoke("save");
		}
		function directoryopen(){
			external.invoke("directory");
		}
		function resp(data){
			document.getElementById('res').value=data;
		}
		function fresp(data){
			document.getElementById('fres').value=data;
		}
		</script>
	</head>
	<body>
	type:
	<select id="type">
		<option value="0">MessageMsg</option>
		<option value="1">MessageInfo</option>
		<option value="2">MessageWarning</option>
		<option value="3">MessageError</option>
	</select>
	<br/>
	button:
	<select id="btn">
		<option value="0">MessageButtonOK</option>
		<option value="4">MessageButtonOKCancel</option>
		<option value="8">MessageButtonYesNo</option>
		<option value="12">MessageButtonYesNoCancel</option>
	</select>
	<button onclick="msg()">ok</button>
	<br/>
	respose<input id="res" value="closable" type="text" />
	<br/>
	file:
	<button onclick="fileopen()">open</button>
	<button onclick="filesave()">save</button>
	<button onclick="directoryopen()">directory</button>
	<br/>
	respose<input id="fres" value="" type="text" />
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
	switch data {
	case "directory":
		resss := w.DirectoryOpen()
		w.Eval("fresp(\"" + strings.ReplaceAll(resss, "\\", "\\\\") + "\");")
	case "open":
		resss := w.FileOpen("*.go;*.rc;*.exe")
		w.Eval("fresp(\"" + strings.ReplaceAll(resss, "\\", "\\\\") + "\");")
	case "save":
		resss := w.FileSave("*.go;*.rc;*.exe")
		w.Eval("fresp(\"" + strings.ReplaceAll(resss, "\\", "\\\\") + "\");")
	default:
		d := strings.Split(data, ",")
		icon, _ := strconv.Atoi(d[1])
		btn, _ := strconv.Atoi(d[2])
		out := w.Message("title", "this is message for test,in msg part", webui.MessageFlag(icon|btn))
		str := strconv.Itoa(int(out))
		w.Eval("resp(\"" + str + "\");")
	}

}

func main() {
	url := startServer()
	w := webui.New(webui.Settings{
		Width:                  windowWidth,
		Height:                 windowHeight,
		Title:                  "message box",
		URL:                    url,
		Debug:                  true,
		ExternalInvokeCallback: handleRPC,
	})
	w.SetMinSize(300, 300)
	w.SetColor(255, 255, 255, 255)
	defer w.Exit()
	w.Run()
}
