import { prisma } from "../lib/prisma.js";
import { hashPassword, verifyPassword, generateTokens, verifyRefreshToken } from "../utils/auth.js";
import type { LoginInput, RegisterInput } from "../schemas/auth.schema.js";

export const createUserService = async (data: RegisterInput) => {
  // 1. Check if user already exists
  const existingUser = await prisma.user.findUnique({
    where: { email: data.email },
  });

  if (existingUser) {
    throw new Error("User with this email already exists");
  }

  // 2. Hash the password
  const hashedPassword = await hashPassword(data.password);

  // 3. Create the user in the database
  const newUser = await prisma.user.create({
    data: {
      email: data.email,
      password: hashedPassword,
    },
    select: {
      id: true,
      email: true,
    },
  });

  return newUser;
};

export const loginUserService = async (data: LoginInput) => {
  const existingUser = await prisma.user.findUnique({
    where: {
      email: data.email,
    },
  });

  if (!existingUser) {
    throw new Error("Invalid email or password");
  }

  const isPasswordValid = await verifyPassword(data.password, existingUser.password);
  
  if (!isPasswordValid) {
    throw new Error("Invalid email or password");
  }

  const { accessToken, refreshToken } = generateTokens(existingUser.id);

  return {
    user: { id: existingUser.id, email: existingUser.email },
    accessToken,
    refreshToken
  };
};

export const refreshUserService = async (refreshToken: string) => {
  const payload = verifyRefreshToken(refreshToken);

  if (!payload) {
    throw new Error("Invalid or expired refresh token");
  }

  const user = await prisma.user.findUnique({
    where: { id: payload.userId },
  });

  if (!user) {
    throw new Error("User not found");
  }

  const tokens = generateTokens(user.id);

  return {
    accessToken: tokens.accessToken,
    refreshToken: tokens.refreshToken,
  };
};
