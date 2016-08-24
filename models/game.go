package models

import "time"

// Game model
type Game struct {
	ID          string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Timestamp   time.Time `json:"timestamp" gorethink:"timestamp"`
	Creator     string    `json:"creator" gorethink:"creator"`
	Faction     string    `json:"faction" gorethink:"faction"`
	Description string    `json:"description" gorethink:"description"`
	Accepted    bool      `json:"accepted" gorethink:"accepted"`
}
