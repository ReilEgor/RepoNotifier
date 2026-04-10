# 🚀 RepoNotifier

> **RepoNotifier** is a production-ready Go monolith that tracks GitHub repository releases and sends real-time email notifications to subscribers.

---

[![codecov](https://codecov.io/gh/ReilEgor/NotifierTest/graph/badge.svg?token=S8KWDBMUQ7)](https://codecov.io/gh/ReilEgor/NotifierTest)

---
## ⚒ Core Stack

### Backend & API
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=go&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-4285F4?style=for-the-badge&logo=grpc&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)

### Storage & Caching
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)

### Resilience & Monitoring
![Circuit Breaker](https://img.shields.io/badge/Circuit_Breaker-gobreaker-orange?style=for-the-badge&logo=go)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)

### Infrastructure
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![GitHub API](https://img.shields.io/badge/GitHub_API-181717?style=for-the-badge&logo=github&logoColor=white)

---

## ⭐ Overview

RepoNotifier continuously monitors GitHub repositories and notifies users when new releases are published.

It is designed with **Clean Architecture**, strong **resilience patterns**, and **observability** in mind.

---

## ⭐ Features

- 🔔 **Automated Tracking**  
  Background scanner detects new releases using `last_seen_tag` strategy

- 📫 **Email Notifications**  
  Instant alerts via SMTP or Email API (Mailtrap / SendGrid-ready)

- 🛡 **Rate Limit Awareness**  
  Handles GitHub API `429 Too Many Requests` gracefully

- ⚡ **Caching Layer**  
  Redis used for caching and rate-limiting optimization

- 🧱 **Clean Architecture**  
  Decoupled layers with dependency injection

- 🌐 **Dual Interface**  
  REST API + gRPC support

- 📊 **Observability**  
  Prometheus metrics + Grafana dashboards

- 🔥 **Resilience Patterns**
  - Circuit Breaker (gobreaker)
  - Retry strategy
  - Graceful shutdown

---

## 🏗 Architecture

### C4 Model

<img width="4524" height="1768" src="https://github.com/user-attachments/assets/15231bf2-ac06-43d8-b861-b3b8e1e63163" />
<img width="7400" height="3444" src="https://github.com/user-attachments/assets/cc60b912-f9e4-4bb4-b371-85b2a8344b45" />

---

### 💾 Database Schema

<img width="617" height="671" alt="image" src="https://github.com/user-attachments/assets/f7fd9d6b-6119-4cc7-82bf-0edf38f16ba9" />

---

## 🚀 Quick Start

```bash
git clone https://github.com/ReilEgor/RepoNotifier.git
cd RepoNotifier

cp .env.example .env

docker-compose -f deployments/docker-compose.yml up --build
```

## ◀ Verify
- REST API: http://localhost:8080
- Swagger: http://localhost:9080
- Metrics: http://localhost:8080/metrics
