#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include <JavaScriptCore/JavaScript.h>
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>

#ifdef WEBUI_STATIC
#define WEBUI_API static
#else
#define WEBUI_API extern
#endif

struct webui_priv {
  GtkWidget *window;
  GtkWidget *scroller;
  GtkWidget *webui;
  GtkWidget *inspector_window;
  GAsyncQueue *queue;
  int ready;
  int js_busy;
  int should_exit;
};

struct webui;

typedef void (*webui_external_invoke_cb_t)(struct webui *w,
                                             const char *arg);
typedef int (*webui_close_cb)(struct webui *w);

enum webui_border_type{
  WEBUI_BORDER_NONE=2,
  WEBUI_BORDER_DIALOG=1,
  WEBUI_BORDER_SIZABLE=0
};

struct webui {
  const char *url;
  const char *title;
  int width;
  int height;
  int minWidth;
  int minHeight;
  int border;
  int debug;
  webui_external_invoke_cb_t external_invoke_cb;
  webui_close_cb close_cb;
  struct webui_priv priv;
  void *userdata;
};

enum webui_dialog_type {
  WEBUI_DIALOG_TYPE_OPEN = 0,
  WEBUI_DIALOG_TYPE_SAVE = 1,
  WEBUI_DIALOG_TYPE_ALERT = 2
};
#define WEBUI_MSG_ICON_MASK (3 << 0)
#define WEBUI_MSG_BUTTON_MASK (3 << 2)
enum webui_msg_type{
  /* 2 bit for msg type*/
  WEBUI_MSG_MSG=0,
  WEBUI_MSG_INFO=1,
  WEBUI_MSG_WARNING=2,
  WEBUI_MSG_ERROR=3,
  /*other bit for button*/
  WEBUI_MSG_OK=0,
  WEBUI_MSG_OK_CANCEL=4,
  WEBUI_MSG_YES_NO=8,
  WEBUI_MSG_YES_NO_CANCEL=12
  /* other buttons model*/
};
enum webui_response_type{
  WEBUI_RESPONSE_OK=0,
  WEBUI_RESPONSE_CANCEL=1,
  WEBUI_RESPONSE_YES=2,
  WEBUI_RESPONSE_NO=3
};

#define WEBUI_DIALOG_FLAG_FILE (0 << 0)
#define WEBUI_DIALOG_FLAG_DIRECTORY (1 << 0)

#define WEBUI_DIALOG_FLAG_INFO (1 << 1)
#define WEBUI_DIALOG_FLAG_WARNING (2 << 1)
#define WEBUI_DIALOG_FLAG_ERROR (3 << 1)
#define WEBUI_DIALOG_FLAG_ALERT_MASK (3 << 1)

typedef void (*webui_dispatch_fn)(struct webui *w, void *arg);

struct webui_dispatch_arg {
  webui_dispatch_fn fn;
  struct webui *w;
  void *arg;
};

#define DEFAULT_URL                                                            \
  "data:text/"                                                                 \
  "html,%3C%21DOCTYPE%20html%3E%0A%3Chtml%20lang=%22en%22%3E%0A%3Chead%3E%"    \
  "3Cmeta%20charset=%22utf-8%22%3E%3Cmeta%20http-equiv=%22X-UA-Compatible%22%" \
  "20content=%22IE=edge%22%3E%3C%2Fhead%3E%0A%3Cbody%3E%3Cdiv%20id=%22app%22%" \
  "3E%3C%2Fdiv%3E%3Cscript%20type=%22text%2Fjavascript%22%3E%3C%2Fscript%3E%"  \
  "3C%2Fbody%3E%0A%3C%2Fhtml%3E"

#define CSS_INJECT_FUNCTION                                                    \
  "(function(e){var "                                                          \
  "t=document.createElement('style'),d=document.head||document."               \
  "getElementsByTagName('head')[0];t.setAttribute('type','text/"               \
  "css'),t.styleSheet?t.styleSheet.cssText=e:t.appendChild(document."          \
  "createTextNode(e)),d.appendChild(t)})"

static const char *webui_check_url(const char *url) {
  if (url == NULL || strlen(url) == 0) {
    return DEFAULT_URL;
  }
  return url;
}

WEBUI_API int webui(const char *title, const char *url, int width,
                        int height, int border);

