package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Seician/bookings/internal/config"
	"github.com/Seician/bookings/internal/driver"
	"github.com/Seician/bookings/internal/forms"
	"github.com/Seician/bookings/internal/helpers"
	"github.com/Seician/bookings/internal/models"
	"github.com/Seician/bookings/internal/render"
	"github.com/Seician/bookings/internal/repository"
	"github.com/Seician/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi"
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
		helpers.ServerError(writer, errors.New("cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomById(reservation.RoomId)
	if err != nil {
		helpers.ServerError(writer, err)
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
func (m *Repository) PostReservation(writer http.ResponseWriter, request *http.Request) {
	reservation, ok := m.App.Session.Get(request.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(writer, errors.New("cannot get from session"))
		return
	}
	err := request.ParseForm()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}

	reservation.FirstName = request.Form.Get("first_name")
	reservation.LastName = request.Form.Get("last_name")
	reservation.Phone = request.Form.Get("phone")
	reservation.Email = request.Form.Get("email")

	form := forms.New(request.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(writer, request, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationId: int(newReservationID),
		RestrictionId: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	m.App.Session.Put(request.Context(), "reservation", reservation)

	http.Redirect(writer, request, "/reservation-summary", http.StatusSeeOther)
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

	starDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	rooms, err := m.DB.SearchAvailabilityForAllRooms(starDate, endDate)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	if len(rooms) == 0 {
		// no availability
		m.App.Session.Put(request.Context(), "error", "No availability")
		http.Redirect(writer, request, "/search-availability", http.StatusSeeOther)
	}

	data := make(map[string]interface{})

	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: starDate,
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

	sd := request.Form.Get("start")
	ed := request.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(request.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, roomID)
	response := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomID),
	}
	out, err := json.MarshalIndent(response, "", "      ")
	if err != nil {
		helpers.ServerError(writer, err)
	}
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

// BookRoom takes URL parameters, builds a sessional variabl, and take user to make res screen
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
		helpers.ServerError(writer, err)
		return
	}

	res.Room.RoomName = room.RoomName

	res.RoomId = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(request.Context(), "reservation", res)
	http.Redirect(writer, request, "/make-reservation", http.StatusSeeOther)
}
