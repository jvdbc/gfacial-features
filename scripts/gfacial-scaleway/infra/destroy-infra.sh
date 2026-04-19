#!/bin/bash
set -e

CLUSTER_NAME="gfacial-cluster"
REGION="fr-par"

echo "=== Destroying Scaleway Infrastructure ==="

echo "1/4 - Getting cluster ID..."
CLUSTER_ID=$(scw k8s cluster list name="$CLUSTER_NAME" region="$REGION" --output json | jq -r '.[0].id // empty')
if [ -z "$CLUSTER_ID" ]; then
    echo "   No cluster found, skipping..."
else
    echo "   Cluster ID: $CLUSTER_ID"

    echo "2/4 - Getting pool ID..."
    POOL_ID=$(scw k8s pool list cluster-id="$CLUSTER_ID" region="$REGION" --output json | jq -r '.[0].id // empty')
    echo "   Pool ID: $POOL_ID"

    echo "3/4 - Deleting node pool and cluster..."
    if [ -n "$POOL_ID" ]; then
        scw k8s pool delete "$POOL_ID" region="$REGION" --wait
        echo "   Pool deleted"
    fi
    scw k8s cluster delete "$CLUSTER_ID" region="$REGION" --wait
    echo "   Cluster deleted"
fi

echo "4/4 - Removing kubeconfig..."
kubectl config delete-context "$CLUSTER_NAME" 2>/dev/null || true
kubectl config delete-cluster "$CLUSTER_NAME" 2>/dev/null || true

echo "=== Infrastructure destroyed ==="