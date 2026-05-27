import { z } from "zod";
import dotenv from "dotenv";

// Load environment variables from .env file
dotenv.config();

const envSchema = z.object({
  DATABASE_URL: z.string().url({ message: "DATABASE_URL must be a valid URL" }),
  PORT: z.string().default("3000"),
  NODE_ENV: z.enum(["development", "production", "test"]).default("development"),
  JWT_ACCESS_SECRET: z.string().min(1, "JWT_ACCESS_SECRET is required"),
  JWT_REFRESH_SECRET: z.string().min(1, "JWT_REFRESH_SECRET is required"),
});

// Validate the environment variables
const _env = envSchema.safeParse(process.env);

if (!_env.success) {
  console.error("❌ Invalid environment variables:");
  console.error(JSON.stringify(_env.error.format(), null, 2));
  process.exit(1);
}

export const env = _env.data;
