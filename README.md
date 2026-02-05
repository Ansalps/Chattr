# Chattr 🚀  
### A Scalable Microservices Social Media Ecosystem

**Chattr** is a full-stack social media platform engineered for **high scalability**, **real-time engagement**, and **fault tolerance**.  
The project demonstrates a production-grade **microservices architecture**, leveraging **SQL & NoSQL databases**, **event-driven communication**, and **advanced caching strategies**.

---

## 🧠 System Architecture Overview

Chattr follows a **domain-driven microservices architecture**, where each service owns its data and business logic.

### 🔗 Communication Model
- **Client → API Gateway**: REST
- **Inter-service communication**: gRPC
- **Real-time features**: WebSockets
- **Asynchronous events**: Apache Kafka

---

### 📊 Data Ownership
Each service owns its database to ensure loose coupling and independent scalability:

- Auth & Subscription Service → PostgreSQL
- Post & Relation Service → PostgreSQL
- Chat Service → MongoDB
- Notification Service → PostgreSQL
- Redis → Shared cache (feeds & token blacklist)

---

### 🔄 Event Flow Example (Post Creation)

1. User creates a post via **API Gateway**
2. Request is forwarded to **Post & Relation Service**
3. Post is persisted in PostgreSQL
4. A `post_like` event is published to Kafka
5. Consumers:
   - Notification Service → sends notifications

This ensures **low latency**, **high throughput**, and **eventual consistency**.

---

### 🧵 Real-Time Messaging Flow

1. Client establishes a WebSocket connection
2. Chat Service handles message delivery
3. Messages are stored in MongoDB
4. `message_sent` event is emitted to Kafka
5. Notification Service pushes real-time alerts

---

### 🛡 Fault Tolerance & Resilience
- Kafka ensures no event loss during service downtime
- Redis reduces database load for hot data
- Stateless services allow horizontal scaling


---

### 🔐 API Gateway
**Central Entry Point**
- Handles all incoming client requests
- Routes traffic to internal services

**Security**
- Authentication & authorization handling
- JWT validation

**Token Management**
- Redis-backed JWT blacklist on logout
- Prevents token reuse and ensures secure session invalidation

---

### 👤 Auth & Subscription Service
**Identity Management**
- Manages user profiles and authentication
- PostgreSQL as the primary data store

**Monetization**
- Implements **Blue Tick verification** via Razorpay Subscription API

**Webhook Integration**
- Listens to Razorpay lifecycle events:
  - `subscription.activated`
  - `subscription.charged`
  - `subscription.halted`
  - `subscription.cancelled`
  - `subscription.completed`
- Maintains real-time subscription status updates

---

### 📝 Post & Relation Service
**Social Graph**
- Manages follows, likes, and comments
- PostgreSQL for relational consistency

**High-Performance Newsfeed**
- Uses a **Hybrid Feed Model**
  - Cache-Aside + Push (Fan-out) strategy
- Active user feeds are cached in Redis for **sub-millisecond retrieval**

**Event Producer**
- Publishes user activity events to Apache Kafka

---

### 💬 Chat Service
**Real-time Messaging**
- Supports 1-on-1 and group chats
- WebSocket-based communication

**NoSQL Storage**
- MongoDB for flexible, document-based message storage
- Optimized for horizontal scaling and high write throughput

---

### 🔔 Notification Service
**Event-Driven Architecture**
- Consumes events from Kafka (Post & Chat services)

**Real-time Delivery**
- Pushes notifications to clients via WebSockets

**Persistence**
- Stores notification history in PostgreSQL

---

## 🛠 Tech Stack

| Layer | Technology |
|------|-----------|
| Language | Go (Golang) |
| Communication | gRPC (Inter-service), REST (Gateway), WebSockets (Real-time) |
| Databases | PostgreSQL, MongoDB |
| Caching | Redis |
| Message Broker | Apache Kafka |
| Payments | Razorpay SDK |
| Infrastructure | Docker, Docker Compose |

---

## 🚀 Key Technical Challenges Solved

### 📰 The Newsfeed Problem (Hybrid Push/Pull)
To balance **write-heavy** and **read-heavy** workloads:
- **Active users** receive pushed updates directly into Redis
- **Inactive users** fetch posts on login
- Dramatically reduces unnecessary fan-out operations

---

### 🔄 Event-Driven Notifications
- Kafka decouples producers (Post & Chat services) from consumers
- Ensures **no data loss** if the Notification service is temporarily down
- Improves system resilience and scalability

---

### 💳 Razorpay Subscription Lifecycle Management
- Fully manages subscription states via webhooks
- Automatically revokes **Blue Tick** access on:
  - Cancellation
  - Halted payments
- Ensures real-time entitlement enforcement

---

## 🔧 Getting Started

### Prerequisites
- Go **1.21+**
- Docker & Docker Compose
- Razorpay API Keys

---

### Installation

#### 1️⃣ Clone the Repository
```bash
git clone https://github.com/Ansalps/Chattr.git
cd Chattr


#### 2️⃣ Spin Up Infrastructure
Start all required dependencies (Kafka, Redis, PostgreSQL, MongoDB) using Docker Compose:

```bash
docker-compose up -d

#### 3️⃣ Run Services

Each microservice in **Chattr** runs independently.  
Open a new terminal window for each service and start them one by one.

##### API Gateway
```bash
cd api-gateway
go run cmd/main.go

##### Auth_Subcription Service
```bash
cd auth_subscription_svc
go run cmd/main.go

##### Post_Relation Service
```bash
cd post_relation_svc
go run cmd/main.go

##### Chat Service
```bash
cd chat_svc
go run cmd/main.go

##### Notification Service
```bash
cd notifcation_svc
go run cmd/main.go

