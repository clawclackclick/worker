#!/bin/bash
# Setup script for ClawClack SHKeeper VPS
# Run this on your SHKEEPER server (separate from bot!)

set -e

echo "💰 Setting up ClawClack SHKeeper VPS..."

# Update system
apt-get update
apt-get upgrade -y

# Install Docker
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    usermod -aG docker $USER
    systemctl enable docker
    systemctl start docker
fi

# Install Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "🐳 Installing Docker Compose..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

# Create app directory
mkdir -p /opt/shkeeper
chown $USER:$USER /opt/shkeeper

# Setup firewall - RESTRICTIVE (this is the secure server!)
echo "🔒 Configuring firewall (RESTRICTIVE)..."
apt-get install -y ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
# Only allow HTTP/HTTPS from VPN or specific IPs
# ufw allow from YOUR_BOT_VPN_IP to any port 80,443
ufw --force enable

# Install WireGuard
apt-get install -y wireguard wireguard-tools

# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf

# Setup automatic backups
echo "💾 Setting up automatic backups..."
mkdir -p /opt/shkeeper/backups
cat > /opt/shkeeper/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/shkeeper/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup database
docker exec shkeeper-postgres pg_dump -U shkeeper shkeeper > "$BACKUP_DIR/shkeeper_db_$DATE.sql"

# Backup SHKeeper data
tar -czf "$BACKUP_DIR/shkeeper_data_$DATE.tar.gz" /opt/shkeeper/shkeeper_data

# Keep only last 7 days of backups
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
EOF
chmod +x /opt/shkeeper/backup.sh

# Add to crontab (daily at 3 AM)
(crontab -l 2>/dev/null; echo "0 3 * * * /opt/shkeeper/backup.sh") | crontab -

echo "✅ SHKeeper VPS setup complete!"
echo ""
echo "⚠️  CRITICAL SECURITY STEPS:"
echo ""
echo "1. Create .env file in /opt/shkeeper/ with your secrets:"
echo "   - Wallet private keys (generate new ones!)"
echo "   - Strong API key (openssl rand -hex 32)"
echo "   - Database password"
echo ""
echo "2. Configure WireGuard VPN:"
echo "   - Generate keypair"
echo "   - Exchange public keys with bot server"
echo "   - Set internal IPs (e.g., bot=10.0.0.1, shkeeper=10.0.0.2)"
echo ""
echo "3. Fund wallets:"
echo "   - Send small test amounts first!"
echo "   - TRON wallet needs TRX for gas"
echo "   - Polygon wallet needs MATIC for gas"
echo ""
echo "4. Test thoroughly before production!"
