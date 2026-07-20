# Go Notification System

A highly scalable, production-grade notification system built in Go. I am constructing this project in iterative stages to demonstrate architectural evolution—from a simple API to a high-throughput, distributed asynchronous architecture.

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

## Next Steps (Stage 7: User Preferences & Templating)
- **User Preferences:** Building a real-world subscription mechanism where users can explicitly opt-in or opt-out of specific notification channels (e.g., "Send me emails, but turn off SMS alerts").
- **Dynamic Templating:** Transitioning from hardcoded string messages to rich, dynamic templates (e.g., passing in an event payload and injecting the user's real name and order details).
