# BLUEM  
### Zero-Downtime Blue-Green Deployment Tool for Dockerized Applications  
**Version: v0.1.0**

BLUEM is a fast, lightweight, Go-based **blue-green deployment engine** designed for Dockerized applications. It provides zero-downtime deployments, safe rollbacks, health-based failover, and pluggable proxy backends such as **Traefik** and **Nginx**.

BLUEM makes deployments predictable, reproducible, and fully automated — without requiring Kubernetes or complex CI/CD tools.

---

## ✨ Features

- 🚀 **Zero-downtime deployments** via BLUE/GREEN switching  
- 🛡 **Automatic rollback** when new version is unhealthy  
- 🔄 **Traffic switching** handled through Traefik or Nginx  
- 🧠 **Pluggable proxy system** with ability to add custom providers  
- 🐳 Built on Docker Compose (no extra dependencies)  
- ⚙ Configurable via a simple YAML file (`bluem.yml`)  
- 🔧 Customizable templates for compose & proxy configs  
- 🛑 **Stop command** to shut down entire stack (blue, green, proxy)  
- 🧪 Includes unit tests  
- 🖥 Prebuilt binaries for macOS, Linux, and Windows  

---

## 📦 Installation

### Option 1 — Download Prebuilt Binaries

Available under:

```
binary/
  macos/arm64/bluem
  macos/amd64/bluem
  linux/arm64/bluem
  linux/amd64/bluem
  windows/arm64/bluem.exe
  windows/amd64/bluem.exe
```

Example installation (macOS ARM):

```bash
sudo cp binary/macos/arm64/bluem /usr/local/bin/
chmod +x /usr/local/bin/bluem
```

---

### Option 2 — Build from source

```bash
git clone https://github.com/pkshahid/bluem
cd bluem
go build -o bluem ./cmd/bluem
```

---

## 🚀 Quick Start

### 1. Copy example config

```bash
cp bluem.yml.example bluem.yml
```

### 2. Initialize template files

```bash
bluem init
```

This generates:

- docker-compose.blue.yml  
- docker-compose.green.yml  
- docker-compose.<proxy>.yml  
- Traefik/Nginx proxy configs  

### 3. Start GREEN environment

```bash
bluem up
```

### 4. Deploy new release

```bash
bluem deploy
```

This:

- Starts BLUE  
- Performs health checks  
- Switches traffic  
- Drains & stops GREEN  
- Enables auto rollback  

### 5. Rollback manually (if needed)

```bash
bluem rollback
```

### 6. Stop everything

```bash
bluem stop
```

---

## ⚙ Configuration (`bluem.yml`)

A complete example:

```yaml
project_name: Demo Shop
project_slug: myshop

registry: "myshop"
image: "web"
tag: "latest"

app_port: 8000
green_port: 8010
blue_port: 8011
health_path: /health/

network_name: myshop-net
network_external: false

proxy:
  type: traefik
  port: 8085

services:
  django:
    enabled: true
    env_file: .env

failover_timeout: "120s"
container_draining_timeout: "20s"
health_check_delay: "5s"
```

### Important fields

| Field | Description |
|-------|-------------|
| `project_slug` | Used for container names (`myshop_blue`, `myshop_green`) |
| `health_path` | Must return HTTP 200 for container to be considered healthy |
| `failover_timeout` | Time BLUE must remain healthy to complete deployment |
| `container_draining_timeout` | Delay before stopping GREEN |
| `health_check_delay` | Interval for checking BLUE container's health |

All durations use Go format (`5s`, `2m`, `1h`).

---

## 🧠 Deployment Flow

1. Build/push new image  
2. Run `bluem deploy`  
3. BLUE container starts  
4. BLUEM waits until BLUE is **healthy**  
5. Proxy switches traffic → BLUE  
6. BLUE is monitored for a failover window  
7. GREEN is drained + shut down  
8. Deployment completes  

### Auto rollback occurs when:

- BLUE container health fails  
- Proxy cannot connect to BLUE  
- Failover timeout expires prematurely  

BLUEM will immediately:

- switch traffic → GREEN  
- stop BLUE  

---

## 🧩 Commands

### `init` — Generate configuration templates

```bash
bluem init
```

### `up` — Start GREEN + proxy

```bash
bluem up
```

### `deploy` — Perform blue-green deployment

```bash
bluem deploy
```

### `rollback` — Switch back to GREEN

```bash
bluem rollback
```

### `status` — Show active stack & health

```bash
bluem status
```

### `stop` — Stop blue, green, and proxy containers

```bash
bluem stop
```

---

## 🔌 Proxy Support

BLUEM supports two battle-tested proxy systems:

### Traefik (default)
- Auto-reload via file provider  
- Simple traffic switching  
- Dynamic config generated automatically  

### Nginx
- Template-based upstream switching  
- Supports custom Nginx configs  

### Custom Proxies
Implement a Go interface:

```go
type Proxy interface {
    Name() string
    Init(...)
    Up(...)
    SwitchToBlue(...)
    SwitchToGreen(...)
    Stop(...)
}
```

Drop it into `internal/proxy/` to extend BLUEM.

---

## 🧪 Testing

BLUEM includes tests for:

- Template rendering  
- Proxy interface switching  
- Config loader correctness  

Run tests:

```bash
go test ./...
```

---

## 📁 Project Structure

```
bluem/
  cmd/bluem/
  internal/
    cli/
    config/
    deploy/
    docker/
    proxy/
    template/
  bluem.yml.example
  binary/
  README.md
```

---

## 🧱 Example Architecture

```
               ┌──────────────┐
               │   BLUE APP   │
               └──────┬───────┘
                      │
        User ─▶ Proxy │────────▶ traffic
                      │
               ┌──────┴───────┐
               │   GREEN APP  │
               └──────────────┘
```

Proxy chooses BLUE or GREEN based on BLUEM’s deployment logic.

---

## 🛣 Roadmap

### v0.2.0
- Remote Docker host support  
- Multiple service deployments  
- More proxy backends  

### v0.3.0
- Canary deployments  

### v0.4.0  
- Kubernetes support  

### v0.5.0  
- Multi-application orchestration  

---

## 📜 License

Released under the **MIT License**.  
See [`LICENSE`](LICENSE) for details.

---

## ❤️ Contributing

Contributions welcome!

1. Fork repo  
2. Create feature branch  
3. Add tests  
4. Submit PR  

---

## 🙌 Final Notes

BLUEM v0.1.0 is a production-ready solution for zero-downtime deployments using Docker. It is ideal for small to medium applications that want the safety of blue-green deployment without heavy systems like Kubernetes, Consul, or Spinnaker.
