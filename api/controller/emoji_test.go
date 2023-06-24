package controller_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/podhmo/emoji-api/api/controller"
	oapigen "github.com/podhmo/emoji-api/api/oapigen"
)

// TODO: 404 with application/json

func TestEmojiTranslate(t *testing.T) {
	c := &controller.ControllerController{} // uggly name
	c.EmojiController = controller.NewEmojiController()
	h := newHandler(c)

	ts := httptest.NewServer(h)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/emoji/translate", bytes.NewBufferString(`{"text": "hmm :dizzy:"}`))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	b, err2 := httputil.DumpResponse(res, true)
	fmt.Fprintln(os.Stderr, "----------------------------------------")
	fmt.Fprintln(os.Stderr, string(b), err, err2)
	fmt.Fprintln(os.Stderr, "----------------------------------------")
}

func newHandler(ssi oapigen.StrictServerInterface) http.Handler {
	router := chi.NewRouter()
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		router.Use(middleware.Logger)
	}
	return oapigen.HandlerFromMux(oapigen.NewStrictHandler(ssi, nil), router)
}