WEBUI_API int webui_init(struct webui *w);
WEBUI_API int webui_loop(struct webui *w, int blocking);
WEBUI_API int webui_eval(struct webui *w, const char *js);
WEBUI_API int webui_inject_css(struct webui *w, const char *css);
WEBUI_API void webui_set_title(struct webui *w, const char *title);
WEBUI_API void webui_set_fullscreen(struct webui *w, int fullscreen);
WEBUI_API void webui_set_color(struct webui *w, uint8_t r, uint8_t g,
                                   uint8_t b, uint8_t a);
WEBUI_API void webui_dialog(struct webui *w,
                                enum webui_dialog_type dlgtype, int flags,
                                const char *title, const char *arg,
                                char *result, size_t resultsz);
WEBUI_API void webui_dispatch(struct webui *w, webui_dispatch_fn fn,
                                  void *arg);
WEBUI_API void webui_terminate(struct webui *w);
WEBUI_API void webui_exit(struct webui *w);
WEBUI_API void webui_debug(const char *format, ...);
WEBUI_API void webui_print_log(const char *s);
WEBUI_API void webui_set_min_size(struct webui *w,int width,int height);
WEBUI_API int webui_msg(struct webui *w,enum webui_msg_type flag,const char *title,const char *msg);


WEBUI_API int webui(const char *title, const char *url, int width,
                        int height, int border) {
  struct webui webui;
  memset(&webui, 0, sizeof(webui));
  webui.title = title;
  webui.url = url;
  webui.width = width;
  webui.height = height;
  webui.border = border;
  int r = webui_init(&webui);
  if (r != 0) {
    return r;
  }
  while (webui_loop(&webui, 1) == 0) {
  }
  webui_exit(&webui);
  return 0;
}

WEBUI_API void webui_debug(const char *format, ...) {
  char buf[4096];
  va_list ap;
  va_start(ap, format);
  vsnprintf(buf, sizeof(buf), format, ap);
  webui_print_log(buf);
  va_end(ap);
}

static int webui_js_encode(const char *s, char *esc, size_t n) {
  int r = 1; /* At least one byte for trailing zero */
  for (; *s; s++) {
    const unsigned char c = *s;
    if (c >= 0x20 && c < 0x80 && strchr("<>\\'\"", c) == NULL) {
      if (n > 0) {
        *esc++ = c;
        n--;
      }
      r++;
    } else {
      if (n > 0) {
        snprintf(esc, n, "\\x%02x", (int)c);
        esc += 4;
        n -= 4;
      }
      r += 4;
    }
  }
  return r;
}

WEBUI_API int webui_inject_css(struct webui *w, const char *css) {
  int n = webui_js_encode(css, NULL, 0);
  char *esc = (char *)calloc(1, sizeof(CSS_INJECT_FUNCTION) + n + 4);
  if (esc == NULL) {
    return -1;
  }
  char *js = (char *)calloc(1, n);
  webui_js_encode(css, js, n);
  snprintf(esc, sizeof(CSS_INJECT_FUNCTION) + n + 4, "%s(\"%s\")",
           CSS_INJECT_FUNCTION, js);
  int r = webui_eval(w, esc);
  free(js);
  free(esc);
  return r;
}

static void external_message_received_cb(WebKitUserContentManager *m,
                                         WebKitJavascriptResult *r,
                                         gpointer arg) {
  (void)m;
  struct webui *w = (struct webui *)arg;
  if (w->external_invoke_cb == NULL) {
    return;
  }
  JSGlobalContextRef context = webkit_javascript_result_get_global_context(r);
  JSValueRef value = webkit_javascript_result_get_value(r);
  JSStringRef js = JSValueToStringCopy(context, value, NULL);
  size_t n = JSStringGetMaximumUTF8CStringSize(js);
  char *s = g_new(char, n);
  JSStringGetUTF8CString(js, s, n);
  w->external_invoke_cb(w, s);
  JSStringRelease(js);
  g_free(s);
}

static void webui_load_changed_cb(WebKitWebView *webui,
                                    WebKitLoadEvent event, gpointer arg) {
  (void)webui;
  struct webui *w = (struct webui *)arg;
  if (event == WEBKIT_LOAD_FINISHED) {
    w->priv.ready = 1;
  }
}

