import type { Request, Response, NextFunction } from "express";
import jwt from "jsonwebtoken";
import { env } from "../lib/env.js";

export interface AuthRequest extends Request {
  user?: {
    userId: string;
  };
}

export const authMiddleware = (
  req: AuthRequest,
  res: Response,
  next: NextFunction
): void => {
  try {
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      res.status(401).json({
        success: false,
        message: "Access token missing or invalid format",
      });
      return;
    }

    const token = authHeader.split(" ")[1];
    
    if (!token) {
      res.status(401).json({
        success: false,
        message: "Access token missing",
      });
      return;
    }

    const decoded = jwt.verify(token, env.JWT_ACCESS_SECRET) as unknown as { userId: string };
    
    req.user = decoded;
    
    next();
  } catch (error) {
    res.status(401).json({
      success: false,
      message: "Invalid or expired access token",
    });
  }
};
