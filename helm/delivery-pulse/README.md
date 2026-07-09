# Delivery Pulse Helm Chart

Deploys the Delivery Pulse application (developer metrics dashboard) to Kubernetes.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.x
- An Azure DevOps PAT with Work Items (Read) and Code (Read) scope
- Container images pushed to a registry accessible by the cluster (for AKS)

## Deployment Modes

This chart supports two deployment modes:

### Local (Docker Desktop)

Uses NodePort to expose the frontend directly. Ideal for development/testing.

```bash
./deploy.sh install local
```

Access at: `http://localhost:30080`

### Dev (AKS with Manual App Gateway)

Uses a LoadBalancer service with an optional reserved IP. Configure the Azure Application Gateway backend pool to point at the UI service IP.

```bash
./deploy.sh install dev
```

Then in Azure portal:
1. **Backend pool**: Point to the UI LoadBalancer IP
2. **Listener**: HTTP/HTTPS on your hostname (e.g., `dpulse.solvasfabric.com`)
3. **Rule**: Route listener traffic to the backend pool
4. **DNS**: Point `dpulse.solvasfabric.com` to the App Gateway's public IP

## Using the Deploy Script

The preferred way to deploy is via `deploy.sh` at the project root:

```bash
./deploy.sh install [local|dev] [--test|-t]
./deploy.sh uninstall [local|dev]
./deploy.sh help
```

The script:
- Switches kubectl context automatically (`docker-desktop` for local, `aks-eus-nonprd-shared-01` for dev)
- Prompts for the ADO PAT securely (never in shell history or process args)
- Detects install vs upgrade automatically
- Supports dry-run mode (`--test`)

### Bash Completion

```bash
source ./deploy-completion.bash
# Or add to ~/.bashrc:
# source /path/to/delivery-pulse/deploy-completion.bash
# alias dpulse='/path/to/delivery-pulse/deploy.sh'
```

## Building & Pushing Images

```bash
# Build images
docker build -t solvassharedservicesacr.azurecr.io/solvas-pulse-server:latest ./backend
docker build -t solvassharedservicesacr.azurecr.io/solvas-pulse-ui:latest ./frontend

# Push to ACR
az acr login --name solvassharedservicesacr
docker push solvassharedservicesacr.azurecr.io/solvas-pulse-server:latest
docker push solvassharedservicesacr.azurecr.io/solvas-pulse-ui:latest
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `imageRegistry` | Container registry prefix | `solvassharedservicesacr.azurecr.io` |
| `server.replicaCount` | Server pod replicas | `1` |
| `server.image.repository` | Server image name | `solvas-pulse-server` |
| `server.image.tag` | Server image tag (defaults to appVersion) | `""` |
| `server.containerPort` | Go server listen port | `8090` |
| `server.service.port` | Server service port | `8090` |
| `ui.replicaCount` | UI pod replicas | `1` |
| `ui.image.repository` | UI image name | `solvas-pulse-ui` |
| `ui.image.tag` | UI image tag (defaults to appVersion) | `""` |
| `ui.service.type` | `NodePort` (local) or `LoadBalancer` (dev) | `ClusterIP` |
| `ui.service.port` | UI service port | `80` |
| `ui.service.nodePort` | NodePort number (local mode) | `` |
| `ui.service.annotations` | Service annotations (for LB IP, etc.) | `{}` |
| `config.adoOrgUrl` | Azure DevOps org URL | `https://dev.azure.com/solvas` |
| `config.adoProject` | Azure DevOps project name | `Solvas` |
| `config.adoTeams` | Comma-separated team names | `AssetMgmt Conversion` |
| `secrets.adoPat` | Azure DevOps PAT (via --set-file) | `""` |

## Uninstalling

```bash
./deploy.sh uninstall local   # or dev
```
