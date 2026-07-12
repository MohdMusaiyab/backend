# Go Notification System

A highly scalable, production-grade notification system built in Go. This project is being constructed in iterative stages to demonstrate architectural evolution—from a simple API to a high-throughput, distributed asynchronous architecture.

---

## Stage 1: The Synchronous Foundation (Completed)

### What We Built
In Stage 1, we established a strict **Clean Architecture** to ensure our codebase is fully decoupled and easy to maintain:
- **Transport Layer:** Handles incoming web requests and ensures the data is valid.
- **Core Service Layer:** The "Brain". It orchestrates the rules of sending messages and saving database records.
- **Data Access Layer:** Safely manages all database operations (PostgreSQL).
- **Provider Layer:** A mock external service (simulating a provider like Twilio or AWS) to send the messages.

### The Bottleneck
While structurally sound, the system was entirely **Synchronous**. If the external email provider took 500ms to send a message, the user was forced to stare at a loading screen for 500ms. If 10,000 users triggered emails at once, the server would freeze waiting for all those delays to finish.

---

## Stage 2: The Asynchronous Powerhouse (Completed)

### What We Built
To solve the bottleneck, we completely decoupled the API from the heavy lifting by introducing a **Message Queue (Redis)** and a **Background Worker Pool**.

1. **The Producer (Lightning Fast API):** 
   Instead of forcing the user to wait for the email to send, our API now instantly drops a "Task" into the Redis queue and immediately replies to the user. **API response times dropped from ~500ms to under 25ms!**
   
2. **The Consumer (Background Workers):** 
   A dedicated pool of background workers continuously watches the queue. When a task appears, a worker silently picks it up and handles the slow 500ms process of sending the email behind the scenes.

### Why This is Incredible (The Benefits)
- **Zero Lag for Users:** The application feels lightning-fast because the user never waits for the slow external email servers.
- **Extreme Scalability:** The server can accept thousands of notification requests instantly without crashing, simply piling them safely into the queue.
- **Automatic Retries:** If the external email service goes down temporarily, the system doesn't lose the email. The worker will automatically wait and try again up to 5 times.
- **Graceful Shutdowns:** If we need to restart our servers for an update, the workers will finish sending their current emails before safely shutting down, ensuring zero lost data.

### The Pitfalls of Stage 2 (Why we need Stage 3)
While this is a massive operational upgrade, the current architecture still has critical flaws:
1. **Coupled Monolith (Single Point of Failure):** Right now, our HTTP API and our Background Workers are running inside the exact same Go binary (`main.go`). If the HTTP server crashes due to a memory leak or bad request, it instantly kills all the background workers with it. 
2. **Inability to Scale Independently:** In a real production environment, you might need 10 API servers to handle a massive spike in incoming web traffic, but only need 2 Worker servers to chew through the queue. Because they are baked into the same file, we are forced to scale them 1:1, which wastes server resources.
3. **No Real-Time Feedback:** Because the API returns instantly, the frontend (client) has no idea when the email *actually* sends, or if it permanently fails after 5 retries. 

---

## Next Steps (Stage 3)
*Coming soon...*
