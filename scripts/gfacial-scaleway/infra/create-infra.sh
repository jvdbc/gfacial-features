#!/bin/bash
set -e

CLUSTER_NAME="gfacial-cluster"
POOL_NAME="gfacial-pool"
NODE_TYPE="BASIC2-A2C-8G"
BUCKET_NAME="gfacial-images"
REGION="fr-par"

echo "=== Creating Scaleway Infrastructure ==="

echo "1/4 - Creating Kapsule cluster..."
EXISTING=$(scw k8s cluster list name="$CLUSTER_NAME" region="$REGION" --output json | jq -r '.[0].id // empty')
if [ -z "$EXISTING" ]; then
    CLUSTER_JSON=$(scw k8s cluster create name="$CLUSTER_NAME" region="$REGION" --output json)
    CLUSTER_ID=$(echo "$CLUSTER_JSON" | jq -r '.id')
    echo "   Cluster created: $CLUSTER_ID"
else
    CLUSTER_ID="$EXISTING"
    echo "   Cluster already exists: $CLUSTER_ID"
fi

echo "   Waiting for cluster to be ready..."
while true; do
    STATUS=$(scw k8s cluster get cluster-id="$CLUSTER_ID" region="$REGION" --output json | jq -r '.status')
    if [ "$STATUS" = "ready" ] || [ "$STATUS" = "pool_required" ]; then
        break
    fi
    echo "   Status: $STATUS (waiting...)"
    sleep 5
done
echo "   Cluster ready!"

echo "2/4 - Creating node pool..."
scw k8s pool create cluster-id="$CLUSTER_ID" name="$POOL_NAME" node-type="$NODE_TYPE" size=1 region="$REGION"
echo "   Waiting for pool to be ready..."
scw k8s pool wait "$POOL_NAME" cluster-id="$CLUSTER_ID" region="$REGION" 2>/dev/null || sleep 10
echo "   Pool created!"

echo "   Waiting for cluster to be ready after pool creation..."
while true; do
    STATUS=$(scw k8s cluster get cluster-id="$CLUSTER_ID" region="$REGION" --output json | jq -r '.status')
    if [ "$STATUS" = "ready" ]; then
        break
    fi
    echo "   Status: $STATUS (waiting...)"
    sleep 5
done

echo "3/4 - Creating S3 bucket (if not exists)..."
EXISTS=$(scw object bucket list region="$REGION" --output json | jq -r ".[] | select(.name == \"$BUCKET_NAME\") | .name // empty")
if [ "$EXISTS" != "$BUCKET_NAME" ]; then
    scw object bucket create "$BUCKET_NAME" region="$REGION" 2>/dev/null || echo "   Bucket already exists"
else
    echo "   Bucket already exists"
fi

echo "4/4 - Installing kubeconfig..."
sleep 5
scw k8s kubeconfig install "$CLUSTER_ID" > ~/.kube/config

echo "=== Infrastructure ready ==="
echo "Cluster: $CLUSTER_NAME (ID: $CLUSTER_ID)"
echo "Node pool: $POOL_NAME"
echo "Bucket: $BUCKET_NAME"