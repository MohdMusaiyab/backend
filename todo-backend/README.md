# Todo API Backend

A production-ready, secure, and highly scalable REST API for managing Todos. Built with modern Node.js practices, it includes robust authentication, dynamic pagination, data validation, and automated Swagger documentation.

## Features

- **Robust Authentication**: Secure registration, login, and token refresh mechanisms using JWT (short-lived access tokens and HttpOnly refresh cookies).
- **Password Recovery**: Complete OTP-based forgot & reset password flow.
- **Type-Safe Validation**: End-to-end type safety using Zod for payload validation and TypeScript for static checking.
- **Advanced Pagination & Filtering**: Filter Todos by completion status, perform case-insensitive searches, and paginate responses seamlessly.
- **Global Error Handling**: Centralized error middleware ensures predictable, standard error formats and completely eliminates noisy `try/catch` blocks in controllers.
- **Auto-Generated Documentation**: Live Swagger UI generated automatically from Zod validation schemas (`/api-docs`).
- **Containerized**: Fully configured Docker environment for 1-click deployments.

## Tech Stack

- **Runtime Environment**: Node.js & TypeScript
- **Web Framework**: Express (v5)
- **Database & ORM**: PostgreSQL & Prisma
- **Validation**: Zod
- **Security**: bcrypt (password hashing), jsonwebtoken (JWT)
- **Containerization**: Docker & Docker Compose

## Quick Start (Docker)

The absolute easiest way to get the project running is using Docker. You don't even need Node or Postgres installed locally.

1. Ensure Docker and Docker Compose are installed.
2. Clone the repository and open the folder.
3. Spin up the Database and the API in one command:
   ```bash
   docker compose up --build
   ```
4. Access the API at `http://localhost:3000`
5. View the interactive Swagger API Docs at `http://localhost:3000/api-docs`

## Manual Local Setup

If you prefer to run it manually on your local machine:

1. Copy the sample environment file:
   ```bash
   cp .env.sample .env
   ```
2. Make sure you have a local PostgreSQL instance running and update the `DATABASE_URL` in your `.env`.
3. Install dependencies:
   ```bash
   npm install
   ```
4. Generate the Prisma Client & push migrations:
   ```bash
   npx prisma generate
   npx prisma db push
   ```
5. Start the development server:
   ```bash
   npm run dev
   ```

## Architecture

- **`src/schemas/`**: Zod schemas defining the shape of incoming requests. This serves as our single source of truth for validation and Swagger docs.
- **`src/routes/`**: Express routers that mount middleware and map HTTP verbs to specific controllers.
- **`src/controllers/`**: Lightweight functions completely free of error-handling noise, passing validated data to services.
- **`src/services/`**: The core business logic and Prisma database interactions.
- **`src/middlewares/`**: JWT Auth validation and our massive Global Error Handler catch-all.
