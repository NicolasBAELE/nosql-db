package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"nosql-db/internal/database"
)

var (
	db         *database.Database
	configPath string
)

type CollectionConfig struct {
	Name    string `json:"name"`
	Indexes []struct {
		Field  string `json:"field"`
		Unique bool   `json:"unique"`
	} `json:"indexes"`
}

type Config struct {
	Collections []CollectionConfig `json:"collections"`
}

type CollectionData struct {
	Name      string              `json:"name"`
	Documents []database.Document `json:"documents"`
}

type PageData struct {
	Collections []CollectionData
}

func main() {
	// Définir le flag pour le chemin du fichier de configuration
	flag.StringVar(&configPath, "config", "config/collections.json", "Chemin vers le fichier de configuration des collections")
	flag.Parse()

	// Créer une nouvelle instance de la base de données
	var err error
	db, err = database.NewDatabase("data")
	if err != nil {
		log.Fatalf("Erreur lors de la création de la base de données: %v", err)
	}

	// Charger la configuration
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration: %v", err)
	}

	// Créer les collections et leurs index
	for _, collectionConfig := range config.Collections {
		collection, err := db.CreateCollection(collectionConfig.Name)
		if err != nil {
			log.Printf("Erreur lors de la création de la collection %s: %v", collectionConfig.Name, err)
			continue
		}

		for _, indexConfig := range collectionConfig.Indexes {
			if err := collection.CreateIndex(indexConfig.Field, indexConfig.Unique); err != nil {
				log.Printf("Erreur lors de la création de l'index %s pour la collection %s: %v",
					indexConfig.Field, collectionConfig.Name, err)
			}
		}
	}

	// Créer un nouveau routeur
	mux := http.NewServeMux()

	// Configurer les routes API en premier
	for _, collectionConfig := range config.Collections {
		collectionName := collectionConfig.Name
		// Handler pour toutes les opérations CRUD sur la collection
		mux.HandleFunc(fmt.Sprintf("/api/%s", collectionName),
			func(w http.ResponseWriter, r *http.Request) {
				handleCollection(w, r, collectionName)
			})
		// Handler pour les opérations CRUD sur un document spécifique
		mux.HandleFunc(fmt.Sprintf("/api/%s/", collectionName),
			func(w http.ResponseWriter, r *http.Request) {
				handleCollection(w, r, collectionName)
			})
		// Handler pour la recherche
		mux.HandleFunc(fmt.Sprintf("/api/%s/search", collectionName),
			func(w http.ResponseWriter, r *http.Request) {
				handleCollectionSearch(w, r, collectionName)
			})
	}

	// Routes pour les transactions
	mux.HandleFunc("/api/transaction/begin", handleTransactionBegin)
	mux.HandleFunc("/api/transaction/commit", handleTransactionCommit)
	mux.HandleFunc("/api/transaction/rollback", handleTransactionRollback)
	mux.HandleFunc("/api/transaction/", handleTransactionOperation)

	// Configurer les routes pour les interfaces
	mux.HandleFunc("/", handleIndex)                         // Interface HTML classique
	mux.HandleFunc("/vue", handleVueApp)                     // Interface Vue.js
	mux.HandleFunc("/api/collections", handleCollectionsAPI) // API pour Vue.js

	// Démarrer le serveur avec le routeur personnalisé
	log.Println("Serveur démarré sur http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("index.html").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	})

	tmpl, err := tmpl.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupérer les données de toutes les collections
	var collections []CollectionData
	for name := range db.GetCollections() {
		collection, err := db.GetCollection(name)
		if err != nil {
			continue
		}

		documents, err := collection.GetAllDocuments()
		if err != nil {
			continue
		}

		collections = append(collections, CollectionData{
			Name:      name,
			Documents: documents,
		})
	}

	data := PageData{
		Collections: collections,
	}

	tmpl.Execute(w, data)
}

func handleVueApp(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/vue-app.html")
}

func handleCollectionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer les données de toutes les collections
	var collections []CollectionData
	for name := range db.GetCollections() {
		collection, err := db.GetCollection(name)
		if err != nil {
			continue
		}

		documents, err := collection.GetAllDocuments()
		if err != nil {
			continue
		}

		collections = append(collections, CollectionData{
			Name:      name,
			Documents: documents,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collections)
}

func handleCollection(w http.ResponseWriter, r *http.Request, collectionName string) {
	collection, err := db.GetCollection(collectionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Collection %s does not exist", collectionName), http.StatusNotFound)
		return
	}

	// Extraire l'ID du document de l'URL si présent
	parts := strings.Split(r.URL.Path, "/")
	var documentID string
	if len(parts) > 3 {
		documentID = parts[3]
	}

	switch r.Method {
	case http.MethodGet:
		if documentID != "" {
			// Récupérer un document spécifique
			doc, err := collection.FindByID(documentID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(doc)
		} else {
			// Lister tous les documents
			documents, err := collection.GetAllDocuments()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(documents)
		}

	case http.MethodPost:
		// Créer un nouveau document
		var doc database.Document
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := collection.Insert(doc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":       id,
			"document": doc,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	case http.MethodPut:
		if documentID == "" {
			http.Error(w, "Document ID is required for update", http.StatusBadRequest)
			return
		}

		// Mettre à jour un document
		var doc database.Document
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := collection.Update(documentID, doc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Retourner le document mis à jour
		updatedDoc, err := collection.FindByID(documentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updatedDoc)

	case http.MethodDelete:
		if documentID == "" {
			http.Error(w, "Document ID is required for deletion", http.StatusBadRequest)
			return
		}

		// Supprimer un document
		if err := collection.Delete(documentID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCollectionSearch gère les requêtes de recherche
func handleCollectionSearch(w http.ResponseWriter, r *http.Request, collectionName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collection, err := db.GetCollection(collectionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Collection %s does not exist", collectionName), http.StatusNotFound)
		return
	}

	// Récupérer les paramètres de recherche
	field := r.URL.Query().Get("field")
	value := r.URL.Query().Get("value")

	if field == "" || value == "" {
		http.Error(w, "Field and value parameters are required", http.StatusBadRequest)
		return
	}

	// Convertir la valeur en interface{} pour la recherche
	var searchValue interface{}
	if err := json.Unmarshal([]byte(value), &searchValue); err != nil {
		// Si la conversion JSON échoue, utiliser la valeur comme string
		searchValue = value
	}

	// Rechercher les documents
	documents, err := collection.FindByField(field, searchValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(documents)
}

// Handlers pour les transactions

// handleTransactionBegin commence une nouvelle transaction
func handleTransactionBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tx := db.BeginTransaction()
	response := map[string]interface{}{
		"transaction_id": tx.ID,
		"status":         "active",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleTransactionCommit valide une transaction
func handleTransactionCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, exists := db.GetTransaction(request.TransactionID)
	if !exists {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if err := db.Commit(tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"transaction_id": request.TransactionID,
		"status":         "committed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTransactionRollback annule une transaction
func handleTransactionRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, exists := db.GetTransaction(request.TransactionID)
	if !exists {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if err := db.Rollback(tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"transaction_id": request.TransactionID,
		"status":         "rolled_back",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTransactionOperation gère les opérations dans une transaction
func handleTransactionOperation(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid transaction operation path", http.StatusBadRequest)
		return
	}

	transactionID := parts[3]
	operation := parts[4]

	tx, exists := db.GetTransaction(transactionID)
	if !exists {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	switch operation {
	case "insert":
		handleTransactionInsert(w, r, tx)
	case "update":
		handleTransactionUpdate(w, r, tx)
	case "delete":
		handleTransactionDelete(w, r, tx)
	default:
		http.Error(w, "Unknown transaction operation", http.StatusBadRequest)
	}
}

// handleTransactionInsert insère un document dans une transaction
func handleTransactionInsert(w http.ResponseWriter, r *http.Request, tx *database.Transaction) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Collection string            `json:"collection"`
		Document   database.Document `json:"document"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.InsertWithTransaction(tx, request.Collection, request.Document)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":       id,
		"document": request.Document,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleTransactionUpdate met à jour un document dans une transaction
func handleTransactionUpdate(w http.ResponseWriter, r *http.Request, tx *database.Transaction) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Collection string            `json:"collection"`
		DocumentID string            `json:"document_id"`
		Updates    database.Document `json:"updates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection, err := db.GetCollection(request.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("Collection %s does not exist", request.Collection), http.StatusNotFound)
		return
	}

	if err := db.UpdateWithTransaction(tx, request.Collection, request.DocumentID, request.Updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retourner le document mis à jour
	updatedDoc, err := collection.FindByID(request.DocumentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDoc)
}

// handleTransactionDelete supprime un document dans une transaction
func handleTransactionDelete(w http.ResponseWriter, r *http.Request, tx *database.Transaction) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Collection string `json:"collection"`
		DocumentID string `json:"document_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DeleteWithTransaction(tx, request.Collection, request.DocumentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
