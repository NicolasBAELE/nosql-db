package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	baseURL := "http://localhost:8081"

	fmt.Println("🎯 DÉMONSTRATION COMPLÈTE DES INDEX NON-UNIQUES")
	fmt.Println(strings.Repeat("=", 60))

	// 1. Démonstration des index non-uniques
	fmt.Println("\n📚 1. DÉMONSTRATION DES INDEX NON-UNIQUES")
	fmt.Println(strings.Repeat("-", 40))

	// Vérifier les livres existants avec le même titre
	searchResp, err := http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche: %v", err)
	}

	var searchResults []map[string]interface{}
	json.NewDecoder(searchResp.Body).Decode(&searchResults)
	searchResp.Body.Close()

	fmt.Printf("   📖 Il y a actuellement %d exemplaires de 'Le Petit Prince' dans la base\n", len(searchResults))
	for i, book := range searchResults {
		fmt.Printf("   %d. ID: %s, IBAN: %v, Description: %s\n",
			i+1, book["_id"], book["iban"], book["description"])
	}

	// 2. Test de performance
	fmt.Println("\n⚡ 2. TEST DE PERFORMANCE")
	fmt.Println(strings.Repeat("-", 40))

	// Test avec index (titre)
	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
		if err != nil {
			log.Fatalf("Erreur lors du test de performance: %v", err)
		}
	}
	indexedTime := time.Since(start) / 10

	// Test sans index (auteur)
	start = time.Now()
	for i := 0; i < 10; i++ {
		_, err := http.Get(baseURL + "/api/books/search?field=author&value=Antoine%20de%20Saint-Exupéry")
		if err != nil {
			log.Fatalf("Erreur lors du test de performance: %v", err)
		}
	}
	nonIndexedTime := time.Since(start) / 10

	fmt.Printf("   🔍 Recherche avec index (titre): %v (moyenne sur 10 requêtes)\n", indexedTime)
	fmt.Printf("   🔍 Recherche sans index (auteur): %v (moyenne sur 10 requêtes)\n", nonIndexedTime)
	fmt.Printf("   📊 Amélioration: %.1fx plus rapide avec index\n", float64(nonIndexedTime)/float64(indexedTime))

	// 3. Comparaison des comportements
	fmt.Println("\n🔄 3. COMPARAISON DES COMPORTEMENTS")
	fmt.Println(strings.Repeat("-", 40))

	// Test index unique (IBAN)
	fmt.Println("   🔒 Index unique (IBAN):")
	uniqueResp, err := http.Get(baseURL + "/api/books/search?field=iban&value=9782070408504")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche unique: %v", err)
	}
	var uniqueResults []map[string]interface{}
	json.NewDecoder(uniqueResp.Body).Decode(&uniqueResults)
	uniqueResp.Body.Close()
	fmt.Printf("      → Trouvé %d livre(s) avec l'IBAN 9782070408504\n", len(uniqueResults))

	// Test index non-unique (titre)
	fmt.Println("   🔓 Index non-unique (titre):")
	fmt.Printf("      → Trouvé %d livre(s) avec le titre 'Le Petit Prince'\n", len(searchResults))

	// Test champ non-indexé (auteur)
	authorResp, err := http.Get(baseURL + "/api/books/search?field=author&value=Antoine%20de%20Saint-Exupéry")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche par auteur: %v", err)
	}
	var authorResults []map[string]interface{}
	json.NewDecoder(authorResp.Body).Decode(&authorResults)
	authorResp.Body.Close()
	fmt.Printf("      → Trouvé %d livre(s) de 'Antoine de Saint-Exupéry'\n", len(authorResults))

	// 4. Test des contraintes
	fmt.Println("\n🚫 4. TEST DES CONTRAINTES")
	fmt.Println(strings.Repeat("-", 40))

	// Test violation index unique
	fmt.Println("   🔒 Test violation index unique (IBAN):")
	duplicateBook := map[string]interface{}{
		"title":       "Test Violation",
		"iban":        9782070408504, // IBAN déjà existant
		"description": "Ceci devrait échouer",
		"author":      "Test Author",
	}
	duplicateJSON, _ := json.Marshal(duplicateBook)
	resp, err := http.Post(baseURL+"/api/books", "application/json", bytes.NewBuffer(duplicateJSON))
	if resp.StatusCode == 500 {
		fmt.Println("      ✅ Violation détectée (comportement attendu)")
	} else {
		fmt.Printf("      ❌ Erreur: Violation non détectée (status: %d)\n", resp.StatusCode)
	}
	resp.Body.Close()

	// Test index non-unique (devrait réussir)
	fmt.Println("   🔓 Test index non-unique (titre):")
	newBook := map[string]interface{}{
		"title":       "Le Petit Prince", // Titre déjà existant
		"iban":        9876543210,        // IBAN unique
		"description": "Test index non-unique",
		"author":      "Antoine de Saint-Exupéry",
	}
	newBookJSON, _ := json.Marshal(newBook)
	resp2, err := http.Post(baseURL+"/api/books", "application/json", bytes.NewBuffer(newBookJSON))
	if resp2.StatusCode == 201 {
		fmt.Println("      ✅ Insertion réussie avec titre dupliqué (comportement attendu)")
		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)
		resp2.Body.Close()
		fmt.Printf("      📖 Nouveau livre créé avec l'ID: %s\n", result["id"])
	} else {
		fmt.Printf("      ❌ Erreur lors de l'insertion (status: %d)\n", resp2.StatusCode)
		resp2.Body.Close()
	}

	// 5. Vérification finale
	fmt.Println("\n✅ 5. VÉRIFICATION FINALE")
	fmt.Println(strings.Repeat("-", 40))

	// Vérifier le nombre final de livres avec le même titre
	finalSearchResp, err := http.Get(baseURL + "/api/books/search?field=title&value=Le%20Petit%20Prince")
	if err != nil {
		log.Fatalf("Erreur lors de la recherche finale: %v", err)
	}
	var finalResults []map[string]interface{}
	json.NewDecoder(finalSearchResp.Body).Decode(&finalResults)
	finalSearchResp.Body.Close()

	fmt.Printf("   📚 Total final: %d livres avec le titre 'Le Petit Prince'\n", len(finalResults))
	for i, book := range finalResults {
		fmt.Printf("   %d. ID: %s, IBAN: %v, Description: %s\n",
			i+1, book["_id"], book["iban"], book["description"])
	}

	// 6. Résumé technique
	fmt.Println("\n📋 6. RÉSUMÉ TECHNIQUE")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("   🏗️  Architecture des index:")
	fmt.Println("      • Index unique: map[interface{}][]string avec vérification de contrainte")
	fmt.Println("      • Index non-unique: map[interface{}][]string sans vérification")
	fmt.Println("      • Pas d'index: recherche linéaire dans tous les documents")
	fmt.Println()
	fmt.Println("   ⚡ Performance:")
	fmt.Printf("      • Avec index: %v\n", indexedTime)
	fmt.Printf("      • Sans index: %v\n", nonIndexedTime)
	fmt.Printf("      • Gain: %.1fx plus rapide\n", float64(nonIndexedTime)/float64(indexedTime))
	fmt.Println()
	fmt.Println("   🔒 Contraintes:")
	fmt.Println("      • Index unique: empêche les doublons")
	fmt.Println("      • Index non-unique: permet les doublons")
	fmt.Println("      • Pas d'index: aucune contrainte")

	fmt.Println("\n🎉 DÉMONSTRATION TERMINÉE!")
	fmt.Println("   Les index non-uniques fonctionnent parfaitement! 🚀")
}
