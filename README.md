![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go)
[![Go Reference](https://pkg.go.dev/badge/github.com/cnaize/meds.svg)](https://pkg.go.dev/github.com/cnaize/meds)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Platform](https://img.shields.io/badge/platform-linux-blue)
![Version](https://img.shields.io/badge/version-v1.3.0-blue)
![Status](https://img.shields.io/badge/status-stable-success)

---

# Meds
> Hybrid firewall using public blocklists

It integrates with Linux Netfilter via **NFQUEUE**, inspects inbound traffic in user space, and applies filtering based on public blocklists. Once a connection is checked, the engine "teaches" the Linux kernel to handle it. By assigning **Conntrack marks**, Meds offloads flows back to the kernel space, achieving maximum wire-speed throughput and minimal CPU overhead.

---

## 🚀 Installation

**Requirements:**
- Linux with **iptables** + **NFQUEUE** + **conntrack** support
- **Root privileges** (`sudo`) — required for interacting with Netfilter/Netlink

The application manages iptables and conntrack rules automatically.

### Download

Download the latest binary from [Releases](https://github.com/cnaize/meds/releases) or install via Go.

### Install via Go

```bash
go install github.com/cnaize/meds/cmd/meds@latest
```

## 🧩 Quickstart

```bash
sudo MEDS_USERNAME=admin MEDS_PASSWORD=mypass ./meds
```

### Prometheus metrics  
👉 http://localhost:8000/metrics  

### Swagger UI  
👉 http://localhost:8000/swagger/index.html  

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
  -nats-block-ip-cache-size uint
    	nats cache size for block ip (all entities) (default 10000)
  -nats-block-ip-entity-ttl duration
    	nats cache ttl for block ip (per entity) (default 3m0s)
  -nats-enable
    	enable nats messaging
  -nats-host string
    	nats server host (default "localhost")
  -nats-port int
    	nats server port (default 4222)
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

---

## 🔍 How It Works
```text
                  PACKET
                    │
┌───────────────────▼───────────────────┐
│ KERNEL SPACE (iptables / Netfilter)   │
│ ───────────────────────────────────── │
│  1. Restore Connmark                  │
│                                       │
│  2. Check Blocklist   ──► DROP      ◄─┼──┐
│      (Mark: 0x100000)                 │  │
│                                       │  │
│  3. Check Trustlist   ──► ACCEPT      │  │
│      (Mark: 0x200000)                 │  │
│                                       │  │
│  4. First 10 packets  ──┐             │  │
│                         │             │  │
│  5. Save Connmark       │             │  │
└─────────────────────────┼─────────────┘  │
                          │                │
┌─────────────────────────▼─────────────┐  │
│ USER SPACE (Meds Firewall)            │  │
│ ───────────────────────────────────── │  │
│  1. Rate Limiter (per source IP)      │  │
│  2. L3/L4 Filters (IP, Geo, ASN)      │  │
│  3. L7 Inspection (DNS, SNI, TLS JA3) │  │
│                                       │  │
│ [DECISION ENGINE]                     │  │
│  * BLOCK: Mark 0x100000  ──► REPEAT ──┼──┘
│  * TRUST: Mark 0x200000  ──► ACCEPT   │
└─────────────────────────▲─────────────┘
                          │
               ┌──────────┴─────────────┐
               │ EMBEDDED NATS SERVER ◄─┼──[External]
               └────────────────────────┘
```

- **Early Drop**: Blocked traffic (`0x100000`) is handled by the kernel immediately. This ensures that known threats are dropped at the earliest possible stage, eliminating unnecessary context switches and user-space overhead.

- **Stateful Acceleration**: Once a connection is verified as Trusted (`0x200000`), it is offloaded to the kernel's fast path. Subsequent packets in the flow are processed entirely in-kernel at wire-speed, eliminating user-space overhead for established sessions.

- **Deep Inspection**: Only new or unclassified traffic (the "Decision Phase") is sent to Meds for deep L3/L4/L7 analysis. This phase is limited to a **10-packet window** to extract metadata (DNS, SNI, JA3) before the kernel takes over.

- **Reactive Threat Offloading**: External applications can stream detected malicious IPs to NATS (`meds.block.ip` subject). Meds intercepts these events asynchronously and drops malicious flows.

---

## ✨ Key Features

- **Hybrid Kernel/User space Processing**  
  Meds utilizes a stateful marking architecture. It "teaches" the Linux kernel how to handle specific flows by assigning **Conntrack marks**, achieving wire-speed performance for established connections.

- **Intelligent NFQUEUE Balancing**  
  Intercepts traffic using `NFQUEUE` with `balance` and `bypass` options, ensuring multi-core scaling and system stability even if the user-space process is restarted.

- **Embedded NATS**  
  Exposes an asynchronous reactive API for external applications to offload detected threat vectors to the L3/L4 network layer instantly.

- **Lock-free Core Architecture**  
  The core engine is built for high-concurrency performance: no mutexes in the hot path. All filtering, counters, and rate-limiters utilize atomic operations.

- **Rate Limiting**  
  Uses token bucket algorithm to limit burst and sustained traffic per source IP, protecting the system against high-frequency floods (SYN, DNS, ICMP, or generic packet floods).

- **Blocklist-based filtering**  
  - IP blocklists: [FireHOL](https://iplists.firehol.org/), [Spamhaus DROP](https://www.spamhaus.org/drop/), [Abuse.ch](https://abuse.ch/)
  - ASN blocklists: [Spamhaus ASN DROP](https://www.spamhaus.org/drop/asndrop.json) using [IPLocate.io](https://iplocate.io/) for IP-to-ASN mapping
  - Domain blocklists: [StevenBlack hosts](https://github.com/StevenBlack/hosts/), [SomeoneWhoCares hosts](https://someonewhocares.org/hosts/)

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
  Built-in API server allows dynamically adding or removing IP or Country entries in global allow/block lists.  
  Auth via Basic Auth using `MEDS_USERNAME` / `MEDS_PASSWORD`.

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
meds_core_packets_accepted_total{filter="empty",reason="default"} 44344

# HELP meds_core_packets_dropped_total Total number of dropped packets
# TYPE meds_core_packets_dropped_total counter
meds_core_packets_dropped_total{filter="asn",reason="Spamhaus"} 1732
meds_core_packets_dropped_total{filter="domain",reason="StevenBlack"} 2
meds_core_packets_dropped_total{filter="ip",reason="FireHOL"} 6705
meds_core_packets_dropped_total{filter="ip",reason="Nats"} 432

# HELP meds_core_packets_processed_total Total number of processed packets
# TYPE meds_core_packets_processed_total counter
meds_core_packets_processed_total 53215
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
