FROM golang:1.25-alpine AS builder

WORKDIR /app

# Téléchargement des dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copie du code source
COPY . .

# Build de l'application gfacial-scaleway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /gfacial-scaleway ./cmd/gfacial-scaleway

# Image finale allégée avec Distroless
FROM gcr.io/distroless/static-debian12:latest

WORKDIR /app

# Les certificats SSL et data de timezone sont déjà inclus dans l'image distroless.

# Copie de l'exécutable depuis le builder
COPY --from=builder /gfacial-scaleway /app/gfacial-scaleway

# Copie du répertoire front
COPY front/ /app/front/

# Exposition du port
EXPOSE 8080

# Commande de lancement
CMD ["/app/gfacial-scaleway"]
