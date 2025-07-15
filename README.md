# NoSQL Database

Une base de données NoSQL simple et légère implémentée en Go, avec un serveur HTTP pour l'accès aux données.

## Fonctionnalités

- Stockage de documents JSON
- Collections avec index uniques et non-uniques
- API REST pour les opérations CRUD
- Interface web pour visualiser les données
- Configuration des collections via fichier JSON
- Recherche par champ indexé
- **Système de transactions ACID avec WAL (Write-Ahead Logging)**

## Structure du Projet

```
.
├── cmd/
│   └── server/           # Serveur HTTP principal
├── internal/
│   └── database/         # Implémentation de la base de données
│       ├── storage.go    # Stockage et opérations CRUD
│       └── transaction.go # Système de transactions
├── config/
│   └── collections.json  # Configuration des collections
├── examples/
│   └── transaction_example.go # Exemples d'utilisation des transactions
└── data/                 # Stockage des documents
    └── wal/              # Write-Ahead Log pour les transactions
```

## Installation

1. Cloner le repository :
```bash
git clone https://github.com/NicolasBAELE/nosql-db.git
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

## API Transactions

### Gestion des Transactions

- `POST /api/transaction/begin` - Commence une nouvelle transaction
- `POST /api/transaction/commit` - Valide une transaction
- `POST /api/transaction/rollback` - Annule une transaction

### Opérations dans une Transaction

- `POST /api/transaction/{transactionID}/insert` - Insère un document dans une transaction
- `PUT /api/transaction/{transactionID}/update` - Met à jour un document dans une transaction
- `DELETE /api/transaction/{transactionID}/delete` - Supprime un document dans une transaction

### Exemples de Transactions

1. Commencer une transaction :
```bash
curl -X POST http://localhost:8080/api/transaction/begin
# Réponse: {"transaction_id": "tx_1234567890", "status": "active"}
```

2. Insérer un document dans la transaction :
```bash
curl -X POST http://localhost:8080/api/transaction/tx_1234567890/insert \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "books",
    "document": {
      "title": "Livre Transactionnel",
      "iban": 123456789,
      "description": "Créé dans une transaction"
    }
  }'
```

3. Valider la transaction :
```bash
curl -X POST http://localhost:8080/api/transaction/commit \
  -H "Content-Type: application/json" \
  -d '{"transaction_id": "tx_1234567890"}'
```

4. Annuler une transaction :
```bash
curl -X POST http://localhost:8080/api/transaction/rollback \
  -H "Content-Type: application/json" \
  -d '{"transaction_id": "tx_1234567890"}'
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

### Transactions

- **Système de transactions ACID** avec support complet des propriétés :
  - **Atomicité** : Toutes les opérations d'une transaction sont validées ou annulées ensemble
  - **Cohérence** : La base de données reste dans un état cohérent
  - **Isolation** : Les transactions sont isolées les unes des autres
  - **Durabilité** : Les transactions validées sont persistantes

- **Write-Ahead Logging (WAL)** : Toutes les opérations sont d'abord écrites dans un log avant d'être appliquées
- **Recovery automatique** : Récupération des transactions non terminées au redémarrage
- **API REST complète** pour la gestion des transactions

## Dépendances

- Go 1.24.1 ou supérieur
- Aucune dépendance externe

## Licence

MIT 
