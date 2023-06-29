package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/emoji-api/api/controller"
	oapigen "github.com/podhmo/emoji-api/api/oapigen"
)

// TODO: 404 with application/json

func TestEmojiTranslate(t *testing.T) {
	c := &controller.ControllerController{} // uggly name
	c.EmojiController = controller.NewEmojiController()
	h := newHandler(c)

	req, _ := http.NewRequest("POST", "/emoji/translate", bytes.NewBufferString(`{"text": "hmm :dizzy:"}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()

	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Fatalf("status code: want=%d, but got=%d", want, got)
	}

	var got string
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Errorf("unexpected error (json.Unmarshal): %+v", err)
	}
	defer res.Body.Close()

	want := `hmm 💫`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("response body, mismatch (-want +got):\n%s", diff)
	}
}

func newHandler(ssi oapigen.StrictServerInterface) http.Handler {
	router := chi.NewRouter()
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		router.Use(middleware.Logger)
	}
	return oapigen.HandlerFromMux(oapigen.NewStrictHandler(ssi, nil), router)
}