static void webui_destroy_cb(GtkWidget *widget, gpointer arg) {
  (void)widget;
  struct webui *w = (struct webui *)arg;
  webui_terminate(w);
}

static gboolean webui_delete_event_cb(GtkWidget *widget,GdkEvent  *event, gpointer arg) {
  (void)widget;
  struct webui *w = (struct webui *)arg;
  if(w->close_cb!=NULL){
    int result=w->close_cb(w);
    if(result==0){
      return true;
    }
  }
  return false;
}

static gboolean webui_context_menu_cb(WebKitWebView *webui,
                                        GtkWidget *default_menu,
                                        WebKitHitTestResult *hit_test_result,
                                        gboolean triggered_with_keyboard,
                                        gpointer userdata) {
  (void)webui;
  (void)default_menu;
  (void)hit_test_result;
  (void)triggered_with_keyboard;
  (void)userdata;
  return TRUE;
}

WEBUI_API int webui_init(struct webui *w) {
  if (gtk_init_check(0, NULL) == FALSE) {
    return -1;
  }

  w->priv.ready = 0;
  w->priv.should_exit = 0;
  w->priv.queue = g_async_queue_new();
  w->priv.window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
  gtk_window_set_title(GTK_WINDOW(w->priv.window), w->title);

  switch (w->border){
    case WEBUI_BORDER_SIZABLE:
      gtk_window_set_default_size(GTK_WINDOW(w->priv.window), w->width, w->height);
      gtk_window_set_resizable(GTK_WINDOW(w->priv.window), TRUE);
      break;
    case WEBUI_BORDER_DIALOG:
      gtk_widget_set_size_request(w->priv.window, w->width, w->height);
      gtk_window_set_resizable(GTK_WINDOW(w->priv.window), FALSE);
      break;
    default:// for none border
      gtk_widget_set_size_request(w->priv.window, w->width, w->height);
      gtk_window_set_resizable(GTK_WINDOW(w->priv.window), FALSE);
      gtk_window_set_decorated (GTK_WINDOW (w->priv.window), FALSE);      
      break;
  }
  if(w->minHeight!=0 || w->minHeight!=0){
    gtk_widget_set_size_request(GTK_WIDGET(w->priv.window),w->minWidth,w->minHeight);
  }
  gtk_window_set_position(GTK_WINDOW(w->priv.window), GTK_WIN_POS_CENTER);

  w->priv.scroller = gtk_scrolled_window_new(NULL, NULL);
  gtk_container_add(GTK_CONTAINER(w->priv.window), w->priv.scroller);

  WebKitUserContentManager *m = webkit_user_content_manager_new();
  webkit_user_content_manager_register_script_message_handler(m, "external");
  g_signal_connect(m, "script-message-received::external",
                   G_CALLBACK(external_message_received_cb), w);

  w->priv.webui = webkit_web_view_new_with_user_content_manager(m);
  webkit_web_view_load_uri(WEBKIT_WEB_VIEW(w->priv.webui),
                           webui_check_url(w->url));
  g_signal_connect(G_OBJECT(w->priv.webui), "load-changed",
                   G_CALLBACK(webui_load_changed_cb), w);
  gtk_container_add(GTK_CONTAINER(w->priv.scroller), w->priv.webui);

  if (w->debug) {
    WebKitSettings *settings =
        webkit_web_view_get_settings(WEBKIT_WEB_VIEW(w->priv.webui));
    webkit_settings_set_enable_write_console_messages_to_stdout(settings, true);
    webkit_settings_set_enable_developer_extras(settings, true);
  } else {
    g_signal_connect(G_OBJECT(w->priv.webui), "context-menu",
                     G_CALLBACK(webui_context_menu_cb), w);
  }

  gtk_widget_show_all(w->priv.window);

  webkit_web_view_run_javascript(
      WEBKIT_WEB_VIEW(w->priv.webui),
      "window.external={invoke:function(x){"
      "window.webkit.messageHandlers.external.postMessage(x);}}",
      NULL, NULL, NULL);

  g_signal_connect(G_OBJECT(w->priv.window), "destroy",
                   G_CALLBACK(webui_destroy_cb), w);
  g_signal_connect(G_OBJECT(w->priv.window), "delete-event",
                   G_CALLBACK(webui_delete_event_cb), w);
                   
  return 0;
}

