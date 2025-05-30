package database

import (
	"fmt"
	"time"
)

// generateID génère un ID unique
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// GetPath retourne le chemin de la collection
func (c *Collection) GetPath() string {
	return c.path
}
