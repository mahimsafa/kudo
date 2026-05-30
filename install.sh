#!/usr/bin/env bash
set -euo pipefail

REPO="mahimsafa/kudo"
BINARY_NAME="kudo"
MIN_GO_VERSION="1.23"

# --- Colors -----------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { printf "${CYAN}[INFO]${NC}  %s\n" "$*"; }
ok()    { printf "${GREEN}[OK]${NC}    %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*"; }
fatal() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; exit 1; }

# --- Defaults ----------------------------------------------------------
INSTALL_SCOPE=""       # system | user
VERSION=""             # e.g. v0.1.0, empty = latest
ACTION="install"       # install | uninstall
DEV_MODE=false
SOURCE_DIR=""

# --- Usage -------------------------------------------------------------
usage() {
    cat <<'EOF'
Kudo installer — download and set up the kudo agent.

Usage:
  install.sh [options]

Options:
  --system          Install system-wide (/usr/local/bin, systemd system unit)
  --user            Install for current user (~/.local/bin, systemd user unit)
  --dev             Build from source (clone repo, make build) instead of downloading a release
  --version VER     Install a specific release version (default: latest; ignored with --dev)
  --uninstall       Remove kudo binary, systemd unit, and optionally data
  -h, --help        Show this help
EOF
    exit 0
}

# --- Parse args --------------------------------------------------------
while [[ $# -gt 0 ]]; do
    case "$1" in
        --system)    INSTALL_SCOPE="system"; shift ;;
        --user)      INSTALL_SCOPE="user";   shift ;;
        --dev)       DEV_MODE=true;          shift ;;
        --version)   VERSION="$2";           shift 2 ;;
        --uninstall) ACTION="uninstall";     shift ;;
        -h|--help)   usage ;;
        *) fatal "Unknown option: $1" ;;
    esac
done

# --- Detect OS & arch --------------------------------------------------
detect_platform() {
    local os arch
    os="$(uname -s)"
    arch="$(uname -m)"

    if [[ "$DEV_MODE" == true ]]; then
        case "$os" in
            Linux|Darwin) ;;
            *) fatal "Unsupported OS: $os. Dev install supports Linux and macOS." ;;
        esac
    else
        case "$os" in
            Linux) ;;
            *) fatal "Unsupported OS: $os. Kudo only supports Linux." ;;
        esac
    fi

    case "$arch" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *) fatal "Unsupported architecture: $arch. Kudo supports amd64 and arm64." ;;
    esac

    PLATFORM_OS="$os"
    PLATFORM_ARCH="$arch"
}

# --- Resolve paths based on scope -------------------------------------
resolve_paths() {
    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        BIN_DIR="/usr/local/bin"
        DATA_DIR="/var/lib/kudo"
        CONFIG_DIR="/etc/kudo"
        SYSTEMD_DIR="/etc/systemd/system"
        SYSTEMD_USER_FLAG=""
    else
        BIN_DIR="$HOME/.local/bin"
        DATA_DIR="$HOME/.local/share/kudo"
        CONFIG_DIR="$HOME/.config/kudo"
        SYSTEMD_DIR="$HOME/.config/systemd/user"
        SYSTEMD_USER_FLAG="--user"
    fi
}

# --- Prompt for install scope if not set via flag ----------------------
prompt_scope() {
    if [[ -n "$INSTALL_SCOPE" ]]; then
        return
    fi

    printf "\n${BOLD}Where would you like to install kudo?${NC}\n"
    printf "  1) System-wide  (/usr/local/bin — requires sudo)\n"
    printf "  2) Current user (~/.local/bin)\n"
    printf "\n"

    while true; do
        read -rp "Choose [1/2]: " choice
        case "$choice" in
            1) INSTALL_SCOPE="system"; break ;;
            2) INSTALL_SCOPE="user";   break ;;
            *) printf "Please enter 1 or 2.\n" ;;
        esac
    done
}

# --- Run command with scope-appropriate privileges ---------------------
run_scope() {
    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        sudo env "PATH=${PATH}" "$@"
    else
        "$@"
    fi
}

run_root() {
    sudo env "PATH=${PATH}" "$@"
}

# --- Fetch helpers (curl preferred, wget fallback) ---------------------
has_cmd() { command -v "$1" &>/dev/null; }

http_get() {
    local url="$1"
    if has_cmd curl; then
        curl -fsSL "$url"
    elif has_cmd wget; then
        wget -qO- "$url"
    else
        fatal "Neither curl nor wget found. Install one and retry."
    fi
}

