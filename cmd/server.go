package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func (srv *Server) Start() {
}

func (srv *Server) route() {}

func do(w http.ResponseWriter, r *http.Request) {

}

func ReadRequest(r *http.Request, data interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	_ = r.Body.Close()
	return json.Unmarshal(b, data)
}

type httpResponse struct {
	ErrorMsg string      `json:"error_msg"`
	Result   interface{} `json:"result"`
}

func (v httpResponse) ToByte() []byte {
	b, _ := json.Marshal(v)
	return b
}

func WriteResponse(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Context-Type", "application/json")

	r := httpResponse{
		ErrorMsg: "",
		Result:   data,
	}
	w.Write(r.ToByte())
}

func WriteResponseError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Context-Type", "application/json")
	r := httpResponse{
		ErrorMsg: msg,
		Result:   nil,
	}
	w.Write(r.ToByte())
}
