package main

import "github.com/srfirouzi/webui"

func main() {
	// Open wikipedia in a 800x600 resizable window
	webui.Open("Minimal webui example",
		"https://en.m.wikipedia.org/wiki/Main_Page", 800, 600, webui.BorderResizable)
}
