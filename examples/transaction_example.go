package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Exemple d'utilisation des transactions via l'API REST

func main() {
	baseURL := "http://localhost:8081"

	// 1. Commencer une transaction
	fmt.Println("1. Commencer une transaction...")
	txResponse, err := http.Post(baseURL+"/api/transaction/begin", "application/json", nil)
	if err != nil {
		log.Fatal("Erreur lors du début de transaction:", err)
	}

	var txResult map[string]interface{}
	json.NewDecoder(txResponse.Body).Decode(&txResult)
	txResponse.Body.Close()

	transactionID := txResult["transaction_id"].(string)
	fmt.Printf("Transaction ID: %s\n", transactionID)

	// 2. Insérer un document dans la transaction
	fmt.Println("\n2. Insérer un document dans la transaction...")
	insertData := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Livre Transactionnel",
			"iban":        123456789,
			"description": "Un livre créé dans une transaction",
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

	// 3. Mettre à jour le document dans la transaction
	fmt.Println("\n3. Mettre à jour le document dans la transaction...")
	updateData := map[string]interface{}{
		"collection":  "books",
		"document_id": documentID,
		"updates": map[string]interface{}{
			"description": "Description mise à jour dans la transaction",
			"author":      "Auteur Transactionnel",
		},
	}

	updateJSON, _ := json.Marshal(updateData)
	updateReq, _ := http.NewRequest(
		"PUT",
		baseURL+"/api/transaction/"+transactionID+"/update",
		bytes.NewBuffer(updateJSON),
	)
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		log.Fatal("Erreur lors de la mise à jour:", err)
	}

	var updateResult map[string]interface{}
	json.NewDecoder(updateResp.Body).Decode(&updateResult)
	updateResp.Body.Close()

	fmt.Printf("Document mis à jour: %+v\n", updateResult)

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

	// 5. Vérifier que le document existe maintenant
	fmt.Println("\n5. Vérifier que le document existe...")
	getResp, err := http.Get(baseURL + "/api/books/" + documentID)
	if err != nil {
		log.Fatal("Erreur lors de la récupération:", err)
	}

	var finalDoc map[string]interface{}
	json.NewDecoder(getResp.Body).Decode(&finalDoc)
	getResp.Body.Close()

	fmt.Printf("Document final: %+v\n", finalDoc)

	// Exemple avec rollback
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("EXEMPLE AVEC ROLLBACK")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Commencer une nouvelle transaction
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

	// 2. Insérer un document qui sera annulé
	fmt.Println("\n2. Insérer un document qui sera annulé...")
	insertData2 := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Livre à Annuler",
			"iban":        999999999,
			"description": "Ce document sera supprimé par le rollback",
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

	// 3. Annuler la transaction
	fmt.Println("\n3. Annuler la transaction...")
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

	// 4. Vérifier que le document n'existe plus
	fmt.Println("\n4. Vérifier que le document n'existe plus...")
	getResp2, err := http.Get(baseURL + "/api/books/" + documentID2)
	if err != nil {
		log.Fatal("Erreur lors de la récupération:", err)
	}

	if getResp2.StatusCode == 404 {
		fmt.Println("✅ Document correctement supprimé par le rollback")
	} else {
		fmt.Println("❌ Document existe encore après le rollback")
	}
	getResp2.Body.Close()

	fmt.Println("\n🎉 Exemples de transactions terminés!")
}
