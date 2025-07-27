package handler

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pradiptaagus/go-chat-app/utils"
)

func ClientUI(w http.ResponseWriter, r *http.Request) {
	cwd, err := os.Getwd()
	utils.PanicIfError(err)

	path := filepath.Join(cwd, "resources", "template", "chat.html")

	t := template.Must(template.ParseFiles(path))

	t.ExecuteTemplate(w, "chat.html", "ws://"+r.Host+"/chat")
}
