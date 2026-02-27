# ClawClack Setup Guide

Complete step-by-step guide to deploy your autonomous AI agent.

## Prerequisites

1. **Domain**: clawclack.click (purchased)
2. **Two VPS servers**:
   - Bot VPS: 1 vCPU, 2GB RAM (~$5-10/month)
   - SHKeeper VPS: 1 vCPU, 2GB RAM (~$5-10/month)
3. **GitHub account** with this repository forked
4. **Matrix account**: @yourbot:matrix.org
5. **OpenAI API key** (for AI features)

## Step 1: Domain Setup

### 1.1 Configure DNS

In your domain registrar's DNS settings:

```
Type  Host                    Value                  TTL
A     clawclack.click         BOT_VPS_IP            3600
A     bot.clawclack.click     BOT_VPS_IP            3600
A     shkeeper.clawclack.click SHKEEPER_VPS_IP       3600
```

### 1.2 Get VPS IPs

After creating VPS instances, note their public IPs:
- Bot VPS IP: `xxx.xxx.xxx.xxx`
- SHKeeper VPS IP: `yyy.yyy.yxx.yyy`

## Step 2: VPS Setup

### 2.1 Bot VPS

SSH into your bot VPS and run:

```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/clawclack/main/scripts/setup-bot-vps.sh | bash
```

### 2.2 SHKeeper VPS

SSH into your SHKeeper VPS and run:

```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/clawclack/main/scripts/setup-shkeeper-vps.sh | bash
```

## Step 3: Matrix Bot Account

### 3.1 Create Account

1. Download Element app: https://element.io/download
2. Register at matrix.org
3. Create a new account for your bot (e.g., @clawclack:matrix.org)

### 3.2 Get Access Token

1. Log in to Element with bot account
2. Settings → Help & About → Advanced → Access Token
3. **Copy and save this token!**
4. Also note the Device ID

### 3.3 Test Bot

Join a test room and invite your bot to verify it works.

## Step 4: WireGuard VPN Setup

This creates a secure tunnel between bot and SHKeeper.

### 4.1 On Bot VPS

```bashnwg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey

# Create config
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $(cat /etc/wireguard/privatekey)
Address = 10.0.0.1/24
ListenPort = 51820

[Peer]
PublicKey = SHKEEPER_PUBLIC_KEY  # Get from SHKeeper VPS
AllowedIPs = 10.0.0.2/32
PersistentKeepalive = 25
EOF

chmod 600 /etc/wireguard/wg0.conf
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0
```

### 4.2 On SHKeeper VPS

```bash
wg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey

# Create config
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $(cat /etc/wireguard/privatekey)
Address = 10.0.0.2/24
ListenPort = 51820

[Peer]
PublicKey = BOT_PUBLIC_KEY  # Get from Bot VPS
AllowedIPs = 10.0.0.1/32
PersistentKeepalive = 25
EOF

chmod 600 /etc/wireguard/wg0.conf
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0
```

### 4.3 Exchange Public Keys

```bash
# On Bot VPS
cat /etc/wireguard/publickey

# On SHKeeper VPS
cat /etc/wireguard/publickey

# Add each other's public keys to the [Peer] sections
```

### 4.4 Test Connection

From Bot VPS:
```bash
ping 10.0.0.2  # Should reach SHKeeper
```

From SHKeeper VPS:
```bash
ping 10.0.0.1  # Should reach Bot
```

## Step 5: SHKeeper Configuration

### 5.1 Create Wallets

Generate new wallet addresses:

**TRON (USDT TRC20):**
- Use TronLink or similar
- Generate new wallet
- Save private key
- Fund with small amount of TRX for gas

**Polygon (USDC):**
- Use MetaMask
- Add Polygon network
- Generate new wallet
- Save private key
- Fund with small amount of MATIC for gas

### 5.2 Create .env File

On SHKeeper VPS:

```bash
cd /opt/shkeeper
cp .env.example .env
nano .env  # Edit with your values
```

Fill in:
- `POSTGRES_PASSWORD`: Random strong password
- `SHKEEPER_API_KEY`: `openssl rand -hex 32`
- `TRON_WALLET_PRIVATE_KEY`: Your TRON private key
- `POLYGON_WALLET_PRIVATE_KEY`: Your Polygon private key
- `ACME_EMAIL`: Your email for HTTPS certs

### 5.3 First Deploy

```bash
cd /opt/shkeeper
docker-compose up -d

# Check logs
docker-compose logs -f shkeeper
```

### 5.4 Get API Key

The `SHKEEPER_API_KEY` you set in .env is what the bot will use.

## Step 6: GitHub Secrets

Go to GitHub → Your Repository → Settings → Secrets and variables → Actions

Add these secrets:

### Matrix Secrets
```
MATRIX_HOMESERVER=https://matrix.org
MATRIX_USER_ID=@yourbot:matrix.org
MATRIX_ACCESS_TOKEN=your_access_token_here
MATRIX_DEVICE_ID=your_device_id_here
```

