#!/bin/bash

# Production Setup Script for Hetzner Server
# This script sets up the production environment for Zitadel + Vault

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Production Setup - Hetzner${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "Please run as root (use sudo)"
        exit 1
    fi
}

# Update system
update_system() {
    print_info "Updating system packages..."
    apt update && apt upgrade -y
    print_success "System updated"
}

# Install Docker
install_docker() {
    print_info "Installing Docker..."
    
    # Remove old versions
    apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
    
    # Install dependencies
    apt install -y apt-transport-https ca-certificates curl gnupg lsb-release
    
    # Add Docker GPG key
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # Add Docker repository
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # Install Docker
    apt update
    apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    
    # Start and enable Docker
    systemctl start docker
    systemctl enable docker
    
    print_success "Docker installed"
}

# Install Docker Compose
install_docker_compose() {
    print_info "Installing Docker Compose..."
    
    # Download latest version
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    # Make executable
    chmod +x /usr/local/bin/docker-compose
    
    # Create symlink
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    print_success "Docker Compose installed"
}

# Install additional tools
install_tools() {
    print_info "Installing additional tools..."
    
    apt install -y \
        curl \
        wget \
        git \
        unzip \
        jq \
        htop \
        nginx \
        certbot \
        python3-certbot-nginx \
        ufw \
        fail2ban
    
    print_success "Additional tools installed"
}

# Configure firewall
configure_firewall() {
    print_info "Configuring firewall..."
    
    # Reset UFW
    ufw --force reset
    
    # Default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH
    ufw allow ssh
    
    # Allow HTTP/HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow Docker Swarm (if needed)
    # ufw allow 2376/tcp
    # ufw allow 2377/tcp
    # ufw allow 7946/tcp
    # ufw allow 7946/udp
    # ufw allow 4789/udp
    
    # Enable firewall
    ufw --force enable
    
    print_success "Firewall configured"
}

# Configure fail2ban
configure_fail2ban() {
    print_info "Configuring fail2ban..."
    
    cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
EOF
    
    systemctl enable fail2ban
    systemctl start fail2ban
    
    print_success "Fail2ban configured"
}

# Create project directory
create_project_directory() {
    print_info "Creating project directory..."
    
    PROJECT_DIR="/opt/zitadel-saas"
    mkdir -p $PROJECT_DIR
    cd $PROJECT_DIR
    
    # Create subdirectories
    mkdir -p {config,data,logs,backups,monitoring}
    mkdir -p data/{postgres,vault,zitadel,redis,grafana,prometheus}
    mkdir -p config/{nginx,vault,monitoring}
    
    print_success "Project directory created: $PROJECT_DIR"
}

# Setup environment file
setup_environment() {
    print_info "Setting up environment file..."
    
    if [ ! -f .env.production ]; then
        print_warning "Creating .env.production from example"
        cp .env.production.example .env.production
        
        # Generate secure passwords
        POSTGRES_PASSWORD=$(openssl rand -base64 32)
        ZITADEL_DB_PASSWORD=$(openssl rand -base64 32)
        ZITADEL_MASTERKEY=$(openssl rand -base64 32 | head -c 32)
        ZITADEL_ADMIN_PASSWORD=$(openssl rand -base64 16)
        VAULT_ROOT_TOKEN=$(openssl rand -base64 32)
        REDIS_PASSWORD=$(openssl rand -base64 32)
        GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)
        
        # Update .env.production with generated passwords
        sed -i "s/CHANGE_THIS_SECURE_DB_PASSWORD_123/$POSTGRES_PASSWORD/g" .env.production
        sed -i "s/CHANGE_THIS_ZITADEL_DB_PASSWORD_456/$ZITADEL_DB_PASSWORD/g" .env.production
        sed -i "s/CHANGE_THIS_32_CHARACTER_MASTERKEY_789/$ZITADEL_MASTERKEY/g" .env.production
        sed -i "s/CHANGE_THIS_ADMIN_PASSWORD_ABC/$ZITADEL_ADMIN_PASSWORD/g" .env.production
        sed -i "s/CHANGE_THIS_VAULT_ROOT_TOKEN_XYZ/$VAULT_ROOT_TOKEN/g" .env.production
        sed -i "s/CHANGE_THIS_REDIS_PASSWORD_DEF/$REDIS_PASSWORD/g" .env.production
        sed -i "s/CHANGE_THIS_GRAFANA_PASSWORD_GHI/$GRAFANA_ADMIN_PASSWORD/g" .env.production
        
        print_warning "Please update .env.production with your domain names and other settings"
        print_warning "Generated passwords have been set automatically"
    fi
    
    print_success "Environment file ready"
}

