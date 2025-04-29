# NoSQL Database with ORM

Une base de données NoSQL légère avec un ORM intégré et une interface en ligne de commande, écrite en Go.

## Table des matières

- [Installation](#installation)
- [Fonctionnalités](#fonctionnalités)
- [Utilisation comme bibliothèque](#utilisation-comme-bibliothèque)
  - [Initialisation](#initialisation)
  - [Gestion des collections](#gestion-des-collections)
  - [Opérations CRUD](#opérations-crud)
  - [Recherche](#recherche)
- [Interface en ligne de commande (CLI)](#interface-en-ligne-de-commande-cli)
  - [Démarrage](#démarrage)
  - [Création de collections](#création-de-collections)
  - [Manipulation de collections](#manipulation-de-collections)
- [Exemples](#exemples)
- [Documentation de l'API](#documentation-de-lapi)
  - [Database](#database)
  - [Collection](#collection)
  - [Field](#field)
- [Tests](#tests)
- [Contribution](#contribution)
- [Licence](#licence)

## Installation

### Prérequis

- Go 1.16 ou supérieur

### Installation via Go

```bash
go get github.com/NicolasBAELE/nosql-db
```

### Installation depuis les sources

```bash
git clone https://github.com/NicolasBAELE/nosql-db.git
cd nosql-db
go mod download
```

## Fonctionnalités

### Gestion des collections
- Création de collections dynamiques
- Définition de schémas flexibles
- Support de différents types de champs (string, number, boolean)

### Opérations CRUD
- Insertion de documents
- Lecture de documents
- Mise à jour de documents
- Suppression de documents

### Recherche
- Recherche par champ
- Récupération de tous les documents
- Comptage des documents

### Interface en ligne de commande
- Création interactive de collections
- Manipulation des documents via CLI
- Interface utilisateur intuitive

## Utilisation comme bibliothèque

### Initialisation

```go
package main

import (
    "log"
    "nosql-db/pkg/nosql"
)

func main() {
    // Créer une nouvelle instance de la base de données
    db, err := nosql.NewDatabase()
    if err != nil {
        log.Fatal(err)
    }

    // Démarrer la base de données
    db.Start()
    defer db.Stop()

    // Votre code ici...
}
```

### Gestion des collections

```go
// Créer une nouvelle collection
collection, err := db.CreateCollection("ma_collection")
if err != nil {
    log.Fatal(err)
}

// Ajouter des champs à la collection
collection.AddField("nom", nosql.FieldTypeString)
collection.AddField("age", nosql.FieldTypeNumber)
collection.AddField("actif", nosql.FieldTypeBoolean)

// Récupérer une collection existante
collection, exists := db.GetCollection("ma_collection")
if !exists {
    log.Fatal("Collection non trouvée")
}

// Lister toutes les collections
collections := db.ListCollections()
for _, name := range collections {
    fmt.Println(name)
}
```

### Opérations CRUD

```go
// Insérer un document
doc := map[string]interface{}{
    "nom":    "John Doe",
    "age":    30,
    "actif":  true,
}

id, err := collection.Insert(doc)
if err != nil {
    log.Fatal(err)
}

// Récupérer un document
retrieved, err := collection.Get(id)
if err != nil {
    log.Fatal(err)
}

// Mettre à jour un document
update := map[string]interface{}{
    "age": 31,
}

err = collection.Update(id, update)
if err != nil {
    log.Fatal(err)
}

// Supprimer un document
err = collection.Delete(id)
if err != nil {
    log.Fatal(err)
}
```

### Recherche

```go
// Rechercher des documents par champ
results, err := collection.FindByField("age", 30)
if err != nil {
    log.Fatal(err)
}

// Récupérer tous les documents
allDocs, err := collection.GetAll()
if err != nil {
    log.Fatal(err)
}

// Compter les documents
count := collection.Count()
fmt.Printf("Nombre de documents: %d\n", count)
```

## Interface en ligne de commande (CLI)

La CLI permet d'interagir avec la base de données sans avoir à écrire de code.

### Démarrage

```bash
go run examples/cli/main.go
```

### Création de collections

1. Lancez la CLI
2. Choisissez l'option 1 pour créer une nouvelle collection
3. Entrez le nom de la collection
4. Ajoutez des champs en spécifiant leur nom et leur type
5. Tapez "fin" pour terminer l'ajout de champs

### Manipulation de collections

1. Lancez la CLI
2. Choisissez l'option 3 pour manipuler une collection
3. Sélectionnez la collection à manipuler
4. Choisissez l'opération à effectuer :
   - Ajouter un document
   - Lister les documents
   - Rechercher des documents
   - Mettre à jour un document
   - Supprimer un document

## Exemples

Consultez le dossier `examples` pour des exemples d'utilisation :

- `examples/basic` : Utilisation basique de la bibliothèque
- `examples/cli` : Utilisation de l'interface en ligne de commande

## Documentation de l'API

### Database

```go
type Database struct {
    // ...
}

// NewDatabase crée une nouvelle instance de la base de données
func NewDatabase() (*Database, error)

// Start démarre la base de données
func (db *Database) Start()

// Stop arrête la base de données
func (db *Database) Stop()

// CreateCollection crée une nouvelle collection
func (db *Database) CreateCollection(name string) (*Collection, error)

// GetCollection récupère une collection existante
func (db *Database) GetCollection(name string) (*Collection, bool)

// ListCollections retourne la liste des collections disponibles
func (db *Database) ListCollections() []string
```

### Collection

```go
type Collection struct {
    // ...
}

// AddField ajoute un nouveau champ à la collection
func (c *Collection) AddField(name string, fieldType FieldType)

// GetFields retourne tous les champs de la collection
func (c *Collection) GetFields() []Field

// Insert insère un nouveau document dans la collection
func (c *Collection) Insert(document map[string]interface{}) (string, error)

// Get récupère un document par son ID
func (c *Collection) Get(id string) (map[string]interface{}, error)

// Update met à jour un document existant
func (c *Collection) Update(id string, document map[string]interface{}) error

// Delete supprime un document
func (c *Collection) Delete(id string) error

// FindByField recherche des documents par un champ spécifique
func (c *Collection) FindByField(field string, value interface{}) ([]map[string]interface{}, error)

// GetAll retourne tous les documents de la collection
func (c *Collection) GetAll() ([]map[string]interface{}, error)

// Count retourne le nombre de documents dans la collection
func (c *Collection) Count() int
```

### Field

```go
type FieldType string

const (
    FieldTypeString  FieldType = "string"
    FieldTypeNumber  FieldType = "number"
    FieldTypeBoolean FieldType = "boolean"
)

type Field struct {
    Name string
    Type FieldType
}
```

## Tests

Le projet inclut des tests complets pour toutes les fonctionnalités. Pour exécuter les tests :

```bash
go test ./pkg/nosql/... -v
```

Les tests couvrent :
- Création et gestion de la base de données
- Opérations CRUD sur les collections
- Validation des types de données
- Recherche et filtrage
- Gestion des erreurs

## Contribution

Les contributions sont les bienvenues ! N'hésitez pas à :

1. Fork le projet
2. Créer une branche pour votre fonctionnalité
3. Commiter vos changements
4. Pousser vers la branche
5. Ouvrir une Pull Request

## Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails. 