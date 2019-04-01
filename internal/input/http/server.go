package http

import (
	"fmt"
	"net/http"
)

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
func (s Webserver) index(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		_, _, err := r.FormFile("image")
		if err != nil {
			fmt.Fprintf(w, "You should upload an image ")
			return
		}
		s.Requests <- *r
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ServerMux will have handlers.
func ServerMux(w *Webserver) http.Handler {
	mux := http.NewServeMux()
	//This is just an intial phase of the middlewares.
	next := http.Handler(http.HandlerFunc(w.index))
	mux.Handle("/index", next)
	return mux
}
