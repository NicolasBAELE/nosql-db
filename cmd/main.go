package main

import (
	"fmt"
	"log"
	"nosql-db/internal/orm/models"
)

// Exemple d'utilisation de l'ORM dans différents scénarios
func main() {
	// Initialisation du gestionnaire de base de données
	fmt.Println("=== Exemple d'utilisation de l'ORM ===")
	dbManager, err := models.NewDatabaseManager()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du gestionnaire de base de données: %v", err)
	}

	// Démarrer le gestionnaire
	dbManager.Start()
	log.Println("Gestionnaire de base de données démarré")

	// Récupérer les modèles
	userModel, exists := dbManager.GetModel("users")
	if !exists {
		log.Fatal("Modèle User non trouvé")
	}
	userCollection := userModel.GetCollection().(*models.UserCollection)

	productModel, exists := dbManager.GetModel("products")
	if !exists {
		log.Fatal("Modèle Product non trouvé")
	}
	productCollection := productModel.GetCollection().(*models.ProductCollection)

	// 1. Exemple de création et gestion d'utilisateurs
	demonstrateUserOperations(userCollection)

	// 2. Exemple de création et gestion de produits
	demonstrateProductOperations(productCollection)

	// 3. Exemple de recherche avancée
	demonstrateAdvancedSearch(userCollection, productCollection)

	// 4. Exemple de gestion des erreurs
	demonstrateErrorHandling(userCollection)

	// 5. Exemple de statistiques et métriques
	demonstrateStatistics(userCollection, productCollection)

	// Attendre indéfiniment (dans une vraie application, vous auriez un serveur HTTP ici)
	fmt.Println("\n=== L'application continue de fonctionner ===")
	fmt.Println("Appuyez sur Ctrl+C pour arrêter l'application")
	select {}
}

// demonstrateUserOperations montre les opérations CRUD sur les utilisateurs
func demonstrateUserOperations(userCollection *models.UserCollection) {
	fmt.Println("\n=== Opérations sur les utilisateurs ===")

	// Création d'utilisateurs
	user1 := models.User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		IsActive: true,
		Metadata: map[string]interface{}{
			"role": "admin",
			"tags": []string{"golang", "backend"},
		},
	}

	user2 := models.User{
		Name:     "Jane Smith",
		Email:    "jane@example.com",
		Age:      25,
		IsActive: true,
		Metadata: map[string]interface{}{
			"role": "user",
			"tags": []string{"frontend", "design"},
		},
	}

	// Insertion
	fmt.Println("\n--- Insertion d'utilisateurs ---")
	id1, err := userCollection.Insert(user1)
	if err != nil {
		log.Printf("Erreur lors de l'insertion de l'utilisateur 1: %v", err)
	} else {
		fmt.Printf("Utilisateur 1 inséré avec l'ID: %s\n", id1)
	}

	id2, err := userCollection.Insert(user2)
	if err != nil {
		log.Printf("Erreur lors de l'insertion de l'utilisateur 2: %v", err)
	} else {
		fmt.Printf("Utilisateur 2 inséré avec l'ID: %s\n", id2)
	}

	// Lecture
	fmt.Println("\n--- Lecture d'utilisateurs ---")
	user, err := userCollection.Get(id1)
	if err != nil {
		log.Printf("Erreur lors de la récupération de l'utilisateur: %v", err)
	} else {
		fmt.Printf("Utilisateur récupéré: %+v\n", user)
	}

	// Mise à jour
	fmt.Println("\n--- Mise à jour d'utilisateur ---")
	user.Age = 31
	if err := userCollection.Update(id1, user); err != nil {
		log.Printf("Erreur lors de la mise à jour de l'utilisateur: %v", err)
	} else {
		fmt.Println("Utilisateur mis à jour avec succès")
	}

	// Suppression
	fmt.Println("\n--- Suppression d'utilisateur ---")
	if err := userCollection.Delete(id2); err != nil {
		log.Printf("Erreur lors de la suppression de l'utilisateur: %v", err)
	} else {
		fmt.Println("Utilisateur supprimé avec succès")
	}
}

