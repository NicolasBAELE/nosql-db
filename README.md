# NoSQL Database

Une base de données NoSQL simple et légère implémentée en Go, avec un serveur HTTP pour l'accès aux données.

## Fonctionnalités

- Stockage de documents JSON
- Collections avec index uniques et non-uniques
- API REST pour les opérations CRUD
- Interface web pour visualiser les données
- Configuration des collections via fichier JSON
- Recherche par champ indexé

## Structure du Projet

```
.
├── cmd/
│   └── server/           # Serveur HTTP principal
├── internal/
│   └── database/         # Implémentation de la base de données
├── config/
│   └── collections.json  # Configuration des collections
└── data/                 # Stockage des documents
```

## Installation

1. Cloner le repository :
```bash
git clone [URL_DU_REPO]
cd nosql-db
```

2. Installer les dépendances :
```bash
go mod download
```

3. Lancer le serveur :
```bash
go run cmd/server/main.go -config config/collections.json
```

## Configuration

Le fichier `config/collections.json` définit les collections et leurs index :

```json
{
  "collections": [
    {
      "name": "books",
      "indexes": [
        {
          "field": "iban",
          "unique": true
        }
      ]
    },
    {
      "name": "users",
      "indexes": [
        {
          "field": "email",
          "unique": true
        },
        {
          "field": "age",
          "unique": false
        }
      ]
    }
  ]
}
```

## API REST

### Collections

- `GET /api/{collectionName}` - Liste tous les documents d'une collection
- `POST /api/{collectionName}` - Crée un nouveau document
- `GET /api/{collectionName}/{id}` - Récupère un document par son ID
- `PUT /api/{collectionName}/{id}` - Met à jour un document
- `DELETE /api/{collectionName}/{id}` - Supprime un document
- `GET /api/{collectionName}/search?field={field}&value={value}` - Recherche des documents par champ

### Exemples de Requêtes

1. Créer un livre :
```bash
curl -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{"title": "Le Petit Prince", "iban": "9782070408504", "description": "Un classique de la littérature"}'
```

2. Récupérer un livre par ID :
```bash
curl http://localhost:8080/api/books/183b1c653bc080b8
```

3. Mettre à jour un livre :
```bash
curl -X PUT http://localhost:8080/api/books/183b1c653bc080b8 \
  -H "Content-Type: application/json" \
  -d '{"title": "Le Petit Prince - Édition Spéciale", "iban": "9782070408504"}'
```

4. Rechercher des livres par IBAN :
```bash
curl "http://localhost:8080/api/books/search?field=iban&value=9782070408504"
```

## Interface Web

L'interface web est accessible à l'adresse `http://localhost:8080/`. Elle permet de :
- Voir toutes les collections
- Afficher les documents de chaque collection
- Visualiser les données au format JSON

## Fonctionnement de la Base de Données

### Stockage

- Les documents sont stockés au format JSON dans le dossier `data/`
- Chaque collection a son propre sous-dossier
- Chaque document est stocké dans un fichier séparé avec son ID comme nom

### Index

- Les index sont stockés en mémoire
- Support des index uniques et non-uniques
- Les index uniques empêchent la duplication de valeurs
- Les index non-uniques permettent une recherche rapide

### Concurrence

- Toutes les opérations sont thread-safe grâce à l'utilisation de mutex
- Les verrous sont appliqués au niveau de la collection

## Dépendances

- Go 1.24.1 ou supérieur
- Aucune dépendance externe

## Licence

MIT 