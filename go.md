# webui
--
    import "github.com/srfirouzi/webui"

Package webui implements Go bindings to https://github.com/srfirouzi/webui C
library.

Bindings closely repeat the C APIs and include both, a simplified
single-function API to just open a full-screen webui window, and a more advanced
and featureful set of APIs, including Go-to-JavaScript bindings.

The library uses gtk-webkit, Cocoa/Webkit and MSHTML (IE8..11) as a browser
engine and supports Linux, Windows 7..10 respectively.

## Usage

```go
const (
	//MessageMsg default message box style
	MessageMsg MessageFlag = C.WEBUI_MSG_MSG
	//MessageInfo information message box style
	MessageInfo MessageFlag = C.WEBUI_MSG_INFO
	//MessageWarning warning message box style
	MessageWarning MessageFlag = C.WEBUI_MSG_WARNING
	//MessageError error message box style
	MessageError MessageFlag = C.WEBUI_MSG_ERROR
	//MessageButtonOK message box by ok button
	MessageButtonOK MessageFlag = C.WEBUI_MSG_OK
	//MessageButtonOKCancel message box by ok/cancel button
	MessageButtonOKCancel MessageFlag = C.WEBUI_MSG_OK_CANCEL
	//MessageButtonYesNo message box by yes/no button
	MessageButtonYesNo MessageFlag = C.WEBUI_MSG_YES_NO
	//MessageButtonYesNoCancel message box by yes/no/cansel button
	MessageButtonYesNoCancel MessageFlag = C.WEBUI_MSG_YES_NO_CANCEL

	// MessageResponseOk ok click response
	MessageResponseOk MessageResponse = C.WEBUI_RESPONSE_OK
	// MessageResponseCancel cancel click response
	MessageResponseCancel MessageResponse = C.WEBUI_RESPONSE_CANCEL
	// MessageResponseYes yes click response
	MessageResponseYes MessageResponse = C.WEBUI_RESPONSE_YES
	// MessageResponseNo no click response
	MessageResponseNo MessageResponse = C.WEBUI_RESPONSE_NO
)
```

#### func  Debug

```go
func Debug(a ...interface{})
```
Debug prints a debug string using stderr on Linux/BSD OutputDebugString on
Windows.

#### func  Debugf

```go
func Debugf(format string, a ...interface{})
```
Debugf prints a formatted debug string using stderr on Linux/BSD and
OutputDebugString on Windows.

#### func  Open

```go
func Open(title, url string, w, h int, border WindowBorder) error
```
Open is a simplified API to open a single native window with a full-size webui
in it. It can be helpful if you want to communicate with the core app using XHR
or WebSockets (as opposed to using JavaScript bindings).

Window appearance can be customized using title, width, height and resizable
parameters. URL must be provided and can user either a http or https protocol,
or be a local file:// URL. On some platforms "data:" URLs are also supported
(Linux).

#### type CloseCallbackFunc

```go
type CloseCallbackFunc func(w WebUI) bool
```

CloseCallbackFunc is function type for callback in user can close the windows

#### type ExternalInvokeCallbackFunc

```go
type ExternalInvokeCallbackFunc func(w WebUI, data string)
```

ExternalInvokeCallbackFunc is a function type that is called every time
"window.external.invoke()" is called from JavaScript. Data is the only
obligatory string parameter passed into the "invoke(data)" function from
JavaScript. To pass more complex data serialized JSON or base64 encoded string
can be used.

#### type MessageFlag

```go
type MessageFlag int
```

MessageFlag flag for msg function

#### type MessageResponse

```go
type MessageResponse int
```

MessageResponse respond button

#### type Settings

```go
type Settings struct {
	// WebUI main window title
	Title string
	// URL to open in a webui
	URL string
	// Window width in pixels
	Width int
	// Window height in pixels
	Height int
	// Allows/disallows window resizing
	Border WindowBorder
	// Enable debugging tools (Linux/BSD, on Windows use Firebug)
	Debug bool
	// A callback that is executed when JavaScript calls "window.external.invoke()"
	ExternalInvokeCallback ExternalInvokeCallbackFunc
	// A callback for windows close event
	CloseCallback CloseCallbackFunc
}
```

Settings is a set of parameters to customize the initial WebUI appearance and
behavior. It is passed into the webui.New() constructor.

#### type WebUI

```go
type WebUI interface {
	// Run() starts the main UI loop until the user closes the webui window or
	// Terminate() is called.
	Run()
	// Loop() runs a single iteration of the main UI.
	Loop(blocking bool) bool
	// SetTitle() changes window title. This method must be called from the main
	// thread only. See Dispatch() for more details.
	SetTitle(title string)
	// SetFullscreen() controls window full-screen mode. This method must be
	// called from the main thread only. See Dispatch() for more details.
	SetFullscreen(fullscreen bool)
	// SetColor() changes window background color. This method must be called from
	// the main thread only. See Dispatch() for more details.
	SetColor(r, g, b, a uint8)
	// SetMinSize() set min size for window
	// called from the main thread only
	SetMinSize(width int, height int)
	// Eval() evaluates an arbitrary JS code inside the webui. This method must
	// be called from the main thread only. See Dispatch() for more details.
	Eval(js string) error
	// InjectJS() injects an arbitrary block of CSS code using the JS API. This
	// method must be called from the main thread only. See Dispatch() for more
	// details.
	InjectCSS(css string)
	// Message() open message box and return button click by user
	Message(title string, msg string, flags MessageFlag) MessageResponse
	// FileOpen() open file open dialog and response selected file
	FileOpen(filter string) string
	// FileSave() open file save dialog and response selected file
	FileSave(filter string) string
	// DirectoryOpen() open directory open dialog and response selected directory
	DirectoryOpen() string
	// Terminate() breaks the main UI loop. This method must be called from the main thread
	// only. See Dispatch() for more details.
	Terminate()
	// Dispatch() schedules some arbitrary function to be executed on the main UI
	// thread. This may be helpful if you want to run some JavaScript from
	// background threads/goroutines, or to terminate the app.
	Dispatch(func())
	// Exit() closes the window and cleans up the resources. Use Terminate() to
	// forcefully break out of the main UI loop.
	Exit()
	// Bind() registers a binding between a given value and a JavaScript object with the
	// given name.  A value must be a struct or a struct pointer. All methods are
	// available under their camel-case names, starting with a lower-case letter,
	// e.g. "FooBar" becomes "fooBar" in JavaScript.
	// Bind() returns a function that updates JavaScript object with the current
	// Go value. You only need to call it if you change Go value asynchronously.
	Bind(name string, v interface{}) (sync func(), err error)
}
```

WebUI is an interface that wraps the basic methods for controlling the UI loop,
handling multithreading and providing JavaScript bindings.

#### func  New

```go
func New(settings Settings) WebUI
```
New creates and opens a new webui window using the given settings. The returned
object implements the WebUI interface. This function returns nil if a window can
not be created.

#### type WindowBorder

```go
type WindowBorder int
```

WindowBorder is border for window

```go
const (
	// BorderNone is none border for window
	BorderNone WindowBorder = C.WEBUI_BORDER_NONE
	// BorderDialog is dialog border for window
	BorderDialog WindowBorder = C.WEBUI_BORDER_DIALOG
	// BorderResizable is sizable border for window
	BorderResizable WindowBorder = C.WEBUI_BORDER_RESIZABLE
)
```
