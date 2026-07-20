#!/usr/bin/env bash
#
# deploy.sh — install/upgrade or uninstall the delivery-pulse Helm chart.
#
# Runnable from ANY directory: every path resolves relative to this script's own
# location (not the current working dir), so a symlink / PATH entry / alias all work.
#
# Usage:
#   deploy.sh install [local|dev] [--test|-t]
#   deploy.sh uninstall [local|dev]
#   deploy.sh help

set -euo pipefail

# ---------------------------------------------------------------------------
# Resolve this script's directory, following symlinks.
# ---------------------------------------------------------------------------
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
  DIR="$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)"
  SOURCE="$(readlink "$SOURCE")"
  [[ "$SOURCE" != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)"

# ---------------------------------------------------------------------------
# Terminal colors (only when writing to a terminal).
# ---------------------------------------------------------------------------
if [ -t 1 ]; then
  col_cyan=$'\033[1;36m'
  col_green=$'\033[1;32m'
  col_red=$'\033[1;31m'
  col_yellow=$'\033[1;33m'
  col_nc=$'\033[0m'
else
  col_cyan="" col_green="" col_red="" col_yellow="" col_nc=""
fi

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
CHART_DIR="$SCRIPT_DIR/helm/delivery-pulse"
STACK_NAME="delivery-pulse"

# Kubectl contexts per environment
declare -A CONTEXTS=( [local]="docker-desktop" [dev]="aks-eus-nonprd-shared-01" )
declare -A NAMESPACES=( [local]="delivery-pulse" [dev]="delivery-pulse" )

# ---------------------------------------------------------------------------
# Functions
# ---------------------------------------------------------------------------
display_help() {
  cat <<EOF
${col_cyan}deploy.sh${col_nc} — Delivery Pulse Helm deployment helper

Usage:
  deploy.sh <command> [env] [options]

Commands:
  install   [local|dev] [--test|-t]   Install or upgrade the release.
  uninstall [local|dev]               Uninstall the release.
  help                                Show this help.

Environments:
  local   Docker Desktop Kubernetes (default).
          Values file: helm/delivery-pulse/values-local.yaml
  dev     AKS cluster with Azure App Gateway.
          Values file: helm/delivery-pulse/values-dev.yaml

Options:
  --test, -t   Helm dry-run only (no changes applied).

Examples:
  deploy.sh install                   # Install to local docker-desktop
  deploy.sh install dev               # Install to AKS dev cluster
  deploy.sh install dev --test        # Dry-run against AKS dev
  deploy.sh uninstall dev             # Remove from AKS dev
EOF
}

splash() {
  echo ""
  echo "  ${col_cyan}Deployment Summary${col_nc}"
  echo "  ─────────────────────────────────────────"
  printf "  ${col_cyan}%-25s${col_nc}: %s\n" "Environment" "$env_label"
  printf "  ${col_cyan}%-25s${col_nc}: %s\n" "Release Name" "$STACK_NAME"
  printf "  ${col_cyan}%-25s${col_nc}: %s\n" "Kubectl Context" "$context"
  printf "  ${col_cyan}%-25s${col_nc}: %s\n" "Namespace" "$namespace"
  printf "  ${col_cyan}%-25s${col_nc}: %s\n" "Values File" "$values_file"
  if [ "$dry_run" == true ]; then
    printf "  ${col_yellow}%-25s${col_nc}: %s\n" "Mode" "DRY-RUN (no changes)"
  fi
  echo "  ─────────────────────────────────────────"
  echo ""
}

set_k8s_context() {
  local target="$1"
  local current
  current="$(kubectl config current-context 2>/dev/null || true)"
  if [ "$current" != "$target" ]; then
    if ! kubectl config use-context "$target" >/dev/null 2>&1; then
      echo "${col_red}Error: could not switch kubectl context to '$target'. Is it configured?${col_nc}" >&2
      exit 1
    fi
  fi
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
if [ $# -eq 0 ] || [ "$1" == "help" ] || [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
  display_help
  exit 0
fi

helm_command="$1"
shift

if [[ "$helm_command" != "install" && "$helm_command" != "uninstall" ]]; then
  echo "${col_red}Error: Invalid command '$helm_command'. Use 'install' or 'uninstall'.${col_nc}" >&2
  display_help
  exit 1
fi

# Defaults
env_label="local"
dry_run=false

# Parse env label (first non-flag argument)
if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
  env_label="$1"
  shift
  if [[ "$env_label" != "local" && "$env_label" != "dev" ]]; then
    echo "${col_red}Error: Invalid environment '$env_label'. Allowed values: 'local' or 'dev'.${col_nc}" >&2
    exit 1
  fi
fi

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --test|-t) dry_run=true ;;
    *) echo "${col_red}Error: Unknown option '$1'${col_nc}" >&2; exit 1 ;;
  esac
  shift
done

# Resolve environment-specific values
context="${CONTEXTS[$env_label]}"
namespace="${NAMESPACES[$env_label]}"
values_file="$CHART_DIR/values-${env_label}.yaml"

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
if ! command -v helm &>/dev/null; then
  echo "${col_red}Error: helm is not installed or not in PATH.${col_nc}" >&2
  exit 1
fi

if ! command -v kubectl &>/dev/null; then
  echo "${col_red}Error: kubectl is not installed or not in PATH.${col_nc}" >&2
  exit 1
fi

if [ ! -d "$CHART_DIR" ]; then
  echo "${col_red}Error: Helm chart not found at '$CHART_DIR'.${col_nc}" >&2
  echo "  deploy.sh resolves files relative to its own location ($SCRIPT_DIR)." >&2
  echo "  Don't copy it onto your PATH — alias the real file instead, e.g.:" >&2
  echo "    alias dpulse='$SCRIPT_DIR/deploy.sh'" >&2
  exit 1
fi

if [ ! -f "$values_file" ]; then
  echo "${col_red}Error: Values file not found at '$values_file'.${col_nc}" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Secret handling (only for real installs, never shown in terminal history)
# ---------------------------------------------------------------------------
secret_value=""
if [ "$helm_command" == "install" ] && [ "$dry_run" == false ]; then
  read -r -s -p "Enter your Service Principal client secret: " secret_value
  echo
  if [ -z "$secret_value" ]; then
    echo "${col_red}Error: Client secret cannot be empty.${col_nc}" >&2
    exit 1
  fi
fi

# ---------------------------------------------------------------------------
# Context switch & summary
# ---------------------------------------------------------------------------
set_k8s_context "$context"
splash

# ---------------------------------------------------------------------------
# Execute
# ---------------------------------------------------------------------------
if [ "$helm_command" == "uninstall" ]; then
  echo "Uninstalling $STACK_NAME from $env_label..."
  helm uninstall "$STACK_NAME" --namespace "$namespace" || true
  echo "${col_green}Done.${col_nc}"
  exit 0
fi

# --- Install / Upgrade ---

# Detect install vs upgrade
if helm list --namespace "$namespace" -f "^${STACK_NAME}$" --no-headers -q 2>/dev/null | grep -q "$STACK_NAME"; then
  action="upgrade"
  echo "Release '$STACK_NAME' exists — upgrading..."
else
  action="install"
  echo "Release '$STACK_NAME' not found — installing..."
fi

# Write secret to a temp file (never in process args or shell history)
secretfile="$(mktemp)"
trap 'rm -f "$secretfile"' EXIT
printf '%s' "$secret_value" > "$secretfile"

# Read developers list from file (one email per line, comments/blanks ignored)
developers_file="$SCRIPT_DIR/developers.txt"
developers_csv=""
if [ -f "$developers_file" ]; then
  developers_csv=$(grep -v '^\s*#' "$developers_file" | grep -v '^\s*$' | tr '\n' ',' | sed 's/,$//')
  if [ -n "$developers_csv" ]; then
    echo "  ${col_cyan}Developers:${col_nc} $(echo "$developers_csv" | tr ',' '\n' | wc -l | tr -d ' ') entries from developers.txt"
  fi
fi

# Build helm args — clean and minimal, all env-specific config lives in the values file
helm_args=(
  "$action" "$STACK_NAME"
  "$CHART_DIR"
  --namespace "$namespace"
  --create-namespace
  -f "$values_file"
)

# Secret via --set-file so it never appears in process args
if [ -s "$secretfile" ]; then
  helm_args+=(--set-file "secrets.adoClientSecret=$secretfile")
fi

# Developers list via --set-string (commas escaped for helm)
if [ -n "$developers_csv" ]; then
  escaped_devs="${developers_csv//,/\\,}"
  helm_args+=(--set "config.adoDevelopers=$escaped_devs")
fi

# Dry-run mode
if [ "$dry_run" == true ]; then
  helm_args+=(--dry-run --debug)
fi

# Show what helm will receive
echo "  ${col_yellow}Helm command:${col_nc} helm ${action} ${STACK_NAME}"
echo "  ${col_yellow}Chart:${col_nc}        ${CHART_DIR}"
echo "  ${col_yellow}Values:${col_nc}       ${values_file}"
echo "  ${col_yellow}Namespace:${col_nc}    ${namespace}"
echo ""

# Execute helm
helm_output=$(helm "${helm_args[@]}" 2>&1)

if [ "$dry_run" == true ]; then
  manifest_file="$SCRIPT_DIR/release_manifest.yaml"
  echo "$helm_output" > "$manifest_file"
  echo "${col_green}Dry-run complete. Manifest written to: $manifest_file${col_nc}"
else
  echo "$helm_output"
  echo ""
  if [ "$env_label" == "local" ]; then
    echo "${col_green}Access the app at: http://localhost:30080${col_nc}"
  else
    echo "${col_green}Deployed to AKS.${col_nc}"
    echo "  Next steps:"
    echo "  1. Get UI service IP:  kubectl get svc ${STACK_NAME}-ui -n ${namespace}"
    echo "  2. Configure App Gateway backend pool with that IP"
    echo "  3. Point DNS to App Gateway public IP"
  fi
fi