http_download() {
    local url="$1" dest="$2"
    if has_cmd curl; then
        curl -fsSL -o "$dest" "$url"
    elif has_cmd wget; then
        wget -qO "$dest" "$url"
    else
        fatal "Neither curl nor wget found. Install one and retry."
    fi
}

# --- Resolve latest version from GitHub --------------------------------
resolve_version() {
    if [[ -n "$VERSION" ]]; then
        [[ "$VERSION" == v* ]] || VERSION="v$VERSION"
        return
    fi

    info "Fetching latest release version..."
    VERSION=$(http_get "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')

    if [[ -z "$VERSION" ]]; then
        fatal "Could not determine latest version. Specify one with --version."
    fi

    ok "Latest version: $VERSION"
}

# --- Download & verify binary ------------------------------------------
download_binary() {
    local asset="kudo-linux-${PLATFORM_ARCH}"
    local base_url="https://github.com/${REPO}/releases/download/${VERSION}"
    local tmp_dir
    tmp_dir="$(mktemp -d)"

    info "Downloading ${asset} (${VERSION})..."
    http_download "${base_url}/${asset}" "${tmp_dir}/${BINARY_NAME}"

    info "Downloading checksums..."
    http_download "${base_url}/checksums.txt" "${tmp_dir}/checksums.txt"

    info "Verifying checksum..."
    local expected actual
    expected=$(grep "${asset}" "${tmp_dir}/checksums.txt" | awk '{print $1}')
    actual=$(sha256sum "${tmp_dir}/${BINARY_NAME}" | awk '{print $1}')

    if [[ "$expected" != "$actual" ]]; then
        rm -rf "$tmp_dir"
        fatal "Checksum mismatch!\n  Expected: ${expected}\n  Got:      ${actual}"
    fi
    ok "Checksum verified."

    DOWNLOADED_BINARY="${tmp_dir}/${BINARY_NAME}"
}

# --- Dev install prerequisites -----------------------------------------
current_go_version() {
    if ! has_cmd go; then
        return
    fi
    go env GOVERSION 2>/dev/null | sed 's/^go//'
}

version_ge() {
    local ver="$1" min="$2"
    [[ "$(printf '%s\n' "$min" "$ver" | sort -V | head -1)" == "$min" ]]
}

go_bin_dir() {
    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        printf '/usr/local/go/bin'
    else
        printf '%s/.local/go/bin' "$HOME"
    fi
}

ensure_go_in_path() {
    local go_bin
    go_bin="$(go_bin_dir)"
    if [[ -d "$go_bin" ]] && [[ ":$PATH:" != *":${go_bin}:"* ]]; then
        export PATH="${go_bin}:${PATH}"
    fi
}

install_linux_pkg() {
    local pkg="$1"
    if has_cmd apt-get; then
        run_root env DEBIAN_FRONTEND=noninteractive apt-get update -qq
        run_root env DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "$pkg"
    elif has_cmd dnf; then
        run_root dnf install -y "$pkg"
    elif has_cmd yum; then
        run_root yum install -y "$pkg"
    elif has_cmd apk; then
        run_root apk add --no-cache "$pkg"
    elif has_cmd pacman; then
        run_root pacman -Sy --noconfirm "$pkg"
    elif has_cmd zypper; then
        run_root zypper install -y "$pkg"
    else
        fatal "Cannot install ${pkg} automatically. Install it and retry."
    fi
}

install_git() {
    info "Installing git..."
    case "$PLATFORM_OS" in
        Linux) install_linux_pkg git ;;
        Darwin)
            if has_cmd brew; then
                brew install git
            elif ! xcode-select -p &>/dev/null 2>&1; then
                warn "Installing Xcode Command Line Tools (includes git)..."
                xcode-select --install
                fatal "Complete the Xcode Command Line Tools install, then rerun this script."
            else
                fatal "git not found. Install git via Homebrew or Xcode Command Line Tools and retry."
            fi
            ;;
    esac
    ok "git installed."
}

install_make() {
    info "Installing make..."
    case "$PLATFORM_OS" in
        Linux)
            if has_cmd apt-get; then
                run_root env DEBIAN_FRONTEND=noninteractive apt-get update -qq
                run_root env DEBIAN_FRONTEND=noninteractive apt-get install -y -qq build-essential
            else
                install_linux_pkg make
            fi
            ;;
        Darwin)
            if has_cmd make; then
                return
            fi
            if has_cmd brew; then
                brew install make
            elif ! xcode-select -p &>/dev/null 2>&1; then
                warn "Installing Xcode Command Line Tools (includes make)..."
                xcode-select --install
                fatal "Complete the Xcode Command Line Tools install, then rerun this script."
            else
                fatal "make not found. Install make via Homebrew or Xcode Command Line Tools and retry."
            fi
            ;;
    esac
    ok "make installed."
}

