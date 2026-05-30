import type { Request, Response, NextFunction } from "express";
import { ZodError } from "zod";
import { AppError } from "../lib/AppError.js";

export const errorHandler = (
  err: any,
  req: Request,
  res: Response,
  next: NextFunction
): void => {
  if (err instanceof ZodError) {
    res.status(400).json({
      success: false,
      message: "Validation failed",
      errors: err.issues.map((e) => ({
        field: e.path.join("."),
        message: e.message,
      })),
    });
    return;
  }

  if (err instanceof AppError) {
    res.status(err.statusCode).json({
      success: false,
      message: err.message,
    });
    return;
  }

  if (err instanceof Error) {
    let statusCode = 400;
    if (err.message.toLowerCase().includes("unauthorized") || err.message.toLowerCase().includes("invalid")) statusCode = 401;
    if (err.message.toLowerCase().includes("not found")) statusCode = 404;

    res.status(statusCode).json({
      success: false,
      message: err.message,
    });
    return;
  }

  console.error("Unhandled Error:", err);
  res.status(500).json({
    success: false,
    message: "An unexpected error occurred",
  });
};
