package dbrepo

import (
	context2 "context"
	"errors"
	"github.com/Seician/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

	result, err := m.DB.ExecContext(
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

	newId, err = result.LastInsertId()
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

//SearchAvailabilityByDatesByRoomId returns true if availability exists for roomId, and false if no availability exists
func (m *mySqlDBRepo) SearchAvailabilityByDatesByRoomId(start, end time.Time, roomId int) (bool, error) {

	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `SELECT 
    				count(id)
			   FROM
			   	    room_restrictions
                 WHERE 
                     room_id = ? AND
                     ? < end_date AND ? > start_date`

	row := m.DB.QueryRowContext(context, query, roomId, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}
	if numRows == 0 {
		return true, nil
	}
	return false, nil
}

//SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *mySqlDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT 
    				r.id, r.room_name
			   FROM
			   	    rooms r
                 WHERE 
                     r.id not in
                     (select room_id from room_restrictions rr where ? < rr.end_date and ? > rr.start_date);`

	rows, err := m.DB.QueryContext(context, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomById returns a room by id
func (m *mySqlDBRepo) GetRoomById(id int) (models.Room, error) {

	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id = ?`

	row := m.DB.QueryRowContext(context, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}
	return room, nil
}

// GetUserByID returns a user by id
func (m *mySqlDBRepo) GetUserByID(id int) (models.User, error) {

	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, password, access_level, created_at, updated_at
from users where id = ?`

	row := m.DB.QueryRowContext(context, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	return u, nil
}

// UpdateUser updates a user in the database
func (m *mySqlDBRepo) UpdateUser(u models.User) error {
	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	query := `
update users set first_name = ?, last_name=?, email=?, access_level=?, updated_at=?`

	_, err := m.DB.ExecContext(context, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now())
	if err != nil {
		return err
	}
	return nil
}

// Authenticate the user
func (m *mySqlDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	context, cancel := context2.WithTimeout(context2.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(context, "select id, password from bookings.users where email = ?", email)

	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