install_go_latest() {
    local latest go_os url tmp_dir go_root
    info "Fetching latest Go version..."
    latest="$(http_get "https://go.dev/VERSION?mode=text" | tr -d '\n')"
    latest="${latest#go}"

    case "$PLATFORM_OS" in
        Linux)  go_os="linux" ;;
        Darwin) go_os="darwin" ;;
        *) fatal "Unsupported OS for Go install: ${PLATFORM_OS}" ;;
    esac

    url="https://go.dev/dl/go${latest}.${go_os}-${PLATFORM_ARCH}.tar.gz"
    tmp_dir="$(mktemp -d)"

    info "Installing Go ${latest}..."
    http_download "$url" "${tmp_dir}/go.tar.gz"

    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "${tmp_dir}/go.tar.gz"
        go_root="/usr/local/go"
    else
        go_root="${HOME}/.local/go"
        rm -rf "$go_root"
        mkdir -p "${HOME}/.local"
        tar -C "${HOME}/.local" -xzf "${tmp_dir}/go.tar.gz"
    fi

    rm -rf "$tmp_dir"
    export PATH="${go_root}/bin:${PATH}"
    ok "Go ${latest} installed to ${go_root}"

    if [[ "$INSTALL_SCOPE" == "user" ]] && [[ ":$PATH:" != *":${go_root}/bin:"* ]]; then
        warn "${go_root}/bin is not in your PATH."
        warn "Add it with:  export PATH=\"${go_root}/bin:\$PATH\""
    fi
}

ensure_git() {
    if has_cmd git; then
        return
    fi
    install_git
}

ensure_make() {
    if has_cmd make; then
        return
    fi
    install_make
}

ensure_go() {
    ensure_go_in_path

    local current=""
    if has_cmd go; then
        current="$(current_go_version)"
    fi

    if [[ -n "$current" ]] && version_ge "$current" "$MIN_GO_VERSION"; then
        ok "Go ${current} meets minimum (${MIN_GO_VERSION})."
        return
    fi

    if [[ -n "$current" ]]; then
        warn "Go ${current} is below minimum ${MIN_GO_VERSION}; installing latest Go."
    else
        info "Go not found; installing latest Go."
    fi

    install_go_latest

    current="$(current_go_version)"
    if [[ -z "$current" ]] || ! version_ge "$current" "$MIN_GO_VERSION"; then
        fatal "Go install failed or version ${current:-unknown} is still below ${MIN_GO_VERSION}."
    fi
}

ensure_dev_prerequisites() {
    ensure_git
    ensure_make
    ensure_go
    ensure_go_in_path
    ok "Build tools ready (git $(git --version | awk '{print $3}'), go $(current_go_version), make $(make --version | head -1 | awk '{print $3}' | tr -d '\r'))."
}

# --- Clone repo and build from source ----------------------------------
build_from_source() {
    local repo_url="https://github.com/${REPO}.git"
    SOURCE_DIR="${DATA_DIR}/src"

    run_scope mkdir -p "$DATA_DIR"

    if [[ -d "${SOURCE_DIR}/.git" ]]; then
        info "Updating existing source at ${SOURCE_DIR}..."
        run_scope git -C "$SOURCE_DIR" fetch origin
        run_scope git -C "$SOURCE_DIR" checkout main
        run_scope git -C "$SOURCE_DIR" pull --ff-only origin main
    else
        info "Cloning ${repo_url} into ${SOURCE_DIR}..."
        run_scope git clone "$repo_url" "$SOURCE_DIR"
    fi

    info "Building from source..."
    run_scope make -C "$SOURCE_DIR" build

    DOWNLOADED_BINARY="${SOURCE_DIR}/bin/${BINARY_NAME}"
    if [[ ! -f "$DOWNLOADED_BINARY" ]]; then
        fatal "Build succeeded but binary not found at ${DOWNLOADED_BINARY}"
    fi

    local commit
    commit="$(run_scope git -C "$SOURCE_DIR" rev-parse --short HEAD)"
    VERSION="dev (${commit})"
    ok "Built ${VERSION} from main."
}

