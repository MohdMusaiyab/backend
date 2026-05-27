import type { Request, Response } from "express";
import { createUserService, loginUserService, refreshUserService } from "../services/authService.js";
import { registerSchema, loginSchema } from "../schemas/auth.schema.js";
import { ZodError } from "zod";
import { env } from "../lib/env.js";

export const registerController = async (
  req: Request,
  res: Response,
): Promise<void> => {
  try {
    // 1. Validate incoming request body
    const validatedData = registerSchema.parse({ body: req.body });

    // 2. Pass validated data to the Service layer
    const user = await createUserService(validatedData.body);

    // 3. Send response back to client
    res.status(201).json({
      success: true,
      message: "User registered successfully",
      data: user,
    });
  } catch (error: unknown) {
    // Handle Zod validation errors specifically to give detailed feedback
    if (error instanceof ZodError) {
      res.status(400).json({
        success: false,
        message: "Validation failed",
        errors: error.issues.map((err) => ({
          field: err.path.join("."),
          message: err.message,
        })),
      });
      return;
    }

    // Handle generic Errors (like our "User already exists" throw from the service)
    if (error instanceof Error) {
      res.status(400).json({
        success: false,
        message: error.message,
      });
      return;
    }

    // Fallback for completely unknown errors
    res.status(500).json({
      success: false,
      message: "An unexpected error occurred",
    });
  }
};

export const loginController = async (req: Request, res: Response): Promise<void> => {
  try {
    const validatedData = loginSchema.parse({ body: req.body });

    const { user, accessToken, refreshToken } = await loginUserService(validatedData.body);

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
  } catch (error: unknown) {
    if (error instanceof ZodError) {
      res.status(400).json({
        success: false,
        message: "Validation failed",
        errors: error.issues.map((err) => ({
          field: err.path.join("."),
          message: err.message,
        })),
      });
      return;
    }
    
    if (error instanceof Error) {
      res.status(401).json({
        success: false,
        message: error.message,
      });
      return;
    }

    res.status(500).json({
      success: false,
      message: "An unexpected error occurred",
    });
  }
};

export const refreshController = async (req: Request, res: Response): Promise<void> => {
  try {
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
  } catch (error: unknown) {
    if (error instanceof Error) {
      res.status(401).json({
        success: false,
        message: error.message,
      });
      return;
    }

    res.status(500).json({
      success: false,
      message: "An unexpected error occurred",
    });
  }
};

export const logoutController = async (req: Request, res: Response): Promise<void> => {
  res.clearCookie("refreshToken", {
    httpOnly: true,
    secure: env.NODE_ENV === "production",
    sameSite: "strict",
  });
  
  res.status(200).json({
    success: true,
    message: "Logged out successfully",
  });
};