WEBUI_API int webui_loop(struct webui *w, int blocking) {
  gtk_main_iteration_do(blocking);
  return w->priv.should_exit;
}

WEBUI_API void webui_set_title(struct webui *w, const char *title) {
  gtk_window_set_title(GTK_WINDOW(w->priv.window), title);
}

WEBUI_API void webui_set_fullscreen(struct webui *w, int fullscreen) {
  if (fullscreen) {
    gtk_window_fullscreen(GTK_WINDOW(w->priv.window));
  } else {
    gtk_window_unfullscreen(GTK_WINDOW(w->priv.window));
  }
}

WEBUI_API void webui_set_color(struct webui *w, uint8_t r, uint8_t g,
                                   uint8_t b, uint8_t a) {
  GdkRGBA color = {r / 255.0, g / 255.0, b / 255.0, a / 255.0};
  webkit_web_view_set_background_color(WEBKIT_WEB_VIEW(w->priv.webui),
                                       &color);
}
WEBUI_API int webui_msg(struct webui *w,enum webui_msg_type flag,const char *title,const char *msg){
  GtkWidget *dlg;
  GtkMessageType type = GTK_MESSAGE_OTHER;
  switch (flag & WEBUI_MSG_ICON_MASK){
  case WEBUI_MSG_INFO:
    type=GTK_MESSAGE_INFO;
    break;
  case WEBUI_MSG_WARNING:
    type=GTK_MESSAGE_WARNING;
    break;
  case WEBUI_MSG_ERROR:
    type=GTK_MESSAGE_ERROR;
    break;
  default://WEBUI_MSG_MSG = 0
    break;
  }
  dlg = gtk_message_dialog_new(GTK_WINDOW(w->priv.window), GTK_DIALOG_MODAL,
                                 type, GTK_BUTTONS_NONE, "%s", title);
  gtk_message_dialog_format_secondary_text(GTK_MESSAGE_DIALOG(dlg), "%s",msg);
  switch (flag & WEBUI_MSG_BUTTON_MASK){
  case WEBUI_MSG_OK_CANCEL:
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Ok",WEBUI_RESPONSE_OK);
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Cancel",WEBUI_RESPONSE_CANCEL);
    break;
  case WEBUI_MSG_YES_NO:
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Yes",WEBUI_RESPONSE_YES);
    gtk_dialog_add_button(GTK_DIALOG(dlg),"No",WEBUI_RESPONSE_NO);
    break;
  case WEBUI_MSG_YES_NO_CANCEL:
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Yes",WEBUI_RESPONSE_YES);
    gtk_dialog_add_button(GTK_DIALOG(dlg),"No",WEBUI_RESPONSE_NO);
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Cancel",WEBUI_RESPONSE_CANCEL);
    break;
  default://WEBUI_MSG_MSG = 0
    gtk_dialog_add_button(GTK_DIALOG(dlg),"Ok",WEBUI_RESPONSE_OK);
    break;
  }
  int res=gtk_dialog_run(GTK_DIALOG(dlg));
  
  gtk_widget_destroy(dlg);
  return res;
}

