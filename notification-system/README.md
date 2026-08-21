# Go Notification System

A highly scalable, production-grade notification system built in Go. I am constructing this project in iterative stages to demonstrate architectural evolution—from a simple API to a high-throughput, distributed asynchronous architecture.

---

## 🚀 Live Observability Dashboard & Quick Start

To visually track the entire distributed system in real-time, I built a custom **Next.js Observability Dashboard**. It uses WebSockets to hook directly into the Go backend's telemetry, rendering a live architectural heatmap and isolating concurrent requests into dedicated terminal swimlanes.

### How to Run the Full Stack

1. **Start the Infrastructure & Metrics (Docker)**
   ```bash
   docker-compose up -d
   ```
2. **Start the Go Backend (API + Workers)**
   ```bash
   go run cmd/api/main.go
   ```
3. **Start the Next.js Dashboard**
   ```bash
   cd dashboard
   npm install
   npm run dev
   ```

### Service Access Points
*   **Custom Observability UI:** [http://localhost:3001](http://localhost:3001) *(Trigger events and watch the live heatmap!)*
*   **Grafana Dashboards:** [http://localhost:3000](http://localhost:3000) *(Long-term statistical metrics)*
*   **Prometheus:** [http://localhost:9090](http://localhost:9090) *(Raw metric scraping)*
*   **Go API Gateway:** `http://localhost:8080`

---

## Stage 1: The Synchronous Foundation (Completed)

### What I Built
In Stage 1, I established a strict **Clean Architecture** to ensure my codebase is fully decoupled and easy to maintain:
- **Transport Layer:** Handles incoming web requests and ensures the data is valid.
- **Core Service Layer:** The "Brain". It orchestrates the rules of sending messages and saving database records.
- **Data Access Layer:** Safely manages all database operations (PostgreSQL).
- **Provider Layer:** A mock external service (simulating a provider like Twilio or AWS) to send the messages.

### The Bottleneck
While structurally sound, the system was entirely **Synchronous**. If the external email provider took 500ms to send a message, the user was forced to stare at a loading screen for 500ms. If 10,000 users triggered emails at once, the server would freeze waiting for all those delays to finish.

---

## Stage 2: The Asynchronous Powerhouse (Completed)

### What I Built
To solve the bottleneck, I completely decoupled the API from the heavy lifting by introducing a **Message Queue (Redis)** and a **Background Worker Pool**.

1. **The Producer (Lightning Fast API):** 
   Instead of forcing the user to wait for the email to send, my API now instantly drops a "Task" into the Redis queue and immediately replies to the user. **API response times dropped from ~500ms to under 25ms!**
   
2. **The Consumer (Background Workers):** 
   A dedicated pool of background workers continuously watches the queue. When a task appears, a worker silently picks it up and handles the slow 500ms process of sending the email behind the scenes.

### Why This is Incredible (The Benefits)
- **Zero Lag for Users:** The application feels lightning-fast because the user never waits for the slow external email servers.
- **Extreme Scalability:** The server can accept thousands of notification requests instantly without crashing, simply piling them safely into the queue.
- **Automatic Retries:** If the external email service goes down temporarily, the system doesn't lose the email. The worker will automatically wait and try again up to 5 times.
- **Graceful Shutdowns:** If I need to restart my servers for an update, the workers will finish sending their current emails before safely shutting down, ensuring zero lost data.

### The Pitfalls of Stage 2 (Why I needed Stage 3 and beyond)
While this was a massive operational upgrade, the Stage 2 architecture still had critical flaws:
1. **Coupled Monolith (Single Point of Failure):** Right now, my HTTP API and my Background Workers are running inside the exact same Go binary (`main.go`). If the HTTP server crashes due to a memory leak or bad request, it instantly kills all the background workers with it. 
2. **Inability to Scale Independently:** In a real production environment, I might need 10 API servers to handle a massive spike in incoming web traffic, but only need 2 Worker servers to chew through the queue. Because they are baked into the same file, I am forced to scale them 1:1, which wastes server resources.
3. **No Real-Time Feedback:** Because the API returns instantly, the frontend (client) has no idea when the email *actually* sends, or if it permanently fails after 5 retries. 

---

## Stage 3: System Resilience & DLQ (Completed)

### What I Built
In distributed systems, external APIs (like Twilio or AWS SES) will inevitably go down. In Stage 3, I hardened the system against these catastrophic failures:

1. **Exponential Backoff + Jitter:** If the external provider fails, my Go worker does not immediately spam them again (which would cause a "Thundering Herd" server crash). It intelligently calculates an exponential delay with a random time-jitter before retrying.
2. **Strict Queue Prioritization:** I created multiple queue levels. A "critical" password reset email will automatically be processed 6x faster than a "low" priority weekly newsletter.
3. **Dead Letter Queue (DLQ):** If a notification fails 3 times consecutively, it is stripped from the active queue and permanently parked in the DLQ (Archived Queue) to prevent infinite loops.
4. **Visual Monitoring:** I deployed the `hibiken/asynqmon` Docker container, providing a live Web UI to visually monitor queue throughput, track retries, and manually replay DLQ tasks.

---

## Stage 4: Idempotency & Deduplication (Completed)

### What I Built
"Exactly-Once" delivery over a network is a myth. In Stage 4, I engineered for **"Effectively-Once"** delivery by making my architecture inherently skeptical of duplicates. 

1. **API Layer Defense (Type A Duplicates):** If a user's mobile app loses Wi-Fi and they panic-click the "Submit" button 5 times, my API receives 5 identical HTTP requests. I introduced an `Idempotency-Key` HTTP Header backed by a strict PostgreSQL `UNIQUE` constraint. The database brutally rejects the 4 duplicate clicks at the network edge, preventing them from ever reaching the queue.
2. **Worker Layer Defense (Type B Duplicates):** If my background worker successfully sends the email, but the server loses power literally 1 millisecond before it can acknowledge the job, the Redis broker assumes it failed and will redeliver it 5 minutes later. I updated the worker to query the database *before* sending: if the status is already marked as `"sent"`, it acts as a physical database lock, skipping the external API call and gracefully deleting the ghost task.

---

## Stage 5: Multiple Channels & Fan-Out (Completed)

### What I Built
In a real enterprise system, an "Order Shipped" event doesn't just trigger one action—it triggers an Email, an SMS, and a Push Notification. I transitioned my system from a simple 1-to-1 pipeline into a true **Event-Driven Router Architecture**.

1. **The Pub/Sub Router:** My HTTP API no longer pushes direct "Send Email" tasks. Instead, it drops a single generic `event:notification_requested` into the queue. I built a dedicated Router Worker that picks up this event, acts as a middleman, and "fans out" by creating independent tasks for Email and SMS.
2. **Failure Isolation:** I created completely separate Redis queues for Email and SMS. If my SMS provider goes down and thousands of tasks get stuck in the `sms` queue, my `email` queue remains completely empty and lightning fast. This guarantees that one bad dependency never degrades the entire system.
3. **1-to-Many Database Schema:** I updated my PostgreSQL database to use a relational schema. A single broadcast event now physically links to multiple specific delivery records, allowing me to track exactly which channels succeeded and which ones failed.

---

## Stage 6: Rate Limiting & Backpressure (Completed)

### What I Built
When systems scale, infinite traffic is a curse, not a blessing. I built three distinct layers of defense to protect my infrastructure and my downstream dependencies from crashing under massive load spikes.

1. **API Gateway Rate Limiting (Token Bucket):** I implemented an in-memory Token Bucket algorithm middleware for the HTTP API. It maps buckets strictly by IP address, rejecting aggressive spam with a `429 Too Many Requests` status before it can ever touch the Redis queues.
2. **Queue Backpressure (Load Shedding):** I integrated `asynq.Inspector` directly into the producer. If my Redis queues exceed 5,000 pending tasks, my API intentionally triggers *Graceful Degradation*. It stops accepting requests and returns a `503 Service Unavailable` error, preventing the server from running out of RAM. 
3. **Distributed Worker Throttling:** I implemented a globally atomic, Redis-backed Fixed Window Counter for my SMS workers. Even if I spin up 50 servers running 100 SMS workers concurrently, this distributed lock mathematically guarantees I will never exceed my provider's strict requests-per-second limit.

---

## Stage 7: User Preferences & Templating (Completed)

### What I Built
I transitioned the system from a pure infrastructure pipeline into a business-ready product. Real-world systems don't just blindly spam users; they enforce strict rules and offer deep personalization.

1. **JSONB Preference Engine:** I utilized PostgreSQL's `JSONB` data type to build a wildly flexible user preferences schema. When the Router pulls an event, it queries the DB, automatically parses the JSON, and strictly enforces the user's opt-out choices. If a user disabled SMS, the Router intentionally skips the SMS queue, saving server resources and API credits.
2. **Immutable Template Versioning (The Mid-Flight Fix):** To prevent "Distributed Race Conditions", I architected a version-locked templating engine. The API resolves the absolute newest template (e.g., `v2`) and stamps it permanently onto the job payload. Even if a marketer publishes `v3` while the job is stuck in the queue, the worker guarantees it uses `v2` to render, completely eliminating mid-flight crashes.
3. **Dynamic Go Rendering:** I stripped out hardcoded strings and implemented Go's `html/template` engine. The worker mathematically merges the raw database template with the highly dynamic JSON payload sitting in the Redis queue to generate beautifully personalized messages just milliseconds before delivery.

---

## Stage 8: Scheduling & Delayed Notifications (Completed)

### What I Built
I implemented "Time-Travel Deliveries" using Broker-Native Delay mechanisms, allowing the system to perfectly schedule notifications for future execution without blocking the API or constantly polling the database.

1. **Database Schema Expansion:** Safely added an optional `send_at` timestamp to the Postgres `notifications` table to keep a permanent, historical record of intended execution times.
2. **Go Model & API Updates:** Upgraded the JSON handler to accept and natively validate ISO-8601 timestamps (`*time.Time`), explicitly rejecting any attempts to schedule messages in the past.
3. **Redis ZSET Magic:** Integrated `asynq.ProcessAt()` into the Service Layer. If a future timestamp is detected, the task bypasses the active queue and is placed directly into a Redis Sorted Set (ZSET). Asynq silently monitors this set in-memory and drops the task into the active queue at the exact millisecond it is due.

---

## Stage 9: Observability (Completed)

### What I Built
I transformed the system from a "blind box" into a fully observable, production-ready architecture by implementing industry-standard metrics, structured JSON logging, and end-to-end distributed tracing.

1. **The Telemetry Stack:** Added Prometheus and Grafana via Docker Compose. Exposed a `/metrics` endpoint in the Go API using `promhttp` to allow Prometheus to actively scrape the system's health every 15 seconds.
2. **Prometheus Instrumentation:** Injected `prometheus/client_golang` into the background workers. I utilized Go's `defer` and Named Return Variables to guarantee the system records absolute `notification_processed_total` counters and exact `notification_processing_duration_seconds` latency histograms, even if a worker panics.
3. **Structured JSON Logging & Tracing:** Stripped out standard text logs and replaced them with Go's modern `log/slog`. Implemented "Poor Man's Tracing" by generating an `X-Request-ID` at the API Gateway, serializing it into the Redis Queue payload, and binding it to the worker loggers (`slog.With`). This guarantees every single log printed globally contains the exact trace ID of the user's request.

---

## Stage 10: Microservices & Distributed Telemetry (Completed)

### What I Built
I executed a physical architecture split, transforming the system from a single Go binary (Monolith) into a fully distributed microservice architecture. 

1. **Breaking the Monolith:** I shattered the `main.go` file into four completely isolated entry points: `cmd/api`, `cmd/worker-router`, `cmd/worker-email`, and `cmd/worker-sms`. 
2. **Containerization & Orchestration:** I authored a dynamic, Multi-Stage `Dockerfile` capable of building any of the four microservices on demand. Using `docker-compose.yml`, the infrastructure is now orchestrated so I can independently scale workers (e.g., running `docker compose up --scale worker-email=5` to spin up 5 concurrent email nodes to chew through backlogs).
3. **Centralized Redis Logging (The Bridge):** Because the workers are now physically isolated Linux containers, they could no longer write their logs directly into the API's WebSocket memory space. I solved this by engineering a custom `RedisPubSubWriter`. Every worker now publishes its JSON logs to a `global_telemetry` Redis channel. The API Gateway subscribes to this channel and pipes the distributed logs straight back into the WebSocket, keeping the Next.js Dashboard perfectly synced!

### The Benefits
- **Fault Isolation:** If an SMS worker crashes due to an out-of-memory error from a bad dependency, it only kills its specific Docker container. The API Gateway and Email workers continue running flawlessly with zero downtime.
- **Asymmetric Scaling:** I no longer waste CPU/RAM spinning up unnecessary API servers just to get more background workers. I can dynamically assign computing resources strictly to the queues that are backlogged.
- **True Observability:** The centralized logging proves that no matter how many nodes are spun up horizontally across a cloud cluster, telemetry can always be securely routed back to a central observability hub in real-time.
