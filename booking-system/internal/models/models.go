package models

import (
	"time"
)


type Theater struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Location string `gorm:"not null"`
	Halls    []Hall `gorm:"foreignKey:TheaterID"` 
}


type Hall struct {
	ID         uint   `gorm:"primaryKey"`
	TheaterID  uint   `gorm:"not null"`
	Name       string `gorm:"not null"`
	TotalSeats int    `gorm:"not null"`
}


type Movie struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Description string
	DurationMin int
}

type Showtime struct {
	ID        uint      `gorm:"primaryKey"`
	MovieID   uint      `gorm:"not null"`
	HallID    uint      `gorm:"not null"`
	StartTime time.Time `gorm:"not null"`
}


type ShowtimeSeat struct {
	ID         uint   `gorm:"primaryKey"`
	ShowtimeID uint   `gorm:"not null;uniqueIndex:idx_show_seat"`
	SeatNumber string `gorm:"not null;uniqueIndex:idx_show_seat"` 
	Status     string `gorm:"type:varchar(20);default:'AVAILABLE'"`
}


type Booking struct {
	ID             uint      `gorm:"primaryKey"`
	ShowtimeSeatID uint      `gorm:"not null;unique"` 
	UserEmail      string    `gorm:"not null"`
	Status         string    `gorm:"type:varchar(20);default:'PENDING'"`
	ExpiresAt      time.Time 
	CreatedAt      time.Time 
}
