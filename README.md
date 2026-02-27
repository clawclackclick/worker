# ClawClack Agent

Autonomous AI agent for Matrix that earns money offering services and spends money to grow.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  USER (Element/FluffyChat/any Matrix client)                │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│  matrix.org (free homeserver)                               │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│  BOT VPS (bot.clawclack.click)                              │
│  • Go Matrix bot                                            │
│  • AI decision engine                                       │
│  • Connects to SHKeeper via VPN                             │
└──────────────┬──────────────────────────────────────────────┘
               │ WireGuard VPN (secure)
               ▼
┌─────────────────────────────────────────────────────────────┐
│  SHKEEPER VPS (shkeeper.clawclack.click)                    │
│  • USDT/USDC wallets (TRON + Polygon)                       │
│  • Payment processing                                       │
│  • NO public internet access (VPN only)                     │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Fork & Clone

```bash
git clone https://github.com/YOUR_USERNAME/clawclack.git
cd clawclack
```

### 2. Setup GitHub Secrets

Go to Settings → Secrets and variables → Actions, add:

| Secret | Description | Get From |
|--------|-------------|----------|
| `MATRIX_HOMESERVER` | `https://matrix.org` | matrix.org |
| `MATRIX_USER_ID` | `@yourbot:matrix.org` | After registration |
| `MATRIX_ACCESS_TOKEN` | Bot access token | Element settings |
| `MATRIX_DEVICE_ID` | Device ID | Element settings |
| `SHKEEPER_URL` | `http://10.0.0.2:5000` | Internal VPN IP |
| `SHKEEPER_API_KEY` | API key | SHKeeper dashboard |
| `VPS_BOT_HOST` | `bot.clawclack.click` | Your domain |
| `VPS_BOT_USER` | `root` or deploy user | Your VPS |
| `VPS_BOT_SSH_KEY` | SSH private key | Generate new |
| `VPS_SHKEEPER_HOST` | `shkeeper.clawclack.click` | Your domain |
| `VPS_SHKEEPER_USER` | `root` or deploy user | Your VPS |
| `VPS_SHKEEPER_SSH_KEY` | SSH private key | Generate new |
| `SPENDING_LIMIT_USD` | `100` | Your choice |
| `DAILY_BUDGET_USD` | `500` | Your choice |
| `OPENAI_API_KEY` | OpenAI API key | openai.com |

### 3. Buy Domain & VPS

**Domain:** clawclack.click (or your choice)

**VPS 1 - Bot Server:**
- 1 vCPU, 2GB RAM, 20GB SSD
- Ubuntu 22.04 LTS
- Example: Hetzner CX11 (€4.51/mo), DigitalOcean $6/mo

**VPS 2 - SHKeeper Server:**
- 1 vCPU, 2GB RAM, 40GB SSD
- Ubuntu 22.04 LTS
- Same provider, different region for redundancy

### 4. Configure DNS

```
A     clawclack.click         → BOT_VPS_IP
A     bot.clawclack.click     → BOT_VPS_IP
A     shkeeper.clawclack.click → SHKEEPER_VPS_IP
```

### 5. Deploy

```bash
# Bot deploys automatically on push to main
git push origin main

# Or manually:
make deploy-bot
make deploy-shkeeper
```

## Bot Commands

| Command | Description | Cost |
|---------|-------------|------|
| `!help` | Show all commands | Free |
| `!balance` | Show agent treasury | Free |
| `!services` | List available services | Free |
| `!price <crypto>` | Get crypto price | Free |
| `!alert <crypto> <price>` | Set price alert | $0.50 |
| `!summarize <url>` | Summarize article | $1.00 |
| `!image <prompt>` | Generate AI image | $2.00 |
| `!code <description>` | Generate code snippet | $3.00 |
| `!propose <idea>` | Agent proposes service | Variable |

## Agent Autonomy Rules

The AI agent can:
- ✅ Spend up to $1 per transaction without approval
- ✅ Spend up to $5 per day total
- ✅ Offer services priced $0.50 - $1.00
- ✅ Invest earnings in marketing (ads, promotions)
- ✅ Hire other bots/services to complete tasks

The AI agent cannot:
- ❌ Spend over $1 without manual approval
- ❌ Exceed $5 daily budget
- ❌ Withdraw funds to external wallets
- ❌ Change spending limits

## Directory Structure

```
├── bot/              # Go Matrix bot
├── shkeeper/         # SHKeeper Docker setup
├── agent/            # AI prompts & logic
├── scripts/          # Deployment scripts
└── .github/          # CI/CD workflows
```

## License

MIT - See LICENSE file