// demonstrateProductOperations montre les opérations CRUD sur les produits
func demonstrateProductOperations(productCollection *models.ProductCollection) {
	fmt.Println("\n=== Opérations sur les produits ===")

	// Création de produits
	product1 := models.Product{
		Name:        "Laptop",
		Price:       999.99,
		Description: "Un ordinateur portable puissant",
		InStock:     true,
	}

	product2 := models.Product{
		Name:        "Smartphone",
		Price:       499.99,
		Description: "Un smartphone dernier cri",
		InStock:     true,
	}

	// Insertion
	fmt.Println("\n--- Insertion de produits ---")
	pid1, err := productCollection.Insert(product1)
	if err != nil {
		log.Printf("Erreur lors de l'insertion du produit 1: %v", err)
	} else {
		fmt.Printf("Produit 1 inséré avec l'ID: %s\n", pid1)
	}

	pid2, err := productCollection.Insert(product2)
	if err != nil {
		log.Printf("Erreur lors de l'insertion du produit 2: %v", err)
	} else {
		fmt.Printf("Produit 2 inséré avec l'ID: %s\n", pid2)
	}

	// Lecture
	fmt.Println("\n--- Lecture de produits ---")
	product, err := productCollection.Get(pid1)
	if err != nil {
		log.Printf("Erreur lors de la récupération du produit: %v", err)
	} else {
		fmt.Printf("Produit récupéré: %+v\n", product)
	}

	// Mise à jour
	fmt.Println("\n--- Mise à jour de produit ---")
	product.Price = 899.99
	if err := productCollection.Update(pid1, product); err != nil {
		log.Printf("Erreur lors de la mise à jour du produit: %v", err)
	} else {
		fmt.Println("Produit mis à jour avec succès")
	}

	// Suppression
	fmt.Println("\n--- Suppression de produit ---")
	if err := productCollection.Delete(pid2); err != nil {
		log.Printf("Erreur lors de la suppression du produit: %v", err)
	} else {
		fmt.Println("Produit supprimé avec succès")
	}
}

// demonstrateAdvancedSearch montre les fonctionnalités de recherche avancée
func demonstrateAdvancedSearch(userCollection *models.UserCollection, productCollection *models.ProductCollection) {
	fmt.Println("\n=== Recherches avancées ===")

	// Recherche d'utilisateurs par email
	fmt.Println("\n--- Recherche d'utilisateur par email ---")
	userByEmail, err := userCollection.FindByEmail("john@example.com")
	if err != nil {
		log.Printf("Erreur lors de la recherche d'utilisateur par email: %v", err)
	} else {
		fmt.Printf("Utilisateur trouvé par email: %+v\n", userByEmail)
	}

	// Recherche d'utilisateurs actifs
	fmt.Println("\n--- Recherche d'utilisateurs actifs ---")
	activeUsers, err := userCollection.FindActive()
	if err != nil {
		log.Printf("Erreur lors de la recherche d'utilisateurs actifs: %v", err)
	} else {
		fmt.Printf("Nombre d'utilisateurs actifs: %d\n", len(activeUsers))
		for i, user := range activeUsers {
			fmt.Printf("  %d. %s (%s)\n", i+1, user.Name, user.Email)
		}
	}

	// Recherche de produits par prix
	fmt.Println("\n--- Recherche de produits par prix ---")
	expensiveProducts, err := productCollection.FindByPrice(999.99)
	if err != nil {
		log.Printf("Erreur lors de la recherche de produits par prix: %v", err)
	} else {
		fmt.Printf("Nombre de produits à 999.99€: %d\n", len(expensiveProducts))
		for i, product := range expensiveProducts {
			fmt.Printf("  %d. %s (%.2f€)\n", i+1, product.Name, product.Price)
		}
	}

	// Recherche de produits en stock
	fmt.Println("\n--- Recherche de produits en stock ---")
	availableProducts, err := productCollection.FindInStock()
	if err != nil {
		log.Printf("Erreur lors de la recherche de produits en stock: %v", err)
	} else {
		fmt.Printf("Nombre de produits en stock: %d\n", len(availableProducts))
		for i, product := range availableProducts {
			fmt.Printf("  %d. %s (%.2f€)\n", i+1, product.Name, product.Price)
		}
	}
}

// demonstrateErrorHandling montre la gestion des erreurs
func demonstrateErrorHandling(userCollection *models.UserCollection) {
	fmt.Println("\n=== Gestion des erreurs ===")

	// Tentative de récupération d'un utilisateur inexistant
	fmt.Println("\n--- Tentative de récupération d'un utilisateur inexistant ---")
	_, err := userCollection.Get("id_inexistant")
	if err != nil {
		fmt.Printf("Erreur attendue: %v\n", err)
	} else {
		fmt.Println("Erreur: l'utilisateur aurait dû ne pas être trouvé")
	}

	// Tentative d'insertion d'un utilisateur avec un email déjà utilisé
	fmt.Println("\n--- Tentative d'insertion d'un utilisateur avec un email déjà utilisé ---")
	user := models.User{
		Name:     "Duplicate User",
		Email:    "john@example.com", // Email déjà utilisé
		Age:      35,
		IsActive: true,
		Metadata: map[string]interface{}{
			"role": "user",
		},
	}
	_, err = userCollection.Insert(user)
	if err != nil {
		fmt.Printf("Erreur attendue: %v\n", err)
	} else {
		fmt.Println("Erreur: l'insertion aurait dû échouer à cause de l'email dupliqué")
	}
}

// demonstrateStatistics montre les statistiques et métriques
func demonstrateStatistics(userCollection *models.UserCollection, productCollection *models.ProductCollection) {
	fmt.Println("\n=== Statistiques et métriques ===")

	// Nombre d'utilisateurs
	userCount := userCollection.Count()
	fmt.Printf("Nombre total d'utilisateurs: %d\n", userCount)

	// Nombre de produits
	productCount := productCollection.Count()
	fmt.Printf("Nombre total de produits: %d\n", productCount)
}
