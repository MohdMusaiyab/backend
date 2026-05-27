import { prisma } from "../lib/prisma.js";
import {
  hashPassword,
  verifyPassword,
  generateTokens,
  verifyRefreshToken,
} from "../utils/auth.js";
import type {
  LoginInput,
  RegisterInput,
  ResetPasswordInput,
} from "../schemas/auth.schema.js";

export const createUserService = async (data: RegisterInput) => {
  const existingUser = await prisma.user.findUnique({
    where: { email: data.email },
  });

  if (existingUser) {
    throw new Error("User with this email already exists");
  }

  const hashedPassword = await hashPassword(data.password);

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

  const isPasswordValid = await verifyPassword(
    data.password,
    existingUser.password,
  );

  if (!isPasswordValid) {
    throw new Error("Invalid email or password");
  }

  const { accessToken, refreshToken } = generateTokens(existingUser.id);

  return {
    user: { id: existingUser.id, email: existingUser.email },
    accessToken,
    refreshToken,
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

export const forgotPasswordService = async (email: string) => {
  const user = await prisma.user.findUnique({
    where: { email },
  });
  if (!user) {
    throw new Error("Email not found");
  }

  const oneMinuteAgo = new Date(Date.now() - 60 * 1000);

  const recentToken = await prisma.verificationToken.findFirst({
    where: {
      userId: user.id,
      purpose: "FORGOT_PASSWORD",
      expiresAt: {
        gt: new Date(Date.now() + 14 * 60 * 1000), // expires in > 14 minutes
      },
    },
  });

  if (recentToken) {
    throw new Error("Please wait 1 minute before requesting another email");
  }

  await prisma.verificationToken.deleteMany({
    where: {
      userId: user.id,
      purpose: "FORGOT_PASSWORD",
    },
  });

  const otp = Math.floor(100000 + Math.random() * 900000).toString();

  const expiresAt = new Date(Date.now() + 15 * 60 * 1000);

  await prisma.verificationToken.create({
    data: {
      token: otp,
      expiresAt,
      purpose: "FORGOT_PASSWORD",
      userId: user.id,
    },
  });

  console.log(`\n============================`);
  console.log(`✉️ EMAIL SENT TO: ${user.email}`);
  console.log(`🔑 OTP: ${otp}`);
  console.log(`============================\n`);
};

export const resetPasswordService = async (data: ResetPasswordInput) => {
  const user = await prisma.user.findUnique({
    where: { email: data.email },
  });

  if (!user) {
    throw new Error("Invalid email or OTP");
  }

  const verificationRecord = await prisma.verificationToken.findFirst({
    where: {
      userId: user.id,
      token: data.otp,
      purpose: "FORGOT_PASSWORD",
    },
  });

  if (!verificationRecord) {
    throw new Error("Invalid or expired OTP");
  }

  if (verificationRecord.expiresAt < new Date()) {
    throw new Error("OTP has expired. Please request a new one.");
  }

  const hashedPassword = await hashPassword(data.password);

  await prisma.user.update({
    where: { id: user.id },
    data: { password: hashedPassword },
  });

  await prisma.verificationToken.delete({
    where: { id: verificationRecord.id },
  });

  const { accessToken, refreshToken } = generateTokens(user.id);

  return {
    user: { id: user.id, email: user.email },
    accessToken,
    refreshToken,
  };
};
