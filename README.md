# Project: gfacial-features

## Project purpose
- Golang project to analyze a photo of a human face using github.com/googleapis/go-genai
    - obtain the facial features in the image:
        * Eye color
        * Hair color and length
        * Skin color
        * Facial hair: mustache, beard, goatee
        * Male/female
        * Accessories: glasses, piercings

## General Instructions
<!-- - shell client using github.com/urfave/cli/v2 -->

## Coding Style
- Use golang idiomatic coding style

# Project: gfacial-scaleway

## Project purpose

* Test scaleway cloud provider with small AI project

    * Scaleway generative API with Mistral LLM
    * Scaleway object storage (S3)

        * Save face pictures
        * AIStore (new minio server and client) on scaleway kubernetes

## Requirements (k8s)
* Install [kubectl](https://kubernetes.io/fr/docs/tasks/tools/install-kubectl/)

* Install [helm](https://helm.sh/fr/docs/intro/install/)

* Install [minio client](https://www.min.io/download) on your machine
    
    * request [licence key](https://www.min.io/download) and install it into ~/minio/minio.licence
    * install [client](https://www.min.io/download/aistor-client?platform=linux)

* Install [Scaleway client](https://github.com/scaleway/scaleway-cli)

    * Generate your account credentials using [scaleway console](https://console.scaleway.com/iam/api-keys) or follow [client quickstart](https://www.scaleway.com/en/docs/scaleway-cli/quickstart/)
    
    ```bash
    # create ~/.config/scw/config.yaml
    scw init
    # or
    scw login
    ```
    
    * Export your scaleway object config into mc config file
    ```bash
    # Generate mc config file for your scaleway object storage account
    scw object config install region=fr-par type=mc
    ```
    
    * Create your bucket
    ```bash
    scw object bucket create mon-bucket
    ```
    
    * [Create](https://www.scaleway.com/en/docs/object-storage/api-cli/installing-minio-client/) mc alias for your new bucket
    ```bash
    # mc alias set <ALIAS> <YOUR-S3-ENDPOINT> --api <API-SIGNATURE>
    mc alias set s3-scw https://s3.fr-par.scw.cloud --api S3v4
    # Give accessKey and secretKey
    ```
    
    * Create your k8s
    ```bash
    # Default parameters
    # type=kapsule
    # control-plan=Mutualized
    # version=latest
    scw k8s cluster create name=mon-cluster
    
    # Get cluster-id
    scw k8s cluster list
    
    # Add scw k8s config into your kubeconfig for new scw k8s-cluster
    scw k8s kubeconfig install $cluster-id

    # Create a cluster-nodes-pool (ARM 4vCPU 8Go )
    scw k8s pool create cluster-id=$cluster-id name=cluster-pool node-type=BASIC2-A4C-8G size=1
    # scw k8s pool list cluster-id=$cluster-id
    # scw k8s pool delete cluster-id=$cluster-id
    
    # Add minio repository 
    helm repo add minio https://helm.min.io/
    helm repo list


    # https://docs.min.io/enterprise/aistor-object-store/installation/kubernetes/install/deploy-aistor-on-kubernetes/
    # licence in ~/minio/minio.licence
    # optional: customize [helm chart](https://docs.min.io/enterprise/aistor-object-store/reference/kubernetes/object-store-operator-helm-chart/)

    # install aistor
    helm install aistor minio/aistor-operator -n aistor --create-namespace --set license="$(cat ~/minio/minio.license)" # -f custom-aistor.yaml
    # helm uninstall aistor -n aistor

    ```