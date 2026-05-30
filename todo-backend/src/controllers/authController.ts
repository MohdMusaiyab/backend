import type { Request, Response } from "express";
import {
  createUserService,
  loginUserService,
  refreshUserService,
  forgotPasswordService,
  resetPasswordService,
} from "../services/authService.js";
import {
  registerSchema,
  loginSchema,
  forgotPasswordSchema,
  resetPasswordSchema,
} from "../schemas/auth.schema.js";
import { env } from "../lib/env.js";
import { catchAsync } from "../lib/catchAsync.js";

export const registerController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  const validatedData = registerSchema.parse({ body: req.body });

  const user = await createUserService(validatedData.body);

  res.status(201).json({
    success: true,
    message: "User registered successfully",
    data: user,
  });
});

export const loginController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  const validatedData = loginSchema.parse({ body: req.body });

  const { user, accessToken, refreshToken } = await loginUserService(
    validatedData.body,
  );

  res.cookie("refreshToken", refreshToken, {
    httpOnly: true,
    secure: env.NODE_ENV === "production",
    sameSite: "strict",
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
  });

  res.status(200).json({
    success: true,
    message: "Login successful",
    data: {
      user,
      accessToken,
    },
  });
});

export const refreshController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  const { refreshToken } = req.cookies;

  if (!refreshToken) {
    res.status(401).json({
      success: false,
      message: "Refresh token is missing",
    });
    return;
  }

  const tokens = await refreshUserService(refreshToken);

  res.cookie("refreshToken", tokens.refreshToken, {
    httpOnly: true,
    secure: env.NODE_ENV === "production",
    sameSite: "strict",
    maxAge: 7 * 24 * 60 * 60 * 1000,
  });

  res.status(200).json({
    success: true,
    message: "Token refreshed successfully",
    data: {
      accessToken: tokens.accessToken,
    },
  });
});

export const logoutController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  res.clearCookie("refreshToken", {
    httpOnly: true,
    secure: env.NODE_ENV === "production",
    sameSite: "strict",
  });

  res.status(200).json({
    success: true,
    message: "Logged out successfully",
  });
});

export const forgotPasswordController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  const validatedData = forgotPasswordSchema.parse({ body: req.body });

  await forgotPasswordService(validatedData.body.email);

  res.status(200).json({
    success: true,
    message:
      "If an account with this email exists, a password reset OTP has been sent.",
  });
});

export const resetPasswordController = catchAsync(async (
  req: Request,
  res: Response,
) => {
  const validatedData = resetPasswordSchema.parse({ body: req.body });

  const { user, accessToken, refreshToken } = await resetPasswordService(
    validatedData.body,
  );

  res.cookie("refreshToken", refreshToken, {
    httpOnly: true,
    secure: env.NODE_ENV === "production",
    sameSite: "strict",
    maxAge: 7 * 24 * 60 * 60 * 1000,
  });

  res.status(200).json({
    success: true,
    message: "Password reset successful",
    data: {
      user,
      accessToken,
    },
  });
});
