package controller

import (
	"net/http"

	"github.com/pradiptaagus/go-chat-app/resource"
)

var renderer = resource.NewRenderer()

func ClientUI(w http.ResponseWriter, r *http.Request) {

	renderer.Render(w, "chat.html", "ws://"+r.Host+"/chat")
}