### SHKeeper Secrets
```
SHKEEPER_URL=http://10.0.0.2:5000  # Internal VPN IP
SHKEEPER_API_KEY=your_api_key_from_step_5
```

### Agent Secrets
```
SPENDING_LIMIT_USD=100
DAILY_BUDGET_USD=500
OPENAI_API_KEY=sk-...your_openai_key...
```

### VPS SSH Secrets
```
VPS_BOT_HOST=bot.clawclack.click
VPS_BOT_USER=root  # or your deploy user
VPS_BOT_SSH_KEY=-----BEGIN OPENSSH PRIVATE KEY-----
... (generate new keypair)

VPS_SHKEEPER_HOST=shkeeper.clawclack.click
VPS_SHKEEPER_USER=root
VPS_SHKEEPER_SSH_KEY=-----BEGIN OPENSSH PRIVATE KEY-----
... (generate new keypair)
```

### Generate SSH Keys

```bash
# For bot VPS
ssh-keygen -t ed25519 -C "github-actions-bot" -f bot_deploy_key

# For SHKeeper VPS
ssh-keygen -t ed25519 -C "github-actions-shkeeper" -f shkeeper_deploy_key

# Copy public keys to respective VPS authorized_keys
cat bot_deploy_key.pub | ssh root@bot.clawclack.click "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
cat shkeeper_deploy_key.pub | ssh root@shkeeper.clawclack.click "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"

# Copy private keys to GitHub Secrets
cat bot_deploy_key
# → Paste into VPS_BOT_SSH_KEY secret

cat shkeeper_deploy_key
# → Paste into VPS_SHKEEPER_SSH_KEY secret
```

## Step 7: Deploy

### 7.1 Push to GitHub

```bash
git add .
git commit -m "Initial deployment"
git push origin main
```

### 7.2 Verify Deployment

Check GitHub Actions tab to see deployment status.

### 7.3 Test Bot

In Matrix room:
```
!help
!balance
```

### 7.4 Test Payment

```
!pay 1 USDT
```

Should generate a payment link.

## Step 8: Security Hardening

### 8.1 Disable Root Login (both VPS)

```bash
# Create deploy user
adduser deploy
usermod -aG sudo deploy
usermod -aG docker deploy

# Copy SSH key
mkdir -p /home/deploy/.ssh
cp ~/.ssh/authorized_keys /home/deploy/.ssh/
chown -R deploy:deploy /home/deploy/.ssh

# Edit SSH config
nano /etc/ssh/sshd_config
# Set: PermitRootLogin no
# Set: PasswordAuthentication no

systemctl restart sshd
```

### 8.2 Update GitHub Secrets

Change VPS_BOT_USER and VPS_SHKEEPER_USER to `deploy`

### 8.3 Firewall Rules

On SHKeeper, restrict to VPN only:
```bash
ufw allow from 10.0.0.1 to any port 80,443
ufw delete allow 80
ufw delete allow 443
```

## Step 9: Monitoring

### 9.1 Bot Logs

```bash
ssh deploy@bot.clawclack.click "docker logs -f clawclack-bot"
```

### 9.2 SHKeeper Logs

```bash
ssh deploy@shkeeper.clawclack.click "docker-compose -f /opt/shkeeper/docker-compose.yml logs -f"
```

### 9.3 Health Checks

Set up UptimeRobot or similar to ping:
- https://shkeeper.clawclack.click/health

## Troubleshooting

### Bot won't connect to Matrix
- Check MATRIX_ACCESS_TOKEN is correct
- Verify MATRIX_USER_ID format (@bot:matrix.org)
- Check firewall allows outbound connections

### Bot can't reach SHKeeper
- Verify WireGuard is running: `sudo wg show`
- Test ping: `ping 10.0.0.2`
- Check SHKeeper API key matches

### Payments not detected
- Check SHKeeper is running: `docker-compose ps`
- Verify webhook configuration
- Check SHKeeper logs for errors

### Agent won't spend
- Check spending limits in config
- Verify agent has recorded earnings
- Check agent balance in SHKeeper

## Maintenance

### Update Bot

```bash
git pull origin main
# Make changes
git push origin main
# GitHub Actions auto-deploys
```

### Backup SHKeeper

Backups run automatically daily. To restore:

```bash
# List backups
ls -la /opt/shkeeper/backups/

# Restore database
docker exec -i shkeeper-postgres psql -U shkeeper shkeeper < backup_file.sql
```

### Rotate Keys

1. Generate new wallet keys
2. Transfer funds from old wallets
3. Update .env
4. Redeploy
5. Update GitHub Secrets

## Support

- Matrix protocol: https://spec.matrix.org/
- Mautrix Go: https://docs.mau.fi/mautrix-go/
- SHKeeper: https://github.com/vsys-host/shkeeper.io
- OpenAI API: https://platform.openai.com/docs/
