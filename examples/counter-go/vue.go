// +build vue

package main

//go:generate go-bindata -o assets.go js/vue/... js/styles.css

import (
	"github.com/srfirouzi/webui"
)

var uiFrameworkName = "VueJS"

func loadUIFramework(w webui.WebUI) {
	// Inject Vue.js
	w.Eval(string(MustAsset("js/vue/vendor/vue.min.js")))
	// Inject app code
	w.Eval(string(MustAsset("js/vue/app.js")))
}
