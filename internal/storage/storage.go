package storage

import (
	"encoding/json"
	"nosql-db/internal/db"
	"os"
	"path/filepath"
)

const dataFile = "data/database.json"

// SaveDatabase sauvegarde l'état de la base de données dans un fichier
func SaveDatabase(database *db.Database) error {
	// Créer le dossier data s'il n'existe pas
	if err := os.MkdirAll(filepath.Dir(dataFile), 0755); err != nil {
		return err
	}

	// Convertir la base de données en JSON
	data, err := json.MarshalIndent(database, "", "  ")
	if err != nil {
		return err
	}

	// Écrire dans le fichier
	return os.WriteFile(dataFile, data, 0644)
}

// LoadDatabase charge l'état de la base de données depuis un fichier
func LoadDatabase() (*db.Database, error) {
	// Vérifier si le fichier existe
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return db.NewDatabase(), nil
	}

	// Lire le fichier
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}

	// Décoder le JSON
	database := &db.Database{}
	if err := json.Unmarshal(data, database); err != nil {
		return nil, err
	}

	return database, nil
}
