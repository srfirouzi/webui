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
	</head>
	<body>
		<button onclick="external.invoke('close')">Close</button>
		<button onclick="external.invoke('fullscreen')">Fullscreen</button>
		<button onclick="external.invoke('unfullscreen')">Unfullscreen</button>
		<button onclick="external.invoke('open')">Open</button>
		<button onclick="external.invoke('opendir')">Open directory</button>
		<button onclick="external.invoke('save')">Save</button>
		<button onclick="external.invoke('message')">Message</button>
		<button onclick="external.invoke('info')">Info</button>
		<button onclick="external.invoke('warning')">Warning</button>
		<button onclick="external.invoke('error')">Error</button>
		<button onclick="external.invoke('changeTitle:'+document.getElementById('new-title').value)">
			Change title
		</button>
		<input id="new-title" type="text" />
		<button onclick="external.invoke('changeColor:'+document.getElementById('new-color').value)">
			Change color
		</button>
		<input id="new-color" value="#e91e63" type="color" />
		<button onclick="external.invoke('winClose');document.getElementById('close').value='closable'">window is closable</button>
		<button onclick="external.invoke('winUnClose');document.getElementById('close').value='isnt closable'"> window isn't closable</button>
		<input id="close" value="closable" type="text" />
		<script>
		// for resolve bug on webkit2gtk
		if(window.external === undefined){
			window.external={invoke:function(x){window.webkit.messageHandlers.external.postMessage(x);}}
		}
		</script>
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
	switch {
	case data == "close":
		w.Terminate()
	case data == "fullscreen":
		w.SetFullscreen(true)
	case data == "unfullscreen":
		w.SetFullscreen(false)
	case data == "open":
		log.Println("open", w.FileOpen("*.jpg;*.png;*.bmp"))
	case data == "opendir":
		log.Println("open", w.DirectoryOpen())
	case data == "save":
		log.Println("save", w.FileSave("*.txt;*.md"))
	case data == "message":
		w.Message("title", "test", webui.MessageMsg)
	case data == "info":
		w.Message("title", "test", webui.MessageInfo|webui.MessageButtonOKCancel)
	case data == "warning":
		w.Message("title", "test", webui.MessageWarning|webui.MessageButtonYesNo)
	case data == "error":
		w.Message("title", "test", webui.MessageError|webui.MessageButtonYesNoCancel)
	case data == "winUnClose":
		canClose = false
	case data == "winClose":
		canClose = true
	case strings.HasPrefix(data, "changeTitle:"):
		w.SetTitle(strings.TrimPrefix(data, "changeTitle:"))
	case strings.HasPrefix(data, "changeColor:"):
		hex := strings.TrimPrefix(strings.TrimPrefix(data, "changeColor:"), "#")
		num := len(hex) / 2
		if !(num == 3 || num == 4) {
			log.Println("Color must be RRGGBB or RRGGBBAA")
			return
		}
		i, err := strconv.ParseUint(hex, 16, 64)
		if err != nil {
			log.Println(err)
			return
		}
		if num == 3 {
			r := uint8((i >> 16) & 0xFF)
			g := uint8((i >> 8) & 0xFF)
			b := uint8(i & 0xFF)
			w.SetColor(r, g, b, 255)
			return
		}
		if num == 4 {
			r := uint8((i >> 24) & 0xFF)
			g := uint8((i >> 16) & 0xFF)
			b := uint8((i >> 8) & 0xFF)
			a := uint8(i & 0xFF)
			w.SetColor(r, g, b, a)
			return
		}
	}
}

var canClose = true

func wincloseCallback(webui.WebUI) bool {
	return canClose
}

func main() {
	url := startServer()
	w := webui.New(webui.Settings{
		Width:                  windowWidth,
		Height:                 windowHeight,
		Title:                  "Simple window demo",
		URL:                    url,
		ExternalInvokeCallback: handleRPC,
		CloseCallback:          wincloseCallback,
	})
	w.SetMinSize(300, 300)
	w.SetColor(255, 255, 255, 255)
	defer w.Exit()
	w.Run()
}
