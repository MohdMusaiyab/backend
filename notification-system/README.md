# Go Notification System

A highly scalable, production-grade notification system built in Go. This project is being constructed in iterative stages to demonstrate architectural evolution—from a simple synchronous API to a high-throughput, distributed asynchronous architecture.

## Stage 1: The Synchronous Foundation

### What We Built
In Stage 1, we established a strict **3-Tier Clean Architecture** to ensure our codebase is fully decoupled, easily testable, and highly maintainable:
- **Transport Layer (Gin):** Handles HTTP requests and enforces strict JSON validation using `validator/v10`.
- **Core Service Layer:** The "Brain". It orchestrates sending messages and saving database records, remaining completely oblivious to the HTTP framework.
- **Data Access Layer (GORM + Postgres):** Schema versioning and structure are strictly controlled via `golang-migrate` (raw SQL files), while GORM handles the day-to-day Go struct querying.
- **Provider Layer:** A Mock Sender simulating external network calls (e.g., Twilio/AWS SES) injected into the service via Dependency Injection.

### Cons of the Current Architecture (The Bottleneck)
While the code is structurally beautiful, it has a massive operational flaw under high load: **It is entirely Synchronous.**

1. **Latency Bottleneck:** Our HTTP handler waits for the `NotificationSender` to finish before returning a response. If the external provider takes 500ms to send an SMS, the client is forced to wait 500ms.
2. **Resource Exhaustion:** If 10,000 notifications are triggered simultaneously, the API server will spawn 10,000 goroutines that simply sit there sleeping/waiting for 500ms. This ties up memory and maxes out database connection pools unnecessarily.
3. **No Fault Tolerance (Retries):** If the external API (like Twilio) goes down for 5 seconds, the request fails, the API returns a 500 error, and the notification is permanently lost.

### Preparing for Stage 2: Asynchronous Processing
To solve these critical bottlenecks, Stage 2 will introduce an **Asynchronous Message Queue** (such as RabbitMQ, Kafka, or Redis with BullMQ/Asynq). 

In the next stage, the API will simply save the notification to the DB as `"pending"`, push a message to the queue, and instantly return a `202 Accepted` to the user in under `5ms`. Dedicated background "Worker" servers will consume that queue, perform the slow 500ms network calls, update the database, and handle automatic retries if the external provider fails.
