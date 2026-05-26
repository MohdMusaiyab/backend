import type { Request, Response } from "express";
import { createUserService } from "../services/authService.js";
import { registerSchema } from "../schemas/auth.schema.js";
import { ZodError } from "zod";

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

export const loginController = async (req: Request, res: Response) => {
  // Similar pattern: Validate -> Call Login Service -> Return Response
};
