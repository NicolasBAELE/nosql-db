package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TransactionState représente l'état d'une transaction
type TransactionState int

const (
	TransactionActive TransactionState = iota
	TransactionCommitted
	TransactionAborted
)

// OperationType représente le type d'opération dans une transaction
type OperationType string

const (
	OpInsert OperationType = "INSERT"
	OpUpdate OperationType = "UPDATE"
	OpDelete OperationType = "DELETE"
)

// LogEntry représente une entrée dans le WAL
type LogEntry struct {
	TransactionID string                 `json:"transaction_id"`
	Timestamp     int64                  `json:"timestamp"`
	Operation     OperationType          `json:"operation"`
	Collection    string                 `json:"collection"`
	DocumentID    string                 `json:"document_id,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	OldData       map[string]interface{} `json:"old_data,omitempty"`
}

// Transaction représente une transaction
type Transaction struct {
	ID        string           `json:"id"`
	State     TransactionState `json:"state"`
	StartTime int64            `json:"start_time"`
	Log       []LogEntry       `json:"log"`
	walPath   string           `json:"-"` // Chemin vers le répertoire WAL
	mu        sync.RWMutex
}

// TransactionManager gère les transactions
type TransactionManager struct {
	transactions map[string]*Transaction
	walPath      string
	mu           sync.RWMutex
}

// NewTransactionManager crée un nouveau gestionnaire de transactions
func NewTransactionManager(walPath string) (*TransactionManager, error) {
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, fmt.Errorf("erreur création répertoire WAL: %v", err)
	}

	tm := &TransactionManager{
		transactions: make(map[string]*Transaction),
		walPath:      walPath,
	}

	// Récupérer les transactions non terminées au démarrage
	if err := tm.recoverTransactions(); err != nil {
		return nil, fmt.Errorf("erreur récupération transactions: %v", err)
	}

	return tm, nil
}

// BeginTransaction commence une nouvelle transaction
func (tm *TransactionManager) BeginTransaction() *Transaction {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx := &Transaction{
		ID:        generateTransactionID(),
		State:     TransactionActive,
		StartTime: time.Now().UnixNano(),
		Log:       make([]LogEntry, 0),
		walPath:   tm.walPath,
	}

	tm.transactions[tx.ID] = tx
	return tx
}

// Commit valide une transaction
func (tm *TransactionManager) Commit(db *Database, tx *Transaction) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TransactionActive {
		return fmt.Errorf("transaction %s n'est pas active", tx.ID)
	}

	// Appliquer le log de la transaction (déferred writing)
	if err := db.ApplyTransactionLog(tx); err != nil {
		return fmt.Errorf("erreur application log transaction: %v", err)
	}

	// Marquer la transaction comme validée
	tx.State = TransactionCommitted

	// Nettoyer les fichiers WAL de cette transaction
	if err := tm.cleanupTransactionWAL(tx.ID); err != nil {
		fmt.Printf("Avertissement: erreur nettoyage WAL pour transaction %s: %v\n", tx.ID, err)
	}

	// Nettoyer la transaction
	tm.mu.Lock()
	delete(tm.transactions, tx.ID)
	tm.mu.Unlock()

	return nil
}

// Rollback annule une transaction
func (tm *TransactionManager) Rollback(tx *Transaction) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TransactionActive {
		return fmt.Errorf("transaction %s n'est pas active", tx.ID)
	}

	// Marquer la transaction comme annulée
	tx.State = TransactionAborted

	// Nettoyer les fichiers WAL de cette transaction
	if err := tm.cleanupTransactionWAL(tx.ID); err != nil {
		fmt.Printf("Avertissement: erreur nettoyage WAL pour transaction %s: %v\n", tx.ID, err)
	}

	// Nettoyer la transaction
	tm.mu.Lock()
	delete(tm.transactions, tx.ID)
	tm.mu.Unlock()

	return nil
}

// AddLogEntry ajoute une entrée de log à une transaction
func (tx *Transaction) AddLogEntry(entry LogEntry) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	entry.TransactionID = tx.ID
	entry.Timestamp = time.Now().UnixNano()
	tx.Log = append(tx.Log, entry)

	// Écrire immédiatement dans le WAL pour la durabilité
	if err := tx.writeLogEntryToWAL(entry); err != nil {
		fmt.Printf("Erreur écriture WAL pour transaction %s: %v\n", tx.ID, err)
	}
}

// writeLogEntry écrit une entrée dans le WAL
func (tm *TransactionManager) writeLogEntry(entry LogEntry) error {
	filename := fmt.Sprintf("%s_%s.log", entry.TransactionID, fmt.Sprintf("%d", entry.Timestamp))
	path := filepath.Join(tm.walPath, filename)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("erreur sérialisation log: %v", err)
	}

	return os.WriteFile(path, data, 0644)
}

// writeLogEntryToWAL écrit une entrée dans le WAL (méthode de Transaction)
func (tx *Transaction) writeLogEntryToWAL(entry LogEntry) error {
	filename := fmt.Sprintf("%s_%s.log", entry.TransactionID, fmt.Sprintf("%d", entry.Timestamp))
	path := filepath.Join(tx.walPath, filename)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("erreur sérialisation log: %v", err)
	}

	return os.WriteFile(path, data, 0644)
}

// recoverTransactions récupère les transactions non terminées
func (tm *TransactionManager) recoverTransactions() error {
	files, err := os.ReadDir(tm.walPath)
	if err != nil {
		return err
	}

	// Grouper les fichiers par transaction ID
	transactionFiles := make(map[string][]string)
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".log" {
			continue
		}

		// Extraire l'ID de transaction du nom de fichier
		// Format: tx_ID_timestamp.log
		parts := strings.Split(file.Name(), "_")
		if len(parts) >= 3 {
			transactionID := parts[0] + "_" + parts[1]
			transactionFiles[transactionID] = append(transactionFiles[transactionID], file.Name())
		}
	}

	// Pour chaque transaction, analyser les logs et décider de l'action
	for transactionID, fileList := range transactionFiles {
		fmt.Printf("Récupération: transaction %s avec %d fichiers WAL\n", transactionID, len(fileList))

		// Pour l'instant, on supprime les anciens logs
		// Dans une implémentation complète, on analyserait les logs pour récupérer l'état
		for _, fileName := range fileList {
			filePath := filepath.Join(tm.walPath, fileName)
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Erreur suppression log %s: %v\n", fileName, err)
			} else {
				fmt.Printf("Supprimé: %s\n", fileName)
			}
		}
	}

	return nil
}

// generateTransactionID génère un ID unique pour une transaction
func generateTransactionID() string {
	return fmt.Sprintf("tx_%x", time.Now().UnixNano())
}

// GetTransaction récupère une transaction par son ID
func (tm *TransactionManager) GetTransaction(id string) (*Transaction, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tx, exists := tm.transactions[id]
	return tx, exists
}

// cleanupTransactionWAL nettoie les fichiers WAL d'une transaction spécifique
func (tm *TransactionManager) cleanupTransactionWAL(transactionID string) error {
	files, err := os.ReadDir(tm.walPath)
	if err != nil {
		return err
	}

	deletedCount := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".log" {
			continue
		}

		// Vérifier si le fichier appartient à cette transaction
		if strings.HasPrefix(file.Name(), transactionID+"_") {
			filePath := filepath.Join(tm.walPath, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("erreur suppression log %s: %v", file.Name(), err)
			}
			deletedCount++
			fmt.Printf("Nettoyage WAL: supprimé %s\n", file.Name())
		}
	}

	if deletedCount > 0 {
		fmt.Printf("Nettoyage WAL: %d fichiers supprimés pour transaction %s\n", deletedCount, transactionID)
	} else {
		fmt.Printf("Nettoyage WAL: aucun fichier trouvé pour transaction %s\n", transactionID)
	}

	return nil
}

// CleanupWAL nettoie les anciens fichiers WAL
func (tm *TransactionManager) CleanupWAL() error {
	files, err := os.ReadDir(tm.walPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".log" {
			continue
		}

		// Supprimer les anciens logs
		filePath := filepath.Join(tm.walPath, file.Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Erreur suppression log %s: %v\n", file.Name(), err)
		}
	}

	return nil
}
