package homePage

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

const localServerAddress = "127.0.0.1:7879"

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}

	t.templ.Execute(w, data)
}

func HttpServer() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

	})
	http.Handle("/chat", &templateHandler{filename: "chat.html"})

	// start the web server
	log.Println("Starting web server on", localServerAddress)
	if err := http.ListenAndServe(localServerAddress, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
