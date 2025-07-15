package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const baseURL = "http://localhost:8081"

func main() {
	fmt.Println("🧪 Test de l'interface web avec transactions")
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

	// 2. Insérer un document via l'API transaction
	fmt.Println("\n2. Insérer un document...")
	insertData := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Livre via Interface Web",
			"iban":        987654321,
			"description": "Créé via l'interface web avec transaction",
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

	// 3. Vérifier que le document existe dans la transaction
	fmt.Println("\n3. Vérifier le document dans la transaction...")
	getResp, err := http.Get(baseURL + "/api/books/" + documentID)
	if err != nil {
		log.Fatal("Erreur lors de la récupération:", err)
	}

	if getResp.StatusCode == 200 {
		fmt.Println("✅ Document accessible pendant la transaction")
	} else {
		fmt.Println("❌ Document non accessible pendant la transaction")
	}
	getResp.Body.Close()

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

	// 5. Vérifier que le document existe maintenant de manière permanente
	fmt.Println("\n5. Vérifier que le document existe maintenant...")
	getResp2, err := http.Get(baseURL + "/api/books/" + documentID)
	if err != nil {
		log.Fatal("Erreur lors de la récupération:", err)
	}

	if getResp2.StatusCode == 200 {
		fmt.Println("✅ Document accessible après commit")
	} else {
		fmt.Println("❌ Document non accessible après commit")
	}
	getResp2.Body.Close()

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

	// Insérer un document qui sera annulé
	fmt.Println("\n2. Insérer un document qui sera annulé...")
	insertData2 := map[string]interface{}{
		"collection": "books",
		"document": map[string]interface{}{
			"title":       "Livre à Annuler",
			"iban":        555555555,
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

	// Annuler la transaction
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

	// Vérifier que le document n'existe plus
	fmt.Println("\n4. Vérifier que le document n'existe plus...")
	getResp3, err := http.Get(baseURL + "/api/books/" + documentID2)
	if err != nil {
		log.Fatal("Erreur lors de la récupération:", err)
	}

	if getResp3.StatusCode == 404 {
		fmt.Println("✅ Document correctement supprimé par le rollback")
	} else {
		fmt.Println("❌ Document existe encore après le rollback")
	}
	getResp3.Body.Close()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TEST TERMINÉ")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n🌐 Interface web disponible sur: http://localhost:8081")
	fmt.Println("   - Commencez une transaction")
	fmt.Println("   - Ajoutez, modifiez ou supprimez des documents")
	fmt.Println("   - Validez ou annulez la transaction")
}
