// +build !vue,!react

package main

//go:generate go-bindata -o assets.go js/picodom/... js/styles.css

import (
	"github.com/srfirouzi/webui"
)

var uiFrameworkName = "Picodom"

func loadUIFramework(w webui.WebUI) {
	// Inject Picodom.js
	w.Eval(string(MustAsset("js/picodom/vendor/picodom.js")))
	// Inject app code
	w.Eval(string(MustAsset("js/picodom/app.js")))
}
