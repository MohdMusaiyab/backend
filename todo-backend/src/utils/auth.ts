import bcrypt from "bcrypt";
import jwt from "jsonwebtoken";
import { env } from "../lib/env.js";

const SALT_ROUNDS = 10;
export async function hashPassword(password: string): Promise<string> {
  return await bcrypt.hash(password, SALT_ROUNDS);
}
export async function verifyPassword(
  password: string,
  hash: string,
): Promise<boolean> {
  return await bcrypt.compare(password, hash);
}

export function generateTokens(userId: string) {
  const accessToken = jwt.sign({ userId }, env.JWT_ACCESS_SECRET, {
    expiresIn: "15m",
  });

  const refreshToken = jwt.sign({ userId }, env.JWT_REFRESH_SECRET, {
    expiresIn: "7d",
  });

  return { accessToken, refreshToken };
}

export function verifyRefreshToken(token: string) {
  try {
    return jwt.verify(token, env.JWT_REFRESH_SECRET) as { userId: string };
  } catch (error) {
    return null;
  }
}
