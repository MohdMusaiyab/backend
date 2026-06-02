package models

import (
	"time"
)

// Theater represents a physical cinema building.
type Theater struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Location string `gorm:"not null"`
	Halls    []Hall `gorm:"foreignKey:TheaterID"` // Tells GORM that Halls belong to this Theater
}

// Hall represents a specific screening room inside a Theater.
type Hall struct {
	ID         uint   `gorm:"primaryKey"`
	TheaterID  uint   `gorm:"not null"`
	Name       string `gorm:"not null"`
	TotalSeats int    `gorm:"not null"`
}

// Movie represents the actual film.
type Movie struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Description string
	DurationMin int
}

// Showtime links a Movie to a Hall at a specific time.
type Showtime struct {
	ID        uint      `gorm:"primaryKey"`
	MovieID   uint      `gorm:"not null"`
	HallID    uint      `gorm:"not null"`
	StartTime time.Time `gorm:"not null"`
}

// ShowtimeSeat represents a physical seat's availability for a specific Showtime.
type ShowtimeSeat struct {
	ID         uint   `gorm:"primaryKey"`
	ShowtimeID uint   `gorm:"not null;uniqueIndex:idx_show_seat"`
	SeatNumber string `gorm:"not null;uniqueIndex:idx_show_seat"` // The 'idx_show_seat' name groups these two fields into a composite unique index.
	Status     string `gorm:"type:varchar(20);default:'AVAILABLE'"` // We force varchar(20) to save DB space instead of default Text.
}

// Booking is the actual reservation record.
type Booking struct {
	ID             uint      `gorm:"primaryKey"`
	ShowtimeSeatID uint      `gorm:"not null;unique"` // One seat availability row can only ever have ONE successful booking.
	UserEmail      string    `gorm:"not null"`
	Status         string    `gorm:"type:varchar(20);default:'PENDING'"`
	ExpiresAt      time.Time // This will be used by a background worker to cancel unpaid bookings.
	CreatedAt      time.Time // Automatically populated by GORM when the row is inserted.
}