# --- Install binary ----------------------------------------------------
install_binary() {
    mkdir -p "$BIN_DIR"

    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        sudo install -m 755 "$DOWNLOADED_BINARY" "${BIN_DIR}/${BINARY_NAME}"
    else
        install -m 755 "$DOWNLOADED_BINARY" "${BIN_DIR}/${BINARY_NAME}"
    fi

    ok "Binary installed to ${BIN_DIR}/${BINARY_NAME}"

    if [[ "$INSTALL_SCOPE" == "user" ]] && [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        warn "${BIN_DIR} is not in your PATH."
        warn "Add it with:  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
}

# --- Create directories ------------------------------------------------
create_dirs() {
    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        sudo mkdir -p "$DATA_DIR" "$CONFIG_DIR"
    else
        mkdir -p "$DATA_DIR" "$CONFIG_DIR"
    fi
    ok "Created data dir:   ${DATA_DIR}"
    ok "Created config dir: ${CONFIG_DIR}"
}

# --- Ensure Docker -----------------------------------------------------
docker_usable() {
    docker info &>/dev/null
}

wait_for_docker() {
    local attempts="${1:-30}"
    local i

    for ((i = 1; i <= attempts; i++)); do
        if docker_usable; then
            return 0
        fi
        sleep 2
    done
    return 1
}

install_docker_linux() {
    local tmp_script
    info "Installing Docker..."

    if has_cmd apk; then
        run_root apk add --no-cache docker docker-cli-compose
        if has_cmd rc-update; then
            run_root rc-update add docker boot 2>/dev/null || true
        fi
    else
        tmp_script="$(mktemp)"
        http_download "https://get.docker.com" "$tmp_script"
        run_root sh "$tmp_script"
        rm -f "$tmp_script"
    fi

    if has_cmd systemctl; then
        run_root systemctl enable --now docker
    fi

    ok "Docker installed."
}

install_docker_darwin() {
    info "Installing Docker Desktop..."

    if has_cmd brew; then
        if ! brew list --cask docker &>/dev/null 2>&1; then
            brew install --cask docker
        fi
        if ! docker_usable; then
            info "Starting Docker Desktop..."
            open -a Docker 2>/dev/null || true
        fi
    else
        fatal "Install Homebrew (https://brew.sh) or Docker Desktop (https://docs.docker.com/desktop/install/mac-install/) and retry."
    fi
}

ensure_docker_group() {
    if [[ "$PLATFORM_OS" != "Linux" ]]; then
        return
    fi
    if docker_usable; then
        return
    fi
    if id -nG "$USER" 2>/dev/null | tr ' ' '\n' | grep -qx docker; then
        return
    fi
    if getent group docker &>/dev/null; then
        info "Adding ${USER} to the docker group..."
        run_root usermod -aG docker "$USER"
        warn "Added ${USER} to the docker group. Log out and back in, or run: newgrp docker"
    fi
}

ensure_docker_daemon() {
    if docker_usable; then
        return
    fi
    if ! has_cmd docker; then
        return
    fi

    info "Docker is installed but not reachable; trying to start it..."
    case "$PLATFORM_OS" in
        Linux)
            if has_cmd systemctl; then
                run_root systemctl start docker 2>/dev/null || true
            fi
            ;;
        Darwin)
            open -a Docker 2>/dev/null || true
            wait_for_docker 60 || true
            ;;
    esac
}

ensure_docker() {
    if ! has_cmd docker; then
        case "$PLATFORM_OS" in
            Linux)  install_docker_linux ;;
            Darwin) install_docker_darwin ;;
            *)      fatal "Cannot install Docker on ${PLATFORM_OS} automatically." ;;
        esac
    fi

    ensure_docker_daemon

    if docker_usable; then
        ok "Docker is available ($(docker --version | cut -d, -f1))."
        return
    fi

    ensure_docker_group
    wait_for_docker 15 || true

    if docker_usable; then
        ok "Docker is available ($(docker --version | cut -d, -f1))."
        return
    fi

    if run_root docker info &>/dev/null; then
        ok "Docker daemon is running. Log out and back in, or run: newgrp docker"
        return
    fi

    case "$PLATFORM_OS" in
        Darwin)
            fatal "Docker Desktop is installed but not running. Open Docker Desktop and retry."
            ;;
        *)
            fatal "Docker is required for container workloads but is not running. Start Docker and retry."
            ;;
    esac
}

# --- Install systemd service -------------------------------------------
install_service() {
    if ! has_cmd systemctl; then
        warn "systemctl not found — skipping service setup."
        return
    fi

    mkdir -p "$SYSTEMD_DIR"

    local bin_path="${BIN_DIR}/${BINARY_NAME}"
    local unit_file="${SYSTEMD_DIR}/kudo.service"

    if [[ "$INSTALL_SCOPE" == "system" ]]; then
        sudo tee "$unit_file" >/dev/null <<UNIT
[Unit]
Description=Kudo Agent
After=network-online.target docker.service
Wants=network-online.target
Requires=docker.service

[Service]
Type=simple
ExecStart=${bin_path} agent
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
UNIT
        sudo systemctl daemon-reload
    else
        cat > "$unit_file" <<UNIT
[Unit]
Description=Kudo Agent
After=network-online.target docker.service

[Service]
Type=simple
ExecStart=${bin_path} agent
Restart=on-failure
RestartSec=5
Environment=HOME=${HOME}

[Install]
WantedBy=default.target
UNIT
        systemctl --user daemon-reload
    fi

    ok "Systemd unit installed: ${unit_file}"
}

