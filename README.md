Chattr: A Scalable Microservices Social Media Ecosystem
Chattr is a full-stack social media platform engineered for high scalability and real-time engagement. The project demonstrates a sophisticated microservices architecture, leveraging multiple database technologies (SQL & NoSQL), message brokers for asynchronous processing, and advanced caching strategies.

🏗 System Architecture
The system is divided into five core components, each designed to handle specific domains:

1. API Gateway
Central Entry Point: Handles all incoming client requests.

Security: Implements authentication logic.

Token Management: Integrated with Redis to blacklist JWT tokens upon logout, ensuring secure session management.

2. Auth & Subscription Service
Identity Management: Manages user profiles and authentication data in PostgreSQL.

Monetization: Implements "Blue Tick" verification via Razorpay Subscription API.

Webhook Integration: Listens for Razorpay events (subscription.activated, charged, halted, cancelled, completed) to maintain real-time subscription status.

3. Post & Relation Service
Social Graph: Manages follows, likes, and comments using PostgreSQL.

High-Performance Newsfeed: * Uses a Hybrid Model: Combines Cache-Aside and Push Model (Fan-out) patterns.

Caches active user feeds in Redis for sub-millisecond retrieval.

Event Producer: Publishes activity events to Kafka topics.

4. Chat Service
Real-time Messaging: Supports both 1-on-1 and Group chats using WebSockets.

NoSQL Storage: Uses MongoDB for flexible, document-based storage of message history, optimized for horizontal scaling.

5. Notification Service
Event-Driven: Consumes events from Kafka (triggered by Post and Chat services).

Real-time Delivery: Pushes notifications back to clients via WebSockets.

Persistence: Stores notification history in PostgreSQL.

🛠 Tech Stack
Layer             Technology
Languages         Go(Golang)
Communication     "gRPC (Inter-service), WebSockets (Real-time), REST (Gateway)"
Databases         "PostgreSQL, MongoDB"
Caching/Messaging Redis (Feed Cache & Token Blacklist)
Message Broker    Apache Kafka
Payments          Razorpay SDK

🚀 Key Technical Challenges Solved
The Newsfeed Problem (Hybrid Push/Pull)
To balance write-heavy and read-heavy operations, the system uses a hybrid approach. For active users, posts are pushed to their Redis feed cache immediately. For inactive users, the system pulls updates only when they log in, significantly reducing overhead.

Event-Driven Notifications
By using Kafka, the Notification service is decoupled from the Post and Chat services. If the Notification service goes down, events remain in the Kafka queue, preventing data loss and ensuring a reliable user experience.

Razorpay Lifecycle Management
The system doesn't just process payments; it manages the entire lifecycle. By handling webhooks, the platform automatically revokes "Blue Tick" access the moment a subscription is halted or cancelled.

🔧 Getting Started

Prerequisites
Go 1.21+

Docker & Docker Compose (for Kafka, Redis, Postgres, Mongo)

Razorpay API Keys

Installation
1.  Clone the Repo:
   git clone https://github.com/Ansalps/Chattr.git
2.  Spin up Infrastructure:
   docker-compose up -d
3.  Run Services: Navigate to each service directory and run:
   go run cmd/main.go
