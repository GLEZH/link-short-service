package entity

// URL links a short ID to its original address and owner
type URL struct {
	ID          string
	OriginalURL string
	UserID      string
	IsDeleted   bool
}
