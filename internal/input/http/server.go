package http

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"io"
	"net/http"
)

type imageRequest struct {
	Reader     io.Reader
	ImageBytes transaction.ImageBytes
	ImageData  transaction.ImageData
}

// Server SETS THE CONFIGURATION AND STARTS THE SERVER.
func Server(w *Webserver) {
	addr := fmt.Sprintf(":%d", w.port)
	handler := ServerMux(w)
	server := &http.Server{
		Addr:    addr,
		Handler: handler}
	w.Server = server
	err := ListenAndServe(server, w)

	if err != nil {
		//error handling should be implemented here.
	}
}

// ListenAndServe MAY LOOK USELESS, BUT ITS HERE IN CASE WE ARE USING HTTPS.
func ListenAndServe(s *http.Server, w *Webserver) error {
	//Check if http has https files and then start https

	if w.CertFile != "" && w.KeyFile != "" {
		err := s.ListenAndServeTLS(w.CertFile, w.KeyFile)
		if err != nil {
			return err
			//Report failure
		}
		return nil
	}

	err := s.ListenAndServe()
	if err != nil {
		return err
		//Report failure
	}
	return nil

}

//Just an initial handler, receives a request, parses it to make sure its fine and sends a request to the plugin to be sent to processor.
func (w Webserver) index(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
			fmt.Printf("ParseMultipartform() err: %v", err)
			return
		}
		r.ParseForm()
		file, _, err := r.FormFile("image")
		if err != nil {
			fmt.Fprintf(rw, "You should upload an image ")
			return
		}
		iData := make(transaction.ImageData)

		//Setting all parameters
		//some parameters might be in the form of an array which may cause security issues.
		for k := range r.Form {
			iData[k] = r.Form[k][0]
		}
		iData["url"] = r.URL.String()

		w.Requests <- imageRequest{
			Reader:     io.Reader(file),
			ImageBytes: nil,
			ImageData:  iData,
		}
		rw.WriteHeader(http.StatusOK)

	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}

// ServerMux will have handlers.
func ServerMux(w *Webserver) http.Handler {
	mux := http.NewServeMux()
	//This is just an initial phase of the middleware.
	//Accept images at all the provided urls.
	next := http.Handler(http.HandlerFunc(w.index))

	handlers, _ := w.Config.Get("urls", nil)
	h := handlers.Get().Data()
	for k := range h.(map[string]interface{}) {
		mux.Handle(k, next)
	}

	return mux
}
