# ── Build ──
# Image Go compiler le projet
FROM golang:1.25-alpine AS builder

# Dossier de travail dans le conteneur
WORKDIR /app

# Copie les fichiers de dépendances
COPY go.mod go.sum ./

# Télécharge les dépendances
RUN go mod download

# Copie tout le code
COPY . .

# Compile le binaire Go pour Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ── Production ──
# Image légère sans les outils Go
FROM alpine:3.24

RUN apk add --no-cache wget

WORKDIR /app

# Copie uniquement le binaire compilé depuis le stage builder
COPY --from=builder /app/main .

# Copie les fichiers statiques du frontend
COPY --from=builder /app/frontend ./frontend

# Expose le port
EXPOSE 8082

# Active le mode production
ENV ENV=production

# Verifie si le site fonctionne bien
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget --spider -q  http://localhost:8082 || exit 1

# Commande de démarrage
ENTRYPOINT ["./main"]