WEBUI_API void webui_dialog(struct webui *w,
                                enum webui_dialog_type dlgtype, int flags,
                                const char *title, const char *arg,
                                char *result, size_t resultsz) {
  GtkWidget *dlg;
  if (result != NULL) {
    result[0] = '\0';
  }
  if (dlgtype == WEBUI_DIALOG_TYPE_OPEN ||
      dlgtype == WEBUI_DIALOG_TYPE_SAVE) {
    dlg = gtk_file_chooser_dialog_new(
        title, GTK_WINDOW(w->priv.window),
        (dlgtype == WEBUI_DIALOG_TYPE_OPEN
             ? (flags & WEBUI_DIALOG_FLAG_DIRECTORY
                    ? GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER
                    : GTK_FILE_CHOOSER_ACTION_OPEN)
             : GTK_FILE_CHOOSER_ACTION_SAVE),
        "_Cancel", GTK_RESPONSE_CANCEL,
        (dlgtype == WEBUI_DIALOG_TYPE_OPEN ? "_Open" : "_Save"),
        GTK_RESPONSE_ACCEPT, NULL);
    gtk_file_chooser_set_local_only(GTK_FILE_CHOOSER(dlg), FALSE);
    gtk_file_chooser_set_select_multiple(GTK_FILE_CHOOSER(dlg), FALSE);
    gtk_file_chooser_set_show_hidden(GTK_FILE_CHOOSER(dlg), TRUE);
    gtk_file_chooser_set_do_overwrite_confirmation(GTK_FILE_CHOOSER(dlg), TRUE);
    gtk_file_chooser_set_create_folders(GTK_FILE_CHOOSER(dlg), TRUE);
    gint response = gtk_dialog_run(GTK_DIALOG(dlg));
    if (response == GTK_RESPONSE_ACCEPT) {
      gchar *filename = gtk_file_chooser_get_filename(GTK_FILE_CHOOSER(dlg));
      g_strlcpy(result, filename, resultsz);
      g_free(filename);
    }
    gtk_widget_destroy(dlg);
  } else if (dlgtype == WEBUI_DIALOG_TYPE_ALERT) {
    GtkMessageType type = GTK_MESSAGE_OTHER;
    switch (flags & WEBUI_DIALOG_FLAG_ALERT_MASK) {
    case WEBUI_DIALOG_FLAG_INFO:
      type = GTK_MESSAGE_INFO;
      break;
    case WEBUI_DIALOG_FLAG_WARNING:
      type = GTK_MESSAGE_WARNING;
      break;
    case WEBUI_DIALOG_FLAG_ERROR:
      type = GTK_MESSAGE_ERROR;
      break;
    }
    dlg = gtk_message_dialog_new(GTK_WINDOW(w->priv.window), GTK_DIALOG_MODAL,
                                 type, GTK_BUTTONS_OK, "%s", title);
    gtk_message_dialog_format_secondary_text(GTK_MESSAGE_DIALOG(dlg), "%s",
                                             arg);
    gtk_dialog_run(GTK_DIALOG(dlg));
    gtk_widget_destroy(dlg);
  }
}

static void webui_eval_finished(GObject *object, GAsyncResult *result,
                                  gpointer userdata) {
  (void)object;
  (void)result;
  struct webui *w = (struct webui *)userdata;
  w->priv.js_busy = 0;
}

WEBUI_API int webui_eval(struct webui *w, const char *js) {
  while (w->priv.ready == 0) {
    g_main_context_iteration(NULL, TRUE);
  }
  w->priv.js_busy = 1;
  webkit_web_view_run_javascript(WEBKIT_WEB_VIEW(w->priv.webui), js, NULL,
                                 webui_eval_finished, w);
  while (w->priv.js_busy) {
    g_main_context_iteration(NULL, TRUE);
  }
  return 0;
}

static gboolean webui_dispatch_wrapper(gpointer userdata) {
  struct webui *w = (struct webui *)userdata;
  for (;;) {
    struct webui_dispatch_arg *arg =
        (struct webui_dispatch_arg *)g_async_queue_try_pop(w->priv.queue);
    if (arg == NULL) {
      break;
    }
    (arg->fn)(w, arg->arg);
    g_free(arg);
  }
  return FALSE;
}

WEBUI_API void webui_dispatch(struct webui *w, webui_dispatch_fn fn,
                                  void *arg) {
  struct webui_dispatch_arg *context =
      (struct webui_dispatch_arg *)g_new(struct webui_dispatch_arg, 1);
  context->w = w;
  context->arg = arg;
  context->fn = fn;
  g_async_queue_lock(w->priv.queue);
  g_async_queue_push_unlocked(w->priv.queue, context);
  if (g_async_queue_length_unlocked(w->priv.queue) == 1) {
    gdk_threads_add_idle(webui_dispatch_wrapper, w);
  }
  g_async_queue_unlock(w->priv.queue);
}

WEBUI_API void webui_terminate(struct webui *w) {
  w->priv.should_exit = 1;
}

WEBUI_API void webui_exit(struct webui *w) { (void)w; }
WEBUI_API void webui_print_log(const char *s) {
  fprintf(stderr, "%s\n", s);
}

WEBUI_API void webui_set_min_size(struct webui *w,int width,int height){
  w->minWidth=width;
  w->minHeight=height;
  gtk_widget_set_size_request(GTK_WIDGET(w->priv.window),width,height);
}