# gfacial-scaleway

Projet personnel destiné à servir de support pour un article du blog [Blog Ippon Technologies](https://blog.ippon.fr/).

## Objectif

Déployer une application Go sur Scaleway Kapsule (Kubernetes) qui:
1. Analyse une photo de visage humain via l'API Scaleway OpenAI (Pixtral-12B)
2. Stocke les images envoyées sur Scaleway Object Storage (S3)

Utilise le framework [scaleway-sdk-go](https://github.com/scaleway/scaleway-sdk-go).

## Build

- `make build` - builds only `gfacial-scaleway` for linux/arm64
- Output: `build/linux_arm64/`
- Binary: `gfacial-scaleway`

## Commands

- `gfacial-scaleway` (HTTP server): requires `SCW_SECRET_KEY`, serves on :8080

## API

- **IA**: Scaleway OpenAI API (Pixtral-12B) - analyse les traits du visage:
  - Eye color
  - Hair color and length
  - Skin color
  - Facial hair (mustache, beard, goatee)
  - Gender
  - Accessories (glasses, piercings)

- **Storage**: Scaleway Object Storage (S3) - stocke les images uploadées

## Docker

- Dockerfile builds `gfacial-scaleway` for linux/arm64
- Frontend served from `front/`

## Deployment (Scaleway Kapsule)

```bash
# Create K8s cluster
scw k8s cluster create name=gfacial-cluster

# Get cluster-id
scw k8s cluster list

# Add kubeconfig
scw k8s kubeconfig install $cluster-id

# Create node pool (ARM)
scw k8s pool create cluster-id=$cluster-id name=gfacial-pool node-type=BASIC2-A4C8G size=1

# Install aistor (S3) - see README.md for full instructions
```

## Env Vars

- `SCW_SECRET_KEY` - Scaleway OpenAI API key
- `SCW_PROJECT_ID` - Scaleway project ID
- `SCW_REGION` - Scaleway region (default: fr-par)
- `BUCKET_NAME` - Scaleway Object Storage bucket name