# Setup monitoring
setup_monitoring() {
    print_info "Setting up monitoring configuration..."
    
    # Prometheus configuration
    cat > config/monitoring/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'zitadel'
    static_configs:
      - targets: ['zitadel:8080']
    metrics_path: '/debug/metrics'

  - job_name: 'vault'
    static_configs:
      - targets: ['vault:8200']
    metrics_path: '/v1/sys/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']

  - job_name: 'traefik'
    static_configs:
      - targets: ['traefik:8080']
    metrics_path: '/metrics'
EOF

    # Grafana provisioning
    mkdir -p config/monitoring/grafana/provisioning/{dashboards,datasources}
    
    cat > config/monitoring/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    print_success "Monitoring configuration ready"
}

# Setup backup script
setup_backup() {
    print_info "Setting up backup script..."
    
    cat > /opt/zitadel-saas/backup.sh << 'EOF'
#!/bin/bash

# Backup script for Zitadel SaaS production

BACKUP_DIR="/opt/zitadel-saas/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR/$DATE

# Backup PostgreSQL
docker exec postgres-production pg_dump -U postgres zitadel > $BACKUP_DIR/$DATE/postgres_backup.sql

# Backup Vault data
docker exec vault-production vault operator raft snapshot save /vault/data/vault_snapshot_$DATE.snap
docker cp vault-production:/vault/data/vault_snapshot_$DATE.snap $BACKUP_DIR/$DATE/

# Backup Zitadel data
docker cp zitadel-production:/zitadel-data $BACKUP_DIR/$DATE/zitadel-data

# Backup configuration
cp -r /opt/zitadel-saas/config $BACKUP_DIR/$DATE/
cp /opt/zitadel-saas/.env.production $BACKUP_DIR/$DATE/

# Compress backup
tar -czf $BACKUP_DIR/backup_$DATE.tar.gz -C $BACKUP_DIR $DATE
rm -rf $BACKUP_DIR/$DATE

# Keep only last 7 days of backups
find $BACKUP_DIR -name "backup_*.tar.gz" -mtime +7 -delete

echo "Backup completed: backup_$DATE.tar.gz"
EOF

    chmod +x /opt/zitadel-saas/backup.sh
    
    # Add to crontab (daily backup at 2 AM)
    (crontab -l 2>/dev/null; echo "0 2 * * * /opt/zitadel-saas/backup.sh") | crontab -
    
    print_success "Backup script configured"
}

# Main setup function
main() {
    print_header
    
    print_info "Starting production setup on Hetzner server..."
    
    check_root
    update_system
    install_docker
    install_docker_compose
    install_tools
    configure_firewall
    configure_fail2ban
    create_project_directory
    setup_environment
    setup_monitoring
    setup_backup
    
    print_success "Production setup completed!"
    
    echo ""
    print_info "Next steps:"
    echo "1. Update /opt/zitadel-saas/.env.production with your domain names"
    echo "2. Copy your docker-compose.production.yml to /opt/zitadel-saas/"
    echo "3. Run: cd /opt/zitadel-saas && docker-compose -f docker-compose.production.yml up -d"
    echo "4. Configure DNS records to point to this server"
    echo "5. Run Terraform to create SaaS organizations"
    echo ""
    print_warning "Don't forget to secure your .env.production file!"
    echo "chmod 600 /opt/zitadel-saas/.env.production"
}

main "$@"