# RepoNotifier

> **RepoNotifier** is a lightweight Go-based monolith service that tracks GitHub repository releases and sends real-time email notifications to subscribers.

---

## ✨ Features

- 🔔 **Automated Tracking** — Background scanner checks for new GitHub tags/releases using `last_seen_tag` logic.
- 📧 **Email Notifications** — Immediate alerts sent to subscribers via SMTP/Email API upon new release detection.
- 🛡️ **Rate Limit Awareness** — Intelligent GitHub API client that handles `429 Too Many Requests` and optimizes token usage.
- 🏗️ **Clean Architecture** — Dependency injection and interface-driven design for high testability and maintainability.
- 🚀 **Dual Interface** — RESTful API for web integration and gRPC for high-performance internal communication.
- 📊 **Monitoring** — Native Prometheus metrics endpoint (`/metrics`) to track system health and notification stats.

---

## 🏗️ Architecture (C4 Context)
<img width="4524" height="1768" alt="image" src="https://github.com/user-attachments/assets/0c67906a-a438-415d-9e9c-d0b16c5d21df" />

The service follows **Clean Architecture** principles:

- **Delivery Layer** — HTTP / gRPC handlers
- **UseCase Layer** — business logic
- **Domain Layer** — core entities
- **Infrastructure Layer** — DB, GitHub API, Email

---

## 🧠 How It Works

1. User subscribes to a GitHub repository via API
2. Service validates repository via GitHub API
3. Subscription is stored with `last_seen_tag`
4. Background scanner periodically checks for new releases
5. If a new tag is detected:
   - Email notification is sent
   - `last_seen_tag` is updated
