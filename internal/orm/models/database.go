package models

import (
	"fmt"
	"log"
	"nosql-db/internal/db"
	"nosql-db/internal/storage"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Configuration de la base de données
const (
	DB_PATH           = "data/database.json"
	SAVE_INTERVAL_SEC = 60 // Intervalle de sauvegarde en secondes
)

// DatabaseManager représente le gestionnaire de base de données
type DatabaseManager struct {
	database *db.Database
	models   map[string]ModelInterface
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewDatabaseManager crée un nouveau gestionnaire de base de données
func NewDatabaseManager() (*DatabaseManager, error) {
	database := db.NewDatabase()

	manager := &DatabaseManager{
		database: database,
		models:   make(map[string]ModelInterface),
		stopChan: make(chan struct{}),
	}

	// Initialiser les modèles
	if err := manager.initializeModels(); err != nil {
		return nil, err
	}

	return manager, nil
}

// initializeModels initialise tous les modèles
func (dm *DatabaseManager) initializeModels() error {
	// Initialiser le modèle User
	userModel := &UserModel{
		BaseModel: NewBaseModel(dm.database),
	}
	if err := userModel.Initialize(dm.database); err != nil {
		return fmt.Errorf("erreur lors de l'initialisation du modèle User: %w", err)
	}
	dm.models["users"] = userModel

	// Initialiser le modèle Product
	productModel := &ProductModel{
		BaseModel: NewBaseModel(dm.database),
	}
	if err := productModel.Initialize(dm.database); err != nil {
		return fmt.Errorf("erreur lors de l'initialisation du modèle Product: %w", err)
	}
	dm.models["products"] = productModel

	return nil
}

// GetModel retourne un modèle par son nom
func (dm *DatabaseManager) GetModel(name string) (ModelInterface, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	model, exists := dm.models[name]
	return model, exists
}

// Start démarre le gestionnaire de base de données
func (dm *DatabaseManager) Start() {
	// Démarrer la sauvegarde périodique
	go dm.periodicSave()

	// Gérer les signaux d'arrêt
	go dm.handleSignals()
}

// Stop arrête le gestionnaire de base de données
func (dm *DatabaseManager) Stop() {
	close(dm.stopChan)
	// Sauvegarder une dernière fois avant de quitter
	dm.saveDatabase()
}

// periodicSave sauvegarde périodiquement la base de données
func (dm *DatabaseManager) periodicSave() {
	ticker := time.NewTicker(SAVE_INTERVAL_SEC * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dm.saveDatabase()
		case <-dm.stopChan:
			return
		}
	}
}

// saveDatabase sauvegarde la base de données
func (dm *DatabaseManager) saveDatabase() {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if err := storage.SaveDatabase(dm.database); err != nil {
		log.Printf("Erreur lors de la sauvegarde de la base de données: %v", err)
	} else {
		log.Printf("Base de données sauvegardée avec succès à %s", time.Now().Format(time.RFC3339))
	}
}

// handleSignals gère les signaux d'arrêt
func (dm *DatabaseManager) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Signal d'arrêt reçu, arrêt du gestionnaire de base de données...")
	dm.Stop()
	os.Exit(0)
}

// AddModel ajoute un nouveau modèle à la base de données
func (dm *DatabaseManager) AddModel(name string, model ModelInterface) {
	dm.models[name] = model
}

// GetDatabase retourne la base de données associée au gestionnaire
func (dm *DatabaseManager) GetDatabase() *db.Database {
	return dm.database
}

// GetModelNames retourne la liste des noms de modèles
func (dm *DatabaseManager) GetModelNames() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	names := make([]string, 0, len(dm.models))
	for name := range dm.models {
		names = append(names, name)
	}
	return names
}
