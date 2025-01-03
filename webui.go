//
// Package webui implements Go bindings to https://github.com/srfirouzi/webui C library.
//
// Bindings closely repeat the C APIs and include both, a simplified
// single-function API to just open a full-screen webui window, and a more
// advanced and featureful set of APIs, including Go-to-JavaScript bindings.
//
// The library uses gtk-webkit, Cocoa/Webkit and MSHTML (IE8..11) as a browser
// engine and supports Linux, Windows 7..10 respectively.
//
package webui

/*
#cgo linux openbsd freebsd CFLAGS: -DWEBUI_GTK=1
#cgo linux openbsd freebsd pkg-config: gtk+-3.0 webkit2gtk-4.1

#cgo windows CFLAGS: -DWEBUI_WIN=1
#cgo windows LDFLAGS: -lole32 -lcomctl32 -loleaut32 -luuid -lgdi32





#include <stdlib.h>
#include <stdint.h>
#define WEBUI_STATIC

#ifdef WEBUI_GTK
	#include "lib/gtk.h"
#endif

#ifdef WEBUI_WIN
	#include "lib/win.h"
#endif

extern void _WebUiExternalInvokeCallback(void *, void *);

extern int _WebUiCloseCallback(void *);

static inline void CgoWebUiFree(void *w) {
	free((void *)((struct webui *)w)->title);
	free((void *)((struct webui *)w)->url);
	free(w);
}

static inline void *CgoWebUiCreate(int width, int height, char *title, char *url, int border, int debug) {
	struct webui *w = (struct webui *) calloc(1, sizeof(*w));
	w->width = width;
	w->height = height;
	w->title = title;
	w->url = url;
	w->border = border;
	w->debug = debug;
	w->external_invoke_cb = (webui_external_invoke_cb_t) _WebUiExternalInvokeCallback;
	w->close_cb =(webui_close_cb) _WebUiCloseCallback;
	if (webui_init(w) != 0) {
		CgoWebUiFree(w);
		return NULL;
	}
	return (void *)w;
}

static inline int CgoWebUiLoop(void *w, int blocking) {
	return webui_loop((struct webui *)w, blocking);
}

static inline void CgoWebUiTerminate(void *w) {
	webui_terminate((struct webui *)w);
}

static inline void CgoWebUiExit(void *w) {
	webui_exit((struct webui *)w);
}

static inline void CgoWebUiSetTitle(void *w, char *title) {
	webui_set_title((struct webui *)w, title);
}

static inline void CgoWebUiSetFullscreen(void *w, int fullscreen) {
	webui_set_fullscreen((struct webui *)w, fullscreen);
}

static inline void CgoWebUiSetColor(void *w, uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
	webui_set_color((struct webui *)w, r, g, b, a);
}

static inline void CgoWebUiSetMinSize(void *w,  int width,int height) {
	webui_set_min_size((struct webui *)w, width, height);
}
static inline int CgoWebUiMsg(void *w, int flags,char *title, char *msg){
	return webui_msg((struct webui *)w,flags, title,msg);
}

static inline void CgoWebUiFile(void *w, int type,char *filter, char *res, size_t ressz) {
	webui_file(w, type,(const char*) filter, res, ressz);
}

static inline int CgoWebUiEval(void *w, char *js) {
	return webui_eval((struct webui *)w, js);
}

static inline void CgoWebUiInjectCSS(void *w, char *css) {
	webui_inject_css((struct webui *)w, css);
}

extern void _WebUiDispatchGoCallback(void *);
static inline void _webui_dispatch_cb(struct webui *w, void *arg) {
	_WebUiDispatchGoCallback(arg);
}
static inline void CgoWebUiDispatch(void *w, uintptr_t arg) {
	webui_dispatch((struct webui *)w, _webui_dispatch_cb, (void *)arg);
}
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"runtime"
	"sync"
	"unicode"
	"unsafe"
)

// WindowBorder is border for window
type WindowBorder int

const (
	// BorderNone is none border for window
	BorderNone WindowBorder = C.WEBUI_BORDER_NONE
	// BorderDialog is dialog border for window
	BorderDialog WindowBorder = C.WEBUI_BORDER_DIALOG
	// BorderResizable is sizable border for window
	BorderResizable WindowBorder = C.WEBUI_BORDER_RESIZABLE
)

// MessageFlag flag for msg function
type MessageFlag int

//MessageResponse respond button
type MessageResponse int

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

func init() {
	// Ensure that main.main is called from the main thread
	runtime.LockOSThread()
}

// Open is a simplified API to open a single native window with a full-size webui in
// it. It can be helpful if you want to communicate with the core app using XHR
// or WebSockets (as opposed to using JavaScript bindings).
//
// Window appearance can be customized using title, width, height and resizable parameters.
// URL must be provided and can user either a http or https protocol, or be a
// local file:// URL. On some platforms "data:" URLs are also supported
// (Linux).
func Open(title, url string, w, h int, border WindowBorder) error {
	titleStr := C.CString(title)
	defer C.free(unsafe.Pointer(titleStr))
	urlStr := C.CString(url)
	defer C.free(unsafe.Pointer(urlStr))
	bord := C.int(border)
	r := C.webui(titleStr, urlStr, C.int(w), C.int(h), bord)
	if r != 0 {
		return errors.New("failed to create webui")
	}
	return nil
}

// Debug prints a debug string using stderr on Linux/BSD
// OutputDebugString on Windows.
func Debug(a ...interface{}) {
	s := C.CString(fmt.Sprint(a...))
	defer C.free(unsafe.Pointer(s))
	C.webui_print_log(s)
}

// Debugf prints a formatted debug string using stderr on Linux/BSD and OutputDebugString on Windows.
func Debugf(format string, a ...interface{}) {
	s := C.CString(fmt.Sprintf(format, a...))
	defer C.free(unsafe.Pointer(s))
	C.webui_print_log(s)
}

// ExternalInvokeCallbackFunc is a function type that is called every time
// "window.external.invoke()" is called from JavaScript. Data is the only
// obligatory string parameter passed into the "invoke(data)" function from
// JavaScript. To pass more complex data serialized JSON or base64 encoded
// string can be used.
type ExternalInvokeCallbackFunc func(w WebUI, data string)

//CloseCallbackFunc is function type for callback in user can close the windows
type CloseCallbackFunc func(w WebUI) bool

// Settings is a set of parameters to customize the initial WebUI appearance
// and behavior. It is passed into the webui.New() constructor.
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

// WebUI is an interface that wraps the basic methods for controlling the UI
// loop, handling multithreading and providing JavaScript bindings.
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

var (
	m     sync.Mutex
	index uintptr
	fns   = map[uintptr]func(){}
	cbei  = map[WebUI]ExternalInvokeCallbackFunc{}
	cbc   = map[WebUI]CloseCallbackFunc{}
)

type webui struct {
	w unsafe.Pointer
}

var _ WebUI = &webui{}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// New creates and opens a new webui window using the given settings. The
// returned object implements the WebUI interface. This function returns nil
// if a window can not be created.
func New(settings Settings) WebUI {
	if settings.Width == 0 {
		settings.Width = 640
	}
	if settings.Height == 0 {
		settings.Height = 480
	}
	if settings.Title == "" {
		settings.Title = "WebUI"
	}
	w := &webui{}
	w.w = C.CgoWebUiCreate(C.int(settings.Width), C.int(settings.Height),
		C.CString(settings.Title), C.CString(settings.URL),
		C.int(settings.Border), C.int(boolToInt(settings.Debug)))
	m.Lock()
	if settings.ExternalInvokeCallback != nil {
		cbei[w] = settings.ExternalInvokeCallback
	} else {
		cbei[w] = func(w WebUI, data string) {}
	}
	if settings.CloseCallback != nil {
		cbc[w] = settings.CloseCallback
	} else {
		cbc[w] = func(w WebUI) bool { return true }
	}
	m.Unlock()
	return w
}

func (w *webui) Loop(blocking bool) bool {
	block := C.int(0)
	if blocking {
		block = 1
	}
	return C.CgoWebUiLoop(w.w, block) == 0
}

func (w *webui) Run() {
	for w.Loop(true) {
	}
}

func (w *webui) Exit() {
	C.CgoWebUiExit(w.w)
}

func (w *webui) Dispatch(f func()) {
	m.Lock()
	for ; fns[index] != nil; index++ {
	}
	fns[index] = f
	m.Unlock()
	C.CgoWebUiDispatch(w.w, C.uintptr_t(index))
}

func (w *webui) SetTitle(title string) {
	p := C.CString(title)
	defer C.free(unsafe.Pointer(p))
	C.CgoWebUiSetTitle(w.w, p)
}

func (w *webui) SetColor(r, g, b, a uint8) {
	C.CgoWebUiSetColor(w.w, C.uint8_t(r), C.uint8_t(g), C.uint8_t(b), C.uint8_t(a))
}

func (w *webui) SetMinSize(width int, height int) {
	C.CgoWebUiSetMinSize(w.w, C.int(width), C.int(height))
}

func (w *webui) SetFullscreen(fullscreen bool) {
	C.CgoWebUiSetFullscreen(w.w, C.int(boolToInt(fullscreen)))
}

func (w *webui) Message(title string, msg string, flags MessageFlag) MessageResponse {
	titlePtr := C.CString(title)
	defer C.free(unsafe.Pointer(titlePtr))
	msgPtr := C.CString(msg)
	defer C.free(unsafe.Pointer(msgPtr))
	res := C.CgoWebUiMsg(w.w, C.int(flags), titlePtr, msgPtr)
	return MessageResponse(res)
}

func (w *webui) FileOpen(filter string) string {
	return w.file(0, filter)
}

func (w *webui) FileSave(filter string) string {
	return w.file(1, filter)
}

func (w *webui) DirectoryOpen() string {
	return w.file(2, "")
}

func (w *webui) file(Type int, filter string) string {
	const maxPath = 4096
	filterPtr := C.CString(filter)
	defer C.free(unsafe.Pointer(filterPtr))
	resultPtr := (*C.char)(C.calloc((C.size_t)(unsafe.Sizeof((*C.char)(nil))), (C.size_t)(maxPath)))
	defer C.free(unsafe.Pointer(resultPtr))
	C.CgoWebUiFile(w.w, C.int(Type), filterPtr, resultPtr, C.size_t(maxPath))
	return C.GoString(resultPtr)

}

func (w *webui) Eval(js string) error {
	p := C.CString(js)
	defer C.free(unsafe.Pointer(p))
	switch C.CgoWebUiEval(w.w, p) {
	case -1:
		return errors.New("evaluation failed")
	}
	return nil
}

func (w *webui) InjectCSS(css string) {
	p := C.CString(css)
	defer C.free(unsafe.Pointer(p))
	C.CgoWebUiInjectCSS(w.w, p)
}

func (w *webui) Terminate() {
	C.CgoWebUiTerminate(w.w)
}

//export _WebUiDispatchGoCallback
func _WebUiDispatchGoCallback(index unsafe.Pointer) {
	var f func()
	m.Lock()
	f = fns[uintptr(index)]
	delete(fns, uintptr(index))
	m.Unlock()
	f()
}

//export _WebUiCloseCallback
func _WebUiCloseCallback(w unsafe.Pointer) C.int {
	m.Lock()
	var (
		cb CloseCallbackFunc
		wv WebUI
	)
	for wv, cb = range cbc {
		if wv.(*webui).w == w {
			break
		}
	}
	m.Unlock()
	if cb(wv) {
		return C.int(1)
	}
	return C.int(0)
}

//export _WebUiExternalInvokeCallback
func _WebUiExternalInvokeCallback(w unsafe.Pointer, data unsafe.Pointer) {
	m.Lock()
	var (
		cb ExternalInvokeCallbackFunc
		wv WebUI
	)
	for wv, cb = range cbei {
		if wv.(*webui).w == w {
			break
		}
	}
	m.Unlock()
	cb(wv, C.GoString((*C.char)(data)))
}

var bindTmpl = template.Must(template.New("").Parse(`
if (typeof {{.Name}} === 'undefined') {
	{{.Name}} = {};
}
{{ range .Methods }}
{{$.Name}}.{{.JSName}} = function({{.JSArgs}}) {
	window.external.invoke(JSON.stringify({scope: "{{$.Name}}", method: "{{.Name}}", params: [{{.JSArgs}}]}));
};
{{ end }}
`))

type binding struct {
	Value   interface{}
	Name    string
	Methods []methodInfo
}

func newBinding(name string, v interface{}) (*binding, error) {
	methods, err := getMethods(v)
	if err != nil {
		return nil, err
	}
	return &binding{Name: name, Value: v, Methods: methods}, nil
}

func (b *binding) JS() (string, error) {
	js := &bytes.Buffer{}
	err := bindTmpl.Execute(js, b)
	return js.String(), err
}

func (b *binding) Sync() (string, error) {
	js, err := json.Marshal(b.Value)
	if err == nil {
		return fmt.Sprintf("%[1]s.data=%[2]s;if(%[1]s.render){%[1]s.render(%[2]s);}", b.Name, string(js)), nil
	}
	return "", err
}

func (b *binding) Call(js string) bool {
	type rpcCall struct {
		Scope  string        `json:"scope"`
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}

	rpc := rpcCall{}
	if err := json.Unmarshal([]byte(js), &rpc); err != nil {
		return false
	}
	if rpc.Scope != b.Name {
		return false
	}
	var mi *methodInfo
	for i := 0; i < len(b.Methods); i++ {
		if b.Methods[i].Name == rpc.Method {
			mi = &b.Methods[i]
			break
		}
	}
	if mi == nil {
		return false
	}
	args := make([]reflect.Value, mi.Arity(), mi.Arity())
	for i := 0; i < mi.Arity(); i++ {
		val := reflect.ValueOf(rpc.Params[i])
		arg := mi.Value.Type().In(i)
		u := reflect.New(arg)
		if b, err := json.Marshal(val.Interface()); err == nil {
			if err = json.Unmarshal(b, u.Interface()); err == nil {
				args[i] = reflect.Indirect(u)
			}
		}
		if !args[i].IsValid() {
			return false
		}
	}
	mi.Value.Call(args)
	return true
}

type methodInfo struct {
	Name  string
	Value reflect.Value
}

func (mi methodInfo) Arity() int { return mi.Value.Type().NumIn() }

func (mi methodInfo) JSName() string {
	r := []rune(mi.Name)
	if len(r) > 0 {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}

func (mi methodInfo) JSArgs() (js string) {
	for i := 0; i < mi.Arity(); i++ {
		if i > 0 {
			js = js + ","
		}
		js = js + fmt.Sprintf("a%d", i)
	}
	return js
}

func getMethods(obj interface{}) ([]methodInfo, error) {
	p := reflect.ValueOf(obj)
	v := reflect.Indirect(p)
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil, errors.New("object can not be nil")
	}
	k := t.Kind()
	if k == reflect.Ptr {
		k = v.Type().Kind()
	}
	if k != reflect.Struct {
		return nil, errors.New("must be a struct or a pointer to a struct")
	}

	methods := []methodInfo{}
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if !unicode.IsUpper([]rune(method.Name)[0]) {
			continue
		}
		mi := methodInfo{
			Name:  method.Name,
			Value: p.MethodByName(method.Name),
		}
		methods = append(methods, mi)
	}

	return methods, nil
}

func (w *webui) Bind(name string, v interface{}) (sync func(), err error) {
	b, err := newBinding(name, v)
	if err != nil {
		return nil, err
	}
	js, err := b.JS()
	if err != nil {
		return nil, err
	}
	sync = func() {
		if js, err := b.Sync(); err != nil {
			log.Println(err)
		} else {
			w.Eval(js)
		}
	}

	m.Lock()
	cb := cbei[w]
	cbei[w] = func(w WebUI, data string) {
		if ok := b.Call(data); ok {
			sync()
		} else {
			cb(w, data)
		}
	}
	m.Unlock()

	w.Eval(js)
	sync()
	return sync, nil
}
