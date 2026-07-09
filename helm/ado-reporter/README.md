# ADO Reporter Helm Chart

Deploys the Azure DevOps Work Items Reporting application to Kubernetes.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.x
- An Azure DevOps PAT with Work Items (Read) scope
- Container images pushed to a registry accessible by AKS

## Building & Pushing Images

```bash
# Build images
docker build -t <your-acr>.azurecr.io/ado-reporter-backend:latest ./backend
docker build -t <your-acr>.azurecr.io/ado-reporter-frontend:latest ./frontend

# Push to ACR
docker push <your-acr>.azurecr.io/ado-reporter-backend:latest
docker push <your-acr>.azurecr.io/ado-reporter-frontend:latest
```

## Installation

```bash
helm install ado-reporter ./helm/ado-reporter \
  --namespace ado-reporter \
  --create-namespace \
  --set imageRegistry=<your-acr>.azurecr.io \
  --set config.adoOrgUrl="https://dev.azure.com/your-org" \
  --set config.adoProject="your-project" \
  --set secrets.adoPat="your-pat-token" \
  --set ingress.host="ado-reporter.your-domain.com"
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `imageRegistry` | Container registry prefix | `""` |
| `backend.replicaCount` | Backend pod replicas | `2` |
| `backend.image.repository` | Backend image name | `ado-reporter-backend` |
| `backend.image.tag` | Backend image tag | `latest` |
| `frontend.replicaCount` | Frontend pod replicas | `2` |
| `frontend.image.repository` | Frontend image name | `ado-reporter-frontend` |
| `frontend.image.tag` | Frontend image tag | `latest` |
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.host` | Ingress hostname | `ado-reporter.example.com` |
| `ingress.tls.enabled` | Enable TLS | `false` |
| `config.adoOrgUrl` | Azure DevOps org URL | `""` |
| `config.adoProject` | Azure DevOps project name | `""` |
| `secrets.adoPat` | Azure DevOps PAT | `""` |

## Upgrading

```bash
helm upgrade ado-reporter ./helm/ado-reporter \
  --namespace ado-reporter \
  --reuse-values \
  --set backend.image.tag=v1.1.0 \
  --set frontend.image.tag=v1.1.0
```

## Uninstalling

```bash
helm uninstall ado-reporter --namespace ado-reporter
```
