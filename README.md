![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)
[![Go Reference](https://pkg.go.dev/badge/github.com/cnaize/meds.svg)](https://pkg.go.dev/github.com/cnaize/meds)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Platform](https://img.shields.io/badge/platform-linux-blue)
![Version](https://img.shields.io/badge/version-v0.5.0-blue)
![Status](https://img.shields.io/badge/status-stable-success)
[![Go Report Card](https://goreportcard.com/badge/github.com/cnaize/meds)](https://goreportcard.com/report/github.com/cnaize/meds)

---

# Meds: net healing  
> A modern, lock-free firewall powered by NFQUEUE and Go

It integrates with Linux Netfilter via **NFQUEUE**, inspects inbound traffic in user space, and applies filtering to block malicious or unwanted traffic in real-time.

*Meds — "net healing" firewall designed to cure your network from malicious traffic.*

---

## 🚀 Installation

**Requirements:**
- Linux with **iptables** + **NFQUEUE** support
- **Root privileges** (`sudo`) — required for interacting with iptables/NFQUEUE  

The application manages iptables rules automatically.

### Download

Download the latest binary from [Releases](https://github.com/cnaize/meds/releases) or build from sources.

### Build from sources

```bash
go build -o meds ./cmd/daemon
```

## 🧩 Quickstart

```bash
sudo MEDS_USERNAME=admin MEDS_PASSWORD=mypass ./meds
# Metrics available at: http://localhost:8000/metrics
# API available at: http://localhost:8000/swagger/index.html
# Basic Auth: admin / mypass
```

### Command-line options
```text
./meds -help
Usage of ./meds:
  -api-addr string
        api server address (default ":8000")
  -db-path string
        path to database file (default "meds.db")
  -log-level string
        zerolog level (default "info")
  -loggers-count uint
        logger workers count (default 12)
  -max-packets-at-once uint
        max packets per ip at once (default 2000)
  -max-packets-cache-size uint
        max packets per ip cache size (default 10000)
  -max-packets-cache-ttl duration
        max packets per ip cache ttl (default 3m0s)
  -max-packets-per-second uint
        max packets per ip per second (default 100)
  -update-interval duration
        update frequency (default 12h0m0s)
  -update-timeout duration
        update timeout (per filter) (default 10s)
  -workers-count uint
        nfqueue workers count (default 12)
```

### Prometheus metrics  
👉 http://localhost:8000/metrics  

The metrics endpoint is protected by the same **BasicAuth** credentials as the API.

### Swagger UI

Meds provides a **built-in** REST API for managing IPs/Domains whitelist and blacklist.  
It’s automatically served when you run Meds. Protected by the same **BasicAuth** credentials as the API.  

**Interactive API docs:**  
👉 http://localhost:8000/swagger/index.html

You can browse and test all API endpoints directly from your browser.  

**OpenAPI spec (JSON):**  
👉 http://localhost:8000/swagger/doc.json

You can import this spec into Postman, Insomnia, or Hoppscotch.  

---

## ✨ Key Features

- **NFQUEUE-based packet interception**  
  Uses Linux Netfilter queues to copy inbound packets into user space with minimal overhead.

- **Fast packet parsing with [gopacket](https://github.com/google/gopacket)**  
  Parses traffic efficiently (`lazy` and `no copy` modes enabled).

- **Lock-free core**  
  Meds itself does not use any mutexes — all filtering, counters, and rate-limiters use atomic operations.  

- **Blacklist-based filtering**  
  - IP blacklists: [FireHOL](https://iplists.firehol.org/), [Spamhaus DROP](https://www.spamhaus.org/drop/), [Abuse.ch](https://abuse.ch/)  
  - Domain blacklists: [StevenBlack hosts](https://github.com/StevenBlack/hosts/), [SomeoneWhoCares hosts](https://someonewhocares.org/hosts/)

- **TLS SNI & JA3 filtering**  
  Extracts and inspects TLS ClientHello data directly from TCP payload before handshake completion:
  - Filters by SNI (domain in TLS handshake)  
  - Filters by JA3 fingerprint using the [Abuse.ch SSLBL JA3 database](https://sslbl.abuse.ch/ja3-fingerprints/)

  Enables real-time blocking of malicious TLS clients such as malware beacons, scanners, or C2 frameworks.

- **Rate Limiting per IP**  
  Uses token bucket algorithm to limit burst and sustained traffic per source IP.  
  Protects against high-frequency floods (SYN, DNS, ICMP, or generic packet floods).

- **HTTP API for runtime configuration**  
  Built-in API server (powered by [Gin](https://github.com/gin-gonic/gin)) allows dynamically adding or removing IP/Domain entries in global whitelists/blacklists.  
  Auth via BasicAuth using `MEDS_USERNAME` / `MEDS_PASSWORD`.

- **Prometheus metrics export**  
  Exposes metrics for observability:
  - Total packets processed
  - Dropped packets (with reasons)
  - Accepted packets (with reasons)

  Metrics are available at `/metrics` via the built-in API server, compatible with Prometheus scrape targets.
 
- **Asynchronous logging**  
  Uses [zerolog](https://github.com/rs/zerolog) with worker-based async logging for minimal overhead.

- **Efficient lookups**  
  Uses [radix tree](https://github.com/armon/go-radix) and [bart](https://github.com/gaissmai/bart) for IP/domain matching at scale.

- **Extensible design**  
  Modular architecture allows adding new filters (GeoIP, ASN, etc).

---

## 🔍 How It Works
```text
[Kernel] → [NFQUEUE] → [Meds]
                     ↳ Global Whitelist (IP / Domain)
                     ↳ Rate Limiter
                     ↳ IP / Domain Filter
                     ↳ TLS Filter (SNI / JA3)
                     ↳ Global Blacklist (IP / Domain)
                     ↳ Decision: ACCEPT / DROP
```

1. **Packet interception**  
   All inbound packets are queued from Netfilter (`iptables` rule with `-j NFQUEUE`).

2. **Classification pipeline**  
   Packets go through multiple filters:
   - Global IP/Domain whitelist check  
   - Rate Limiting per source IP  
   - IP/Domain check
   - SNI/JA3 check
   - Global IP/Domain blacklist check  

3. **Decision engine**  
   - **ACCEPT** → packet is safe, passed to kernel stack  
   - **DROP** → packet is malicious, discarded immediately  

4. **Metrics & logging**  
   Every decision is counted and exported for monitoring and alerting.  
   Metrics are Prometheus-compatible and can be visualized in Grafana.    
   All events are asynchronously logged to minimize packet processing latency.  

---

## 📊 Example Metrics (Prometheus)

```text
# Total number of accepted packets
meds_core_packets_accepted_total{filter="empty",reason="default"} 2165
meds_core_packets_accepted_total{filter="ip",reason="whitelisted"} 102

# Total number of dropped packets
meds_core_packets_dropped_total{filter="domain",reason="StevenBlack"} 3
meds_core_packets_dropped_total{filter="ip",reason="FireHOL"} 167

# Total number of processed packets
meds_core_packets_processed_total 2437
```

---

## 📜 License

Meds is released under the **MIT License**.  
See [LICENSE](./LICENSE) for details.

---

## 🤝 Contributing

Pull requests and feature suggestions are welcome!  
If you find a bug, please open an issue or submit a fix.

---

Made with ❤️ in Go
