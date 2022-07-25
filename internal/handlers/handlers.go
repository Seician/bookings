package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Seician/bookings/internal/config"
	"github.com/Seician/bookings/internal/driver"
	"github.com/Seician/bookings/internal/forms"
	"github.com/Seician/bookings/internal/helpers"
	"github.com/Seician/bookings/internal/models"
	"github.com/Seician/bookings/internal/render"
	"github.com/Seician/bookings/internal/repository"
	"github.com/Seician/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewMySqlRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(writer http.ResponseWriter, request *http.Request) {

	reservation, ok := m.App.Session.Get(request.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(request.Context(), "error", "can't get reservation from session")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomById(reservation.RoomId)
	if err != nil {
		m.App.Session.Put(request.Context(), "error", "can't find room")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.Room.RoomName = room.RoomName

	m.App.Session.Put(request.Context(), "reservation", reservation)

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(writer, request, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	// 2020-01-01 -- 01/02 03:04:05PM '06 -0700

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomId:    roomID,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "my own error message", http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomId:        roomID,
		ReservationId: int(newReservationID),
		RestrictionId: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	htmlMessage := fmt.Sprintf(`
  <h1> Reservation confirmation </h1>
`)

	// send notifications - first to guest
	msg := models.MailData{
		To:      reservation.Email,
		From:    "me@yahoo.com",
		Subject: "Reservation confirmation",
		Content: htmlMessage,
	}
	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the search availability page
func (m *Repository) PostAvailability(writer http.ResponseWriter, request *http.Request) {

	start := request.Form.Get("start")
	end := request.Form.Get("end")
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(request.Context(), "error", "can't parse start date!")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(request.Context(), "error", "can't parse end date!")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(request.Context(), "error", "can't get availability for rooms")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}
	if len(rooms) == 0 {
		// no availability
		m.App.Session.Put(request.Context(), "error", "No availability")
		http.Redirect(writer, request, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})

	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(request.Context(), "reservation", res)

	render.Template(writer, request, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"roomId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(writer http.ResponseWriter, request *http.Request) {

	// need to parse request body
	err := request.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal Server Error",
		}

		out, _ := json.MarshalIndent(resp, "", "      ")
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(out)
		return
	}
	sd := request.Form.Get("start")
	ed := request.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(request.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, roomID)
	if err != nil {
		// got a database error, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error querying database",
		}

		out, _ := json.MarshalIndent(resp, "", "      ")
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(out)
		return
	}
	response := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(response, "", "      ")

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(writer http.ResponseWriter, request *http.Request) {
	reservation, ok := m.App.Session.Get(request.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(request.Context(), "error", "Can't get reservation from session")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)

		return
	}

	m.App.Session.Remove(request.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.StartDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	render.Template(writer, request, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom display list of available rooms
func (m *Repository) ChooseRoom(writer http.ResponseWriter, request *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}

	res, ok := m.App.Session.Get(request.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(writer, err)
		return
	}

	res.RoomId = roomId

	m.App.Session.Put(request.Context(), "reservation", res)

	http.Redirect(writer, request, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a sessional variable, and take user to make res screen
func (m *Repository) BookRoom(writer http.ResponseWriter, request *http.Request) {
	// id, s, e

	roomID, _ := strconv.Atoi(request.URL.Query().Get("id"))
	sd := request.URL.Query().Get("s")
	ed := request.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomById(roomID)
	if err != nil {
		m.App.Session.Put(request.Context(), "error", "Can't get room from db!")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	res.RoomId = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(request.Context(), "reservation", res)
	http.Redirect(writer, request, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(writer http.ResponseWriter, request *http.Request) {
	_ = m.App.Session.RenewToken(request.Context())

	err := request.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := request.Form.Get("email")
	password := request.Form.Get("password")

	form := forms.New(request.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		m.App.Session.Put(request.Context(), "error", "Invalid login credentials")
		http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Print(err)
	}

	m.App.Session.Put(request.Context(), "user_id", id)
	m.App.Session.Put(request.Context(), "flash", "Logged successfully")
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}
