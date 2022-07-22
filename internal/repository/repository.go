package repository

import "github.com/Seician/bookings/internal/models"

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(reservation models.Reservation) (int64, error)
	InsertRoomRestriction(r models.RoomRestriction) error
}
