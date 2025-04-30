# NoSQL Database

Une base de données NoSQL autonome implémentée en Go, avec support des indexs et persistance sur disque.

## Fonctionnalités

- Stockage de documents JSON
- Collections
- Indexs uniques et non-uniques
- Recherche par ID et par index
- Mise à jour et suppression de documents
- Persistance sur disque
- Thread-safe
- Pas de dépendances externes
- Gestion automatique des dossiers

## Installation

1. Clonez le dépôt :
```bash
git clone https://github.com/votre-username/nosql-db.git
cd nosql-db
```

2. Assurez-vous d'avoir Go 1.24.1 ou supérieur installé.

## Structure du projet

```
.
├── cmd/
│   └── main.go          # Exemple d'utilisation
├── internal/
│   └── database/
│       └── storage.go   # Implémentation de la base de données
├── examples/
│   └── library/         # Exemple d'application
├── go.mod
└── README.md
```

## Utilisation

### Création d'une base de données

```go
db, err := database.NewDatabase("data")
if err != nil {
    log.Fatal(err)
}
```

### Création d'une collection

```go
users, err := db.CreateCollection("users")
if err != nil {
    log.Fatal(err)
}
```

### Création d'indexs

```go
// Index unique
err = users.CreateIndex("email", true)

// Index non-unique
err = users.CreateIndex("age", false)
```

### Insertion de documents

```go
doc := database.Document{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}

id, err := users.Insert(doc)
```

### Recherche de documents

```go
// Par ID
doc, err := users.FindByID(id)

// Par index
docs, err := users.FindByIndex("email", "john@example.com")
```

### Mise à jour de documents

```go
updateDoc := database.Document{
    "name":  "John Doe",
    "email": "john.doe@example.com",
    "age":   31,
}

err = users.Update(id, updateDoc)
```

### Suppression de documents

```go
err = users.Delete(id)
```

## Structure des données

### Organisation des dossiers

La base de données crée et gère automatiquement la structure des dossiers :

```
data/                      # Dossier racine de la base de données
  users/                   # Collection "users"
    document1.json         # Document avec ID unique
    document2.json
    ...
  products/               # Collection "products"
    document1.json
    document2.json
    ...
  orders/                 # Collection "orders"
    document1.json
    document2.json
    ...
```

### Format des documents

Les documents sont stockés au format JSON avec une structure flexible :

```json
{
  "_id": "unique_id",
  "field1": "value1",
  "field2": 123,
  "field3": true,
  "field4": {
    "nested": "value"
  },
  "field5": ["array", "of", "values"]
}
```

## Gestion des indexs

### Types d'indexs

- **Index unique** : Garantit l'unicité des valeurs (ex: email, ISBN)
- **Index non-unique** : Permet plusieurs documents avec la même valeur (ex: age, catégorie)

### Avantages des indexs

- Recherches rapides (O(1) au lieu de O(n))
- Contraintes d'unicité automatiques
- Mise à jour automatique lors des modifications

### Limitations

- Les indexs sont stockés en mémoire
- Plus d'indexs = plus de consommation mémoire
- Les modifications sont plus lentes avec beaucoup d'indexs

## Sécurité et concurrence

- Toutes les opérations sont thread-safe grâce à l'utilisation de mutex
- Les fichiers sont verrouillés pendant les opérations d'écriture
- Les indexs sont protégés contre les accès concurrents

## Bonnes pratiques

1. **Organisation des collections**
   - Utilisez des noms de collections significatifs
   - Regroupez les données logiquement
   - Évitez de créer trop de collections

2. **Gestion des indexs**
   - Créez des indexs uniquement sur les champs fréquemment recherchés
   - Utilisez des indexs uniques pour les champs qui doivent être uniques
   - Évitez de créer trop d'indexs sur une même collection

3. **Structure des documents**
   - Utilisez des noms de champs cohérents
   - Évitez les documents trop volumineux
   - Privilégiez une structure plate quand possible

## Exemple d'application

Un exemple d'application de gestion de bibliothèque est disponible dans le dossier `examples/library/`. Il montre comment utiliser la base de données dans un cas concret.

## Limitations

- Les indexs sont stockés en mémoire (limitation de la RAM)
- Pas de support pour les transactions
- Pas de support pour les requêtes complexes
- Pas de support pour les indexs composés

## Contribution

Les contributions sont les bienvenues ! N'hésitez pas à :
1. Fork le projet
2. Créer une branche pour votre fonctionnalité
3. Commiter vos changements
4. Pousser vers la branche
5. Ouvrir une Pull Request

## Licence

MIT 