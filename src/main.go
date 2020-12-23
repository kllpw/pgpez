package main

import (
	"github.com/kllpw/pgpez/src/pkg/restmanger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/webview/webview"
)

func main() {
	mngr := restmanger.NewManager()
	go mngr.Start()
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("pgpez")
	w.SetSize(1024, 768, webview.HintFixed)
	w.Navigate("http://localhost")
	w.Run()
}
