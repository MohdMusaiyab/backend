# Async Image Processing Pipeline

## Overview
A production-grade, asynchronous image processing backend utilizing a robust message queue architecture. This project demonstrates how to decouple heavy, CPU-intensive tasks (like image manipulation) from the main API thread. This separation of concerns ensures high availability, extremely fast response times, and scalable background processing.

Included is a custom Vanilla JS frontend with a premium SaaS-style dashboard interface, allowing users to upload images and visualize the real-time progress of the background worker via status polling.

---

## Architecture

The system is built on a classic **Producer-Consumer** pattern:

1. **Producer (Express API):** 
   Receives image uploads (via `multer`), saves the raw file locally, creates a `pending` job record in the database, and pushes a job ID to the Redis queue. It immediately returns a `202 Accepted` status to the client, ensuring the main Node.js event loop is never blocked by heavy processing.
   
2. **Message Broker (Redis + BullMQ):** 
   Acts as the highly reliable, in-memory queue. Incoming jobs wait here until a worker has the capacity to process them.

3. **Consumer (Background Worker):** 
   A completely separate Node.js process that listens exclusively to the queue. It picks up pending jobs, updates the database status to `processing`, performs heavy image transformations using `sharp` (resizing, grayscale conversion, and blurring), saves the final image, and marks the database status as `completed` (or `failed` if an error occurs).

---

## Tech Stack & Rationale

*   **Node.js & Express**: The core backend API framework chosen for its lightweight, event-driven nature.
*   **TypeScript**: Enforces end-to-end type safety, preventing runtime errors and vastly improving the developer experience.
*   **PostgreSQL & Prisma 7**: Serves as the persistent "Single Source of Truth" for job states. Prisma 7 is utilized with the `@prisma/adapter-pg` driver for modern, type-safe database interactions.
*   **BullMQ & ioredis**: The industry-standard choice for Node.js message queues. Backed by Redis, it provides robust job management, atomic operations, and built-in failure/retry handling.
*   **Sharp**: A high-performance Node.js image processing library (backed by `libvips`) used to perform the actual image manipulation.
*   **Docker**: Used to seamlessly provision local, reproducible instances of PostgreSQL and Redis.
*   **Vanilla HTML/CSS/JS**: A sleek, dependency-free frontend to demonstrate the API's capabilities without the overhead of a heavy SPA framework.

---

## Getting Started

### Prerequisites
*   Node.js (v18+)
*   Docker & Docker Compose

### 1. Infrastructure Setup
Ensure your local PostgreSQL and Redis instances are running. You can quickly spin them up using standalone Docker commands:

**Start PostgreSQL:**
```bash
docker run --name pg-image-queue -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
```

**Start Redis:**
```bash
docker run --name redis-queue -p 6379:6379 -d redis
```

### 2. Environment Configuration
Create a `.env` file in the root directory:
```env
DATABASE_URL="postgresql://postgres:password@localhost:5432/image_processor?schema=public"
REDIS_URL="redis://localhost:6379"
PORT=3000
```

### 3. Installation & Database Sync
Install the required packages and sync the Prisma schema with your database:
```bash
npm install
npx prisma migrate dev --name init
```

### 4. Running the Application
Because this is a decoupled queue architecture, the API and the Worker **must be run as separate processes in two different terminal windows.**

**Terminal 1 (Start the API Server / Producer):**
```bash
npm run dev:server
```

**Terminal 2 (Start the Background Worker / Consumer):**
```bash
npm run dev:worker
```

### 5. Usage
1. Open your browser and navigate to `http://localhost:3000`.
2. Upload an image (JPEG/PNG) via the drag-and-drop interface.
3. The UI will instantly provide feedback and begin polling the `/status` API endpoint.
4. Observe the worker terminal as it picks up the job, applies the filters, and saves the file.
5. The UI will automatically render the final, processed image once the worker completes the job.
