package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "http://localhost:8081"

func main() {
	fmt.Println("🧪 Test de nettoyage des fichiers WAL après transactions")
	fmt.Println(strings.Repeat("=", 60))

	// 1. Commencer une transaction
	fmt.Println("\n1. Commencer une transaction...")
	txResponse, err := http.Post(baseURL+"/api/transaction/begin", "application/json", nil)
	if err != nil {
		log.Fatal("Erreur lors du début de transaction:", err)
	}

	var txResult map[string]interface{}
	json.NewDecoder(txResponse.Body).Decode(&txResult)
	txResponse.Body.Close()

	transactionID := txResult["transaction_id"].(string)
	fmt.Printf("Transaction ID: %s\n", transactionID)

	// 2. Insérer un document
	fmt.Println("\n2. Insérer un document...")
	insertData := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Test WAL Cleanup",
			"iban":        111111111,
			"description": "Document pour tester le nettoyage WAL",
		},
	}

	insertJSON, _ := json.Marshal(insertData)
	insertResp, err := http.Post(
		baseURL+"/api/transaction/"+transactionID+"/insert",
		"application/json",
		bytes.NewBuffer(insertJSON),
	)
	if err != nil {
		log.Fatal("Erreur lors de l'insertion:", err)
	}

	var insertResult map[string]interface{}
	json.NewDecoder(insertResp.Body).Decode(&insertResult)
	insertResp.Body.Close()

	documentID := insertResult["id"].(string)
	fmt.Printf("Document inséré avec l'ID: %s\n", documentID)

	// 3. Vérifier que les fichiers WAL existent
	fmt.Println("\n3. Vérifier les fichiers WAL avant commit...")
	walFiles, err := os.ReadDir("data/wal")
	if err != nil {
		log.Fatal("Erreur lecture répertoire WAL:", err)
	}

	fmt.Printf("Fichiers WAL avant commit: %d\n", len(walFiles))
	for _, file := range walFiles {
		fmt.Printf("  - %s\n", file.Name())
	}

	// 4. Valider la transaction
	fmt.Println("\n4. Valider la transaction...")
	commitData := map[string]interface{}{
		"transaction_id": transactionID,
	}

	commitJSON, _ := json.Marshal(commitData)
	commitResp, err := http.Post(
		baseURL+"/api/transaction/commit",
		"application/json",
		bytes.NewBuffer(commitJSON),
	)
	if err != nil {
		log.Fatal("Erreur lors du commit:", err)
	}

	var commitResult map[string]interface{}
	json.NewDecoder(commitResp.Body).Decode(&commitResult)
	commitResp.Body.Close()

	fmt.Printf("Transaction validée: %+v\n", commitResult)

	// 5. Vérifier que les fichiers WAL ont été nettoyés
	fmt.Println("\n5. Vérifier les fichiers WAL après commit...")
	time.Sleep(100 * time.Millisecond) // Petit délai pour s'assurer que le nettoyage est terminé

	walFilesAfter, err := os.ReadDir("data/wal")
	if err != nil {
		log.Fatal("Erreur lecture répertoire WAL:", err)
	}

	fmt.Printf("Fichiers WAL après commit: %d\n", len(walFilesAfter))
	for _, file := range walFilesAfter {
		fmt.Printf("  - %s\n", file.Name())
	}

	if len(walFilesAfter) == 0 {
		fmt.Println("✅ SUCCÈS: Les fichiers WAL ont été correctement nettoyés après le commit")
	} else {
		fmt.Println("❌ ÉCHEC: Les fichiers WAL n'ont pas été nettoyés")
	}

	// 6. Test avec rollback
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TEST AVEC ROLLBACK")
	fmt.Println(strings.Repeat("=", 60))

	// Commencer une nouvelle transaction
	fmt.Println("\n1. Commencer une nouvelle transaction...")
	txResponse2, err := http.Post(baseURL+"/api/transaction/begin", "application/json", nil)
	if err != nil {
		log.Fatal("Erreur lors du début de transaction:", err)
	}

	var txResult2 map[string]interface{}
	json.NewDecoder(txResponse2.Body).Decode(&txResult2)
	txResponse2.Body.Close()

	transactionID2 := txResult2["transaction_id"].(string)
	fmt.Printf("Transaction ID: %s\n", transactionID2)

	// Insérer un document
	fmt.Println("\n2. Insérer un document...")
	insertData2 := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Test WAL Rollback",
			"iban":        222222222,
			"description": "Document pour tester le rollback WAL",
		},
	}

	insertJSON2, _ := json.Marshal(insertData2)
	insertResp2, err := http.Post(
		baseURL+"/api/transaction/"+transactionID2+"/insert",
		"application/json",
		bytes.NewBuffer(insertJSON2),
	)
	if err != nil {
		log.Fatal("Erreur lors de l'insertion:", err)
	}

	var insertResult2 map[string]interface{}
	json.NewDecoder(insertResp2.Body).Decode(&insertResult2)
	insertResp2.Body.Close()

	documentID2 := insertResult2["id"].(string)
	fmt.Printf("Document inséré avec l'ID: %s\n", documentID2)

	// Vérifier les fichiers WAL avant rollback
	fmt.Println("\n3. Vérifier les fichiers WAL avant rollback...")
	walFilesBeforeRollback, err := os.ReadDir("data/wal")
	if err != nil {
		log.Fatal("Erreur lecture répertoire WAL:", err)
	}

	fmt.Printf("Fichiers WAL avant rollback: %d\n", len(walFilesBeforeRollback))

	// Annuler la transaction
	fmt.Println("\n4. Annuler la transaction...")
	rollbackData := map[string]interface{}{
		"transaction_id": transactionID2,
	}

	rollbackJSON, _ := json.Marshal(rollbackData)
	rollbackResp, err := http.Post(
		baseURL+"/api/transaction/rollback",
		"application/json",
		bytes.NewBuffer(rollbackJSON),
	)
	if err != nil {
		log.Fatal("Erreur lors du rollback:", err)
	}

	var rollbackResult map[string]interface{}
	json.NewDecoder(rollbackResp.Body).Decode(&rollbackResult)
	rollbackResp.Body.Close()

	fmt.Printf("Transaction annulée: %+v\n", rollbackResult)

	// Vérifier que les fichiers WAL ont été nettoyés
	fmt.Println("\n5. Vérifier les fichiers WAL après rollback...")
	time.Sleep(100 * time.Millisecond)

	walFilesAfterRollback, err := os.ReadDir("data/wal")
	if err != nil {
		log.Fatal("Erreur lecture répertoire WAL:", err)
	}

	fmt.Printf("Fichiers WAL après rollback: %d\n", len(walFilesAfterRollback))
	for _, file := range walFilesAfterRollback {
		fmt.Printf("  - %s\n", file.Name())
	}

	if len(walFilesAfterRollback) == 0 {
		fmt.Println("✅ SUCCÈS: Les fichiers WAL ont été correctement nettoyés après le rollback")
	} else {
		fmt.Println("❌ ÉCHEC: Les fichiers WAL n'ont pas été nettoyés après le rollback")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TEST TERMINÉ")
	fmt.Println(strings.Repeat("=", 60))
}
