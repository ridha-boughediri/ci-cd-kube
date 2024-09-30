# Étape de build

# Étape 1 : Build de l'application Go
FROM golang:1.20-alpine AS builder

# Répertoire de travail dans le conteneur
WORKDIR /app

# Copier uniquement les fichiers nécessaires pour éviter de copier trop de fichiers
COPY go.mod go.sum ./ 
RUN go mod download

# Copier le reste des fichiers du projet
COPY . .

# Exécuter les tests avant la compilation de l'application
RUN go test ./cmd/... -v

# Compiler l'application Go à partir du répertoire cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd

# Étape finale

# Étape 2 : Créer l'image finale minimaliste
FROM alpine:latest

# Installer les certificats CA si nécessaire (pour les requêtes HTTPS)
RUN apk add --no-cache ca-certificates

# Répertoire de travail pour l'image finale
WORKDIR /root/

# Copier l'application compilée depuis l'étape de build
COPY --from=builder /app/myapp .

# Créer le répertoire data dans l'image finale
RUN mkdir -p /root/data

# Exposer le port utilisé par l'application (si nécessaire)
EXPOSE 8080

# Commande par défaut pour lancer l'application
CMD ["./myapp"]
