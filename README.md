![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)
[![Go Reference](https://pkg.go.dev/badge/github.com/cnaize/meds.svg)](https://pkg.go.dev/github.com/cnaize/meds)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Platform](https://img.shields.io/badge/platform-linux-blue)
![Version](https://img.shields.io/badge/version-v1.1.0-blue)
![Status](https://img.shields.io/badge/status-stable-success)
[![Go Report Card](https://goreportcard.com/badge/github.com/cnaize/meds)](https://goreportcard.com/report/github.com/cnaize/meds)

---

# Meds: net healing  
> Intelligent firewall in Go

It integrates with Linux Netfilter via **NFQUEUE**, inspects inbound traffic in user space, and applies filtering to block malicious traffic in real-time. Once a connection is checked, the engine "teaches" the Linux kernel to handle it. By assigning **Conntrack marks**, Meds offloads flows back to the kernel space, achieving maximum wire-speed throughput and minimal CPU overhead.

*Designed to cure your network from malicious traffic*

---

## 🚀 Installation

**Requirements:**
- Linux with **iptables** + **NFQUEUE** + **conntrack** support
- **Root privileges** (`sudo`) — required for interacting with Netfilter/Netlink

The application manages iptables and conntrack rules automatically.

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
  -logger-queue-len uint
    	logger queue length (all workers) (default 2048)
  -loggers-count uint
    	logger workers count (default 3)
  -rate-limiter-burst uint
    	max packets at once (per ip) (default 1500)
  -rate-limiter-cache-size uint
    	rate limiter cache size (all buckets) (default 100000)
  -rate-limiter-cache-ttl duration
    	rate limiter cache ttl (per bucket) (default 5m0s)
  -rate-limiter-rate uint
    	max packets per second (per ip) (default 3000)
  -reader-queue-len uint
    	nfqueue queue length (per reader) (default 8192)
  -readers-count uint
    	nfqueue readers count (default 12)
  -update-interval duration
    	update frequency (default 4h0m0s)
  -update-timeout duration
    	update timeout (per filter) (default 1m0s)
  -workers-count uint
    	nfqueue workers count (per reader) (default 1)
```

### Prometheus metrics  
👉 http://localhost:8000/metrics  

The metrics endpoint is protected by the same **BasicAuth** credentials as the API.

### Swagger UI

**Interactive API docs:**  
👉 http://localhost:8000/swagger/index.html

You can browse and test all API endpoints directly from your browser.  

**OpenAPI spec (JSON):**  
👉 http://localhost:8000/swagger/doc.json

You can import this spec into Postman, Insomnia, or Hoppscotch.  

---

## 🔍 How It Works
```text
                         PACKET
                           │
┌──────────────────────────▼──────────────────────────┐
│ KERNEL SPACE (iptables / Netfilter)                 │
│ ─────────────────────────────────────────────────── │
│ [1. Restore Connmark]             ◄─────────────────┼───────┐
│                                                     │       │
│ [2. White / Block Check]          ──► ACCEPT ───────┼───┐   │
│      (Marks: 0x100000 / 0x200000) ──► DROP          │   │   │
│                                                     │   │   │
│ [3. Rate Limiter] (per source IP) ──► DROP if flood │   │   │
│                                                     │   │   │
│ [4. Trusted Check]                ──► ACCEPT ───────┼───┤   │
│      (Mark: 0x400000)                               │   │   │
│                                                     │   │   │
│ [5. Unclassified] (NFQUEUE)       ──► to User Space │   │   │
│      (First 10 pkts / Balance 0:N)                  │   │   │
└──────────────────────────────────────────┼──────────┘   │   │
                                           │              │   │
┌──────────────────────────────────────────▼──┐           │   │
│ USER SPACE (Meds Firewall)                  │           │   │
│ ─────────────────────────────────────────── │           │   │
│   1. L3/L4 Filters (IP, Geo, ASN)           │           │   │
│   2. L7 Inspection (DNS, SNI, TLS JA3)      │           │   │
│                                             │           │   │
│ [DECISION ENGINE]                           │           │   │
│   * Set Verdict     (DROP / ACCEPT) ────────┼───────────┤   │
│   * Update Connmark (White / Block / Trust) ┼──► [MARK] ────┘
└─────────────────────────────────────────────┘           │
                                                          ▼
                                                   TRAFFIC ALLOWED
```

- **Pre-Limit Bypass**: White Listed (`0x100000`) and Block Listed (`0x200000`) traffic is handled by the kernel immediately. This ensures that verified legitimate traffic has zero overhead from rate limiters, while known threats are dropped at the earliest possible stage.

- **Global Protection**: All unclassified or non-whitelisted traffic is subject to a kernel-level `hashlimit` (PPS per source IP). This acts as a primary shield against volumetric DDoS attacks, protecting the User Space engine from exhaustion.

- **Stateful Acceleration**: Once a connection is verified by Meds as Trusted (`0x400000`), it is offloaded to the kernel's fast path. Subsequent packets bypass deep inspection while remaining under the protection of the rate limiter.

- **Deep Inspection**: Only new or unclassified traffic (the "Decision Phase", limited to the **first 10 packets** via `connbytes`) is sent to the Go engine for deep L3/L4/L7 DPI analysis.

---

## ✨ Key Features

- **Hybrid Kernel/User-space Processing**  
  Meds utilizes a stateful marking architecture. It "teaches" the Linux kernel how to handle specific flows by assigning **Conntrack marks**, achieving wire-speed performance for established connections.

- **Intelligent NFQUEUE Balancing**  
  Intercepts traffic using `NFQUEUE` with `fanout` and `bypass` options, ensuring multi-core scaling and system stability even if the user-space process is restarted.

- **Lock-free Core Architecture**  
  The core engine is built for high-concurrency performance: no mutexes in the hot path. All filtering, counters, and rate-limiters utilize atomic operations.

- **Multi-layer Rate Limiting**  
  Combines **Kernel-level protection** (fast `hashlimit` PPS limiting) with **User-space logic** (token bucket) for sophisticated traffic shaping and flood protection.

- **Blacklist-based filtering**  
  - IP blacklists: [FireHOL](https://iplists.firehol.org/), [Spamhaus DROP](https://www.spamhaus.org/drop/), [Abuse.ch](https://abuse.ch/)
  - ASN blacklists: [Spamhaus ASN DROP](https://www.spamhaus.org/drop/asndrop.json) using [IPLocate.io](https://iplocate.io/) for IP-to-ASN mapping
  - Domain blacklists: [StevenBlack hosts](https://github.com/StevenBlack/hosts/), [SomeoneWhoCares hosts](https://someonewhocares.org/hosts/)

- **Geo-blocking (ASN-based)**  
  Efficiently blocks traffic from specific countries using ASN metadata from [IPLocate.io](https://iplocate.io/):  
  - Lightweight alternative to heavy GeoIP databases
  - Dynamic configuration via API/Swagger

- **TLS SNI & JA3 filtering**  
  Extracts and inspects TLS ClientHello data directly from TCP payload before handshake completion:
  - Filters by SNI (domain in TLS handshake)  
  - Filters by JA3 fingerprint using the [Abuse.ch SSLBL JA3 database](https://sslbl.abuse.ch/ja3-fingerprints/)

  Enables real-time blocking of malicious TLS clients such as malware beacons, scanners, or C2 frameworks.

- **HTTP API for runtime configuration**  
  Built-in API server (powered by [Gin](https://github.com/gin-gonic/gin)) allows dynamically adding or removing IP or Country entries in global white/black lists.  
  Auth via BasicAuth using `MEDS_USERNAME` / `MEDS_PASSWORD`.

- **Prometheus metrics export**  
  Exposes metrics for observability:
  - Total packets processed
  - Dropped packets (with reasons)
  - Accepted packets (with reasons)
  - Internal errors (with types)

  Metrics are available at `/metrics` via the built-in API server, compatible with Prometheus scrape targets.

---

## 📊 Example Metrics (Prometheus)

```text
# HELP meds_core_packets_accepted_total Total number of accepted packets
# TYPE meds_core_packets_accepted_total counter
meds_core_packets_accepted_total{filter="empty",reason="default"} 12021
meds_core_packets_accepted_total{filter="empty",reason="trusted packet"} 420
meds_core_packets_accepted_total{filter="ip",reason="WhiteList"} 139

# HELP meds_core_packets_dropped_total Total number of dropped packets
# TYPE meds_core_packets_dropped_total counter
meds_core_packets_dropped_total{filter="asn",reason="Spamhaus"} 263
meds_core_packets_dropped_total{filter="domain",reason="StevenBlack"} 3
meds_core_packets_dropped_total{filter="geo",reason="IPLocate"} 43
meds_core_packets_dropped_total{filter="ip",reason="FireHOL"} 1443

# HELP meds_core_packets_processed_total Total number of processed packets
# TYPE meds_core_packets_processed_total counter
meds_core_packets_processed_total 14332
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
