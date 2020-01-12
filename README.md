# webui



A tiny cross-platform webui library for C/Golang to build modern cross-platform GUIs
this fork from [webview](https://github.com/zserge/webview/)

## new feature (todo/doing/done)

- [-] different window border
  - resolve border-none bug in wine on linux
- [x] set window icon if exist in resource by ID 100(windows only)
- [ ] callback for window events
  - [ ] close
- [ ] add min size for window
- [ ] new design for dialog
  - add filter to open/save file
  - add new method for open dialog
  - multi button dialog box and return result
  - add color selector dialog (ie11 don't support color input on html)

 


It supports two-way JavaScript bindings (to call JavaScript from C/Go and to call C/Go from JavaScript).

It uses gtk-webkit2 on Linux and MSHTML (IE10/11) on Windows.



## WebUI for Go developers

If you are interested in writing WebUI apps in C, [skip to the next section](#webview-for-c-developers).

### Getting started

Install WebUI library with `go get`:

```
$ go get github.com/srfirouzi/webui
```

Import the package and start using it:

```go
package main

import "github.com/srfirouzi/webui"

func main() {
	// Open wikipedia in a 800x600 resizable window
	webui.Open("Minimal webui example",
		"https://en.m.wikipedia.org/wiki/Main_Page", 800, 600, webui.WEBUI_BORDER_SIZABLE)
}
```

It is not recommended to use `go run` (although it works perfectly fine on Linux). Use `go build` instead:

```bash
# Linux
$ go build -o webui-example && ./webui-example

# Windows requires special linker flags for GUI apps.
# It's also recommended to use TDM-GCC-64 compiler for CGo.
# http://tdm-gcc.tdragon.net/download
$ go build -ldflags="-H windowsgui" -o webui-example.exe
```



### How to serve or inject the initial HTML/CSS/JavaScript into the webui?

First of all, you probably want to embed your assets (HTML/CSS/JavaScript) into the binary to have a standalone executable. Consider using [go-bindata](https://github.com/go-bindata/go-bindata) or any other similar tools.

Now there are two major approaches to deploy the content:

* Serve HTML/CSS/JS with an embedded HTTP server
* Injecting HTML/CSS/JS via the JavaScript binding API

To serve the content it is recommended to use ephemeral ports:

```go
ln, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
	log.Fatal(err)
}
defer ln.Close()
go func() {
 	// Set up your http server here
	log.Fatal(http.Serve(ln, nil))
}()
webui.Open("Hello", "http://"+ln.Addr().String(), 400, 300, webui.WEBUI_BORDER_DIALOG)
```

Injecting the content via JS bindings is a bit more complicated, but feels more solid and does not expose any additional open TCP ports.

Leave `webui.Settings.URL` empty to start with bare minimal HTML5. It will open a webui with `<div id="app"></div>` in it. Alternatively, use a data URI to inject custom HTML code (don't forget to URL-encode it):

```go
const myHTML = `<!doctype html><html>....</html>`
w := webui.New(webui.Settings{
  URL: `data:text/html,` + url.PathEscape(myHTML),
  Border:WEBUI_BORDER_SIZABLE,
})
```

Keep your initial HTML short (a few kilobytes maximum).

Now you can inject more JavaScript once the webui becomes ready using `webui.Eval()`. You can also inject CSS styles using JavaScript:

```go
w.Dispatch(func() {
	// Inject CSS
	w.Eval(fmt.Sprintf(`(function(css){
		var style = document.createElement('style');
		var head = document.head || document.getElementsByTagName('head')[0];
		style.setAttribute('type', 'text/css');
		if (style.styleSheet) {
			style.styleSheet.cssText = css;
		} else {
			style.appendChild(document.createTextNode(css));
		}
		head.appendChild(style);
	})("%s")`, template.JSEscapeString(myStylesCSS)))
	// Inject JS
	w.Eval(myJSFramework)
	w.Eval(myAppJS)
})
```

This works fairly well across the platforms, see `counter-go` example for more details about how make a webui app with no web server. It also demonstrates how to use ReactJS, VueJS or Picodom with webui.

### How to communicate between native Go and web UI?

You already have seen how to use `w.Eval()` to run JavaScript inside the webui. There is also a way to call Go code from JavaScript.

On the low level there is a special callback, `webui.Settings.ExternalInvokeCallback` that receives a string argument. This string can be passed from JavaScript using `window.external.invoke(someString)`.

This might seem very inconvenient, and that is why there is a dedicated `webui.Bind()` API call. It binds an existing Go object (struct or struct pointer) and creates/injects JS API for it. Now you can call JS methods and they will result in calling native Go methods. Even more, if you modify the Go object - it can be automatically serialized to JSON and passed to the web UI to keep things in sync.

Please, see `counter-go` example for more details about how to bind Go controllers to the web UI.

## Debugging and development tips

If terminal output is unavailable (e.g. if you launch app bundle on MacOS or
GUI app on Windows) you may use `webui.Debug()` and `webui.Debugf()` to
print logs. On MacOS such logs will be printed via NSLog and can be seen in the
`Console` app. On Windows they use `OutputDebugString` and can be seen using
`DebugView` app. On Linux logging is done to stderr and can be seen in the
terminal or redirected to a file.

To debug the web part of your app you may use `webui.Settings.Debug` flag. It
enables the Web Inspector in WebKit and works on Linux and MacOS (use popup menu
to open the web inspector). On Windows there is no easy to way to enable
debugging, but you may include Firebug in your HTML code:

```html
<script type="text/javascript" src="https://getfirebug.com/firebug-lite.js"></script>
```

Even though Firebug browser extension development has been stopped, Firebug
Lite is still available and just works.

## Distributing webui apps

On Linux you get a standalone executable. It will depend on GTK3 and GtkWebkit2, so if you distribute your app in DEB or RPM format include those dependencies. An application icon can be specified by providing a `.desktop` file.

On Windows you probably would like to have a custom icon for your executable. It can be done by providing a resource file, compiling it and linking with it,icon by id 100 in resource if exist used for window icon, by [rsrc](https://github.com/srfirouzi/rsrc) can make this elements

## WebUI for C developers

### Getting started

Download [lib/gtk.h](https://raw.githubusercontent.com/srfirouzi/webui/master/lib/gtk.h) for linux or Download [lib/win.h](https://raw.githubusercontent.com/srfirouzi/webui/master/lib/win.h) for window and include it in your C code:

```c
// main.c
#ifdef WIN32
  #include "win.h" 
#else
   #include "gtk.h"
#endif

#ifdef WIN32
int WINAPI WinMain(HINSTANCE hInt, HINSTANCE hPrevInst, LPSTR lpCmdLine,
                   int nCmdShow) {
#else
int main() {
#endif
  /* Open wikipedia in a 800x600 resizable window */
  /*
  border can set this value
  WEBUI_BORDER_NONE=2,
  WEBUI_BORDER_DIALOG=1,
  WEBUI_BORDER_SIZABLE=0
  */
  webui("Minimal webui example",
	  "https://en.m.wikipedia.org/wiki/Main_Page", 800, 600, WEBUI_BORDER_SIZABLE);
  return 0;
}
```

Build it:

```bash
# Linux
$ cc main.c `pkg-config --cflags --libs gtk+-3.0 webkit2gtk-4.0` -o webui-example

# Windows (mingw)
$ cc main.c -lole32 -lcomctl32 -loleaut32 -luuid -mwindows -o webui-example.exe
```

### API

For the most simple use cases there is only one function:

```c
int webui(const char *title, const char *url, int width, int height, int border);
```

The following URL schemes are supported:

* `http://` and `https://`, no surprises here.
* `file:///` can be useful if you want to unpack HTML/CSS assets to some
  temporary directory and point a webui to open index.html from there.
* `data:text/html,<html>...</html>` allows to pass short HTML data inline
  without using a web server or polluting the file system. Further
  modifications of the webui contents can be done via JavaScript bindings.

If have chosen a regular http URL scheme, you can use Mongoose or any other web server/framework you like.

If you want to have more control over the app lifecycle you can use the following functions:

```c
  struct webui webui = {
      .title = title,
      .url = url,
      .width = w,
      .height = h,
      .debug = debug,
      .border = border,
  };
  /* Create webui window using the provided options */
  webui_init(&webui);
  /* Main app loop, can be either blocking or non-blocking */
  while (webui_loop(&webui, blocking) == 0);
  /* Destroy webui window, often exits the app */
  webui_exit(&webui);

  /* To change window title later: */
  webui_set_title(&webui, "New title");

  /* To terminate the webui main loop: */
  webui_terminate(&webui);

  /* To print logs to stderr, MacOS Console or DebugView: */
  webui_debug("exited: %d\n", 1);
```

To evaluate arbitrary JavaScript code use the following C function:

```c
webui_eval(&webui, "alert('hello, world');");
```

There is also a special callback (`webui.external_invoke_cb`) that can be invoked from JavaScript:

```javascript
// C
void my_cb(struct webui *w, const char *arg) {
	...
}

// JS
window.external.invoke('some arg');
// Exactly one string argument must be provided, to pass more complex objects
// serialize them to JSON and parse it in C. To pass binary data consider using
// base64.
window.external.invoke(JSON.stringify({fn: 'sum', x: 5, y: 3}));
```

webui library is meant to be used from a single UI thread only. So if you
want to call `webui_eval` or `webui_terminate` from some background thread
- you have to use `webui_dispatch` to post some arbitrary function with some
context to be executed inside the main UI thread:

```c
// This function will be executed on the UI thread
void render(struct webui *w, void *arg) {
  webui_eval(w, ......);
}

// Dispatch render() function from another thread:
webui_dispatch(w, render, some_arg);
```

You may find some C examples in this repo that demonstrate the API above.

## cross compile

cross compile need active cgo and installed cross compile tools for c language.

### cross compile on linux for windows

first install mingw
```bash
#ubuntu
sudo apt install binutils-mingw-w64
```
then set envierment elements to use by go compiler 

```bash
#64bit
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags "-H windowsgui"

#32bit
GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go build -ldflags "-H windowsgui"
```

## License

Code is distributed under MIT license, feel free to use it in your proprietary
projects as well.
