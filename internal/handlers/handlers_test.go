package handlers

import (
	"context"
	"github.com/Seician/bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	//{"home", "/", "GET", []postData{}, http.StatusOK},
	//{"about", "/about", "GET", []postData{}, http.StatusOK},
	//{"generals-quarters", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	//{"majors-suite", "/majors-suite", "GET", []postData{}, http.StatusOK},
	//{"majors-suite", "/majors-suite", "GET", []postData{}, http.StatusOK},
	//{"search-availability", "/search-availability", "GET", []postData{}, http.StatusOK},
	//{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	//{"make-reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	//
	//{"post-search-avail", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2022-01-09"},
	//	{key: "end", value: "2022-01-10"},
	//}, http.StatusOK},
	//
	//{"post-search-avail-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2022-01-09"},
	//	{key: "end", value: "2022-01-10"},
	//}, http.StatusOK}, {"make-reservation", "/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "Seician"},
	//	{key: "last_name", value: "Aurel"},
	//	{key: "email", value: "me@yahoo.com"},
	//	{key: "phone", value: "555-555-555"},
	//}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			response, err := testServer.Client().Get(testServer.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			if response.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, response.StatusCode)
			}
		} else {
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			response, err := testServer.Client().PostForm(testServer.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			if response.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, response.StatusCode)
			}
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	request, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getContext(request)
	request = request.WithContext(ctx)

	requestRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(requestRecorder, request)

	if requestRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrog response code: got %d, wanted %d", requestRecorder.Code, http.StatusOK)
	}

}

func getContext(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Print(err)
	}
	return ctx
}