# --- Print post-install instructions -----------------------------------
print_next_steps() {
    local ctl="systemctl"
    [[ "$INSTALL_SCOPE" == "user" ]] && ctl="systemctl --user"

    printf "\n${GREEN}${BOLD}Kudo ${VERSION} installed successfully!${NC}\n"
    if [[ "$DEV_MODE" == true ]]; then
        printf "Built from ${SOURCE_DIR}\n"
    fi
    printf "\n"

    cat <<EOF
Next steps:

  1. Bootstrap a new cluster:
       ${ctl} start kudo
     Or run manually:
       ${BIN_DIR}/${BINARY_NAME} agent --bootstrap --name \$(hostname)

  2. Generate a join token (on leader):
       ${BINARY_NAME} token create --ttl 24h

  3. Join a node to an existing cluster:
       Edit ${CONFIG_DIR}/kudo.yaml then:
       ${ctl} start kudo
     Or run manually:
       ${BINARY_NAME} agent --join <leader-ip>:7946 --token <token> --name \$(hostname)

  4. Manage the service:
       ${ctl} start kudo
       ${ctl} stop kudo
       ${ctl} status kudo
       ${ctl} enable kudo    # start on boot

EOF
}

# --- Uninstall ---------------------------------------------------------
do_uninstall() {
    prompt_scope
    resolve_paths

    info "Uninstalling kudo (${INSTALL_SCOPE})..."

    local ctl="systemctl"
    [[ "$INSTALL_SCOPE" == "user" ]] && ctl="systemctl --user"

    if has_cmd systemctl; then
        $ctl stop kudo 2>/dev/null || true
        $ctl disable kudo 2>/dev/null || true
    fi

    local unit_file="${SYSTEMD_DIR}/kudo.service"
    if [[ -f "$unit_file" ]]; then
        if [[ "$INSTALL_SCOPE" == "system" ]]; then
            sudo rm -f "$unit_file"
            sudo systemctl daemon-reload
        else
            rm -f "$unit_file"
            systemctl --user daemon-reload
        fi
        ok "Removed systemd unit."
    fi

    local bin_path="${BIN_DIR}/${BINARY_NAME}"
    if [[ -f "$bin_path" ]]; then
        if [[ "$INSTALL_SCOPE" == "system" ]]; then
            sudo rm -f "$bin_path"
        else
            rm -f "$bin_path"
        fi
        ok "Removed binary."
    fi

    printf "\n"
    read -rp "Remove data directory ${DATA_DIR}? [y/N]: " remove_data
    if [[ "$remove_data" =~ ^[Yy]$ ]]; then
        if [[ "$INSTALL_SCOPE" == "system" ]]; then
            sudo rm -rf "$DATA_DIR"
        else
            rm -rf "$DATA_DIR"
        fi
        ok "Removed data directory."
    fi

    read -rp "Remove config directory ${CONFIG_DIR}? [y/N]: " remove_config
    if [[ "$remove_config" =~ ^[Yy]$ ]]; then
        if [[ "$INSTALL_SCOPE" == "system" ]]; then
            sudo rm -rf "$CONFIG_DIR"
        else
            rm -rf "$CONFIG_DIR"
        fi
        ok "Removed config directory."
    fi

    printf "\n${GREEN}${BOLD}Kudo has been uninstalled.${NC}\n"
}

# --- Main --------------------------------------------------------------
main() {
    printf "\n${BOLD}Kudo Installer${NC}\n\n"

    if [[ "$ACTION" == "uninstall" ]]; then
        do_uninstall
        exit 0
    fi

    if [[ "$DEV_MODE" == true ]]; then
        detect_platform
        prompt_scope
        resolve_paths
        ensure_dev_prerequisites
        ensure_docker
        build_from_source
        install_binary
        create_dirs
        install_service
        print_next_steps
    else
        detect_platform
        prompt_scope
        resolve_paths
        resolve_version
        ensure_docker
        download_binary
        install_binary
        create_dirs
        install_service
        print_next_steps

        rm -f "$DOWNLOADED_BINARY"
    fi
}

main
