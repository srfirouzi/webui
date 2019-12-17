TARGET = minimal

CFLAGS ?= -std=c99 -Wall -Wextra -pedantic -I../..

TARGET_OS ?= $(OS)
ifeq ($(TARGET_OS),Windows_NT)
	TARGET=minimal.exe
	WEBVIEW_LDFLAGS := -lole32 -lcomctl32 -loleaut32 -luuid -mwindows
else ifeq ($(shell uname -s),Linux)
	WEBVIEW_CFLAGS :=  $(shell pkg-config --cflags gtk+-3.0 webkit2gtk-4.0)
	WEBVIEW_LDFLAGS := $(shell pkg-config --libs gtk+-3.0 webkit2gtk-4.0)
endif

$(TARGET): main.c
	$(CC) $(CFLAGS) $(WEBVIEW_CFLAGS) main.c $(LDFLAGS) $(WEBVIEW_LDFLAGS) -o $@

clean:
	rm -f $(TARGET)

.PHONY: clean
