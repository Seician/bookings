package dbrepo

import (
	context2 "context"
	"github.com/Seician/bookings/internal/models"
	"time"
)

func (m *mySqlDBRepo) AllUsers() bool {
	return true
}

//InsertReservation insert reservation into the database
func (m *mySqlDBRepo) InsertReservation(reservation models.Reservation) (int64, error) {
	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	var newId int64

	statement := `INSERT INTO reservations (first_name, last_name, email, phone,
					start_date, end_date, room_id, created_at, updated_at)
					values(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := m.DB.ExecContext(
		context, statement,
		reservation.FirstName,
		reservation.LastName,
		reservation.Email,
		reservation.Phone,
		reservation.StartDate,
		reservation.EndDate,
		reservation.RoomId,
		time.Now(),
		time.Now())
	if err != nil {
		return 0, err
	}

	newId, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return newId, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *mySqlDBRepo) InsertRoomRestriction(restriction models.RoomRestriction) error {
	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	statement := `INSERT INTO room_restrictions(start_date, end_date, room_id, reservation_id,
					created_at, updated_at, restriction_id) values (?,?,?,?,?,?,?)`

	_, err := m.DB.ExecContext(context, statement,
		restriction.StartDate,
		restriction.EndDate,
		restriction.RoomId,
		restriction.ReservationId,
		time.Now(),
		time.Now(),
		restriction.RestrictionId)

	if err != nil {
		return err
	}
	return nil
}
