package dbrepo

import (
	"github.com/Seician/bookings/internal/models"
	"time"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

//InsertReservation insert reservation into the database
func (m *testDBRepo) InsertReservation(reservation models.Reservation) (int64, error) {
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *testDBRepo) InsertRoomRestriction(restriction models.RoomRestriction) error {
	return nil
}

//SearchAvailabilityByDatesByRoomId returns true if availability exists for roomId, and false if no availability exists
func (m *testDBRepo) SearchAvailabilityByDatesByRoomId(start, end time.Time, roomId int) (bool, error) {

	return false, nil
}

//SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

func (m *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	return room, nil
}
