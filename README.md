# Forum-JS

Forum web collaboratif développé en Go avec SQLite, dans le cadre du projet B1 Informatique.

## Fonctionnalités

### Obligatoires
- **Inscription / Connexion / Déconnexion** — avec hash bcrypt des mots de passe
- **Sessions utilisateur** — via cookies sécurisés (gorilla/sessions)
- **Posts** — création avec titre, contenu, image, catégorie (thème)
- **Commentaires** — création, édition et suppression
- **Like / Dislike** — sur les posts et les commentaires (toggle)
- **Filtrage des posts** — par thème, posts likés, mes posts (combinables)
- **Images** — upload avec redimensionnement (JPEG, PNG, GIF, WebP ≤ 20 Mo)
- **Gestion des erreurs HTTP** — pages 404, 403, 401, 500
- **Docker** — conteneurisation complète

### Bonus
- **OAuth Google** — connexion via compte Google
- **Tests unitaires** — 3 fichiers de test (auth, database, server)

## Structure du projet

```
Forum-JS/
├── Main.go              # Point d'entrée
├── go.mod / go.sum      # Dépendances Go
├── Dockerfile           # Image Docker multi-stage
├── docker-compose.yml   # Orchestration Docker
├── .env                 # Variables d'environnement
├── README.md
├── backend/
│   ├── Server.go        # Routes HTTP, templates, gestion erreurs
│   ├── Database.go      # Connexion SQLite, requêtes CRUD
│   ├── auth.go          # Authentification, OAuth Google
│   ├── session.go       # Sessions et cookies
│   ├── posts.go         # Création / suppression de posts
│   ├── comments.go      # Création / édition de commentaires
│   ├── likes.go         # Like / Dislike toggle
│   ├── filter.go        # Filtrage combiné des posts
│   ├── auth_test.go     # Tests authentification
│   ├── Database_test.go # Tests base de données
│   └── Server_test.go   # Tests routes
├── frontend/
│   ├── html/            # Templates Go (index, login, register, error, 404)
│   ├── css/             # Styles (Index.css, Auth.css, 404.css)
│   └── js/              # Scripts (post.js, oauth.js)
└── database/            # Fichier SQLite (généré automatiquement)
```

## Installation et lancement

### Avec Docker (recommandé)

1. **Prérequis** : Docker et Docker Compose installés

2. **Cloner le dépôt**
   ```bash
   git clone https://github.com/enzoandria1-oss/Forum-JS.git
   cd Forum-JS
   ```

3. **Configurer les variables d'environnement**
   ```bash
   # Copier le fichier .env.example ou éditer .env avec vos clés
   # GOOGLE_CLIENT_ID et GOOGLE_CLIENT_SECRET sont requis pour OAuth
   # SESSION_KEY peut être modifiée pour la production
   ```

4. **Lancer le forum**
   ```bash
   docker-compose up --build
   ```

5. **Accéder au forum**
   ```
   http://localhost:8082
   ```

### Sans Docker

1. **Prérequis** : Go 1.25+ installé

2. **Cloner et lancer**
   ```bash
   git clone https://github.com/enzoandria1-oss/Forum-JS.git
   cd Forum-JS
   go mod download
   go run Main.go
   ```

3. **Accéder au forum**
   ```
   http://localhost:8082
   ```

## Utilisation

### Pages
| Route | Accès | Description |
|---|---|---|
| `/` | Public | Accueil avec liste des posts et filtres |
| `/login` | Public | Connexion (email + Google) |
| `/register` | Public | Inscription |
| `/auth/logout` | Connecté | Déconnexion |

### API Endpoints
| Route | Méthode | Auth | Description |
|---|---|---|---|
| `/db/login` | POST | Non | Connexion (form/JSON) |
| `/db/register` | POST | Non | Inscription (form/JSON) |
| `/db/create_post` | POST | Oui | Créer un post |
| `/db/delete_post` | POST/DELETE | Oui | Supprimer son post |
| `/db/create_comment` | POST | Oui | Commenter un post |
| `/db/edit_comment` | POST | Oui | Éditer son commentaire |
| `/db/delete_comment` | POST/DELETE | Oui | Supprimer son commentaire |
| `/db/toggle_like` | GET/POST | Oui | Like/unlike un post |
| `/db/toggle_dislike` | GET/POST | Oui | Dislike/undislike un post |
| `/db/toggle_comment_like` | GET/POST | Oui | Like/unlike commentaire |
| `/db/toggle_comment_dislike` | GET/POST | Oui | Dislike/undislike commentaire |
| `/db/posts` | GET | Non | Liste des posts (JSON) |
| `/db/comments` | GET | Non | Commentaires d'un post (JSON) |
| `/api/me` | GET | Oui | Infos utilisateur connecté |

### Filtres
Les filtres sont accessibles depuis la page d'accueil et sont combinables :

- **Par thème** — sélectionner un thème dans la liste déroulante
- **Posts likés** — cocher "Posts likés"
- **Mes posts** — cocher "My posts"

## Technologies

- **Backend** : Go (standard library, sans framework)
- **Base de données** : SQLite (modernc.org/sqlite)
- **Sessions** : gorilla/sessions
- **Authentification** : bcrypt (mots de passe), OAuth 2.0 (Google)
- **Images** : imaging, uuid
- **Frontend** : HTML, CSS, JavaScript (vanilla)
- **Conteneurisation** : Docker, Docker Compose