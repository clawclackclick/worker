#!/bin/bash
# Setup script for ClawClack Bot VPS
# Run this on your BOT server (not SHKeeper)

set -e

echo "🤖 Setting up ClawClack Bot VPS..."

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
mkdir -p /opt/clawclack
chown $USER:$USER /opt/clawclack

# Setup firewall (allow only SSH, HTTP, HTTPS)
echo "🔒 Configuring firewall..."
apt-get install -y ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable

# Install WireGuard (for secure SHKeeper connection)
echo "🔐 Installing WireGuard..."
apt-get install -y wireguard wireguard-tools

# Enable IP forwarding for VPN
sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf

echo "✅ Bot VPS setup complete!"
echo ""
echo "Next steps:"
echo "1. Configure WireGuard VPN to SHKeeper server"
echo "2. Add GitHub Actions SSH keys"
echo "3. Deploy the bot"
echo ""
echo "To configure WireGuard:"
echo "  wg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey"
