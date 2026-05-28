import type { Response } from "express";
import { createTodoSchema, updateTodoSchema, getTodoSchema } from "../schemas/todo.schema.js";
import { createToDoService, updateToDoService, getAllTodosService, getTodoByIdService } from "../services/todoServices.js";
import { ZodError } from "zod";
import type { AuthRequest } from "../middlewares/authMiddleware.js";

export const createTodoController = async (
  req: AuthRequest,
  res: Response
): Promise<void> => {
  try {
    const validatedData = createTodoSchema.parse({ body: req.body });
    const userId = req.user!.userId;

    const todo = await createToDoService(validatedData.body, userId);

    res.status(201).json({
      success: true,
      message: "Todo created successfully",
      data: todo,
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
    res.status(500).json({ success: false, message: "An unexpected error occurred" });
  }
};

export const updateTodoController = async (
  req: AuthRequest,
  res: Response
): Promise<void> => {
  try {
    const validatedData = updateTodoSchema.parse({
      params: req.params,
      body: req.body,
    });

    const userId = req.user!.userId;

    const todo = await updateToDoService(
      validatedData.params.id,
      validatedData.body,
      userId
    );

    res.status(200).json({
      success: true,
      message: "Todo updated successfully",
      data: todo,
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
      res.status(error.message.includes("Unauthorized") ? 403 : 404).json({
        success: false,
        message: error.message,
      });
      return;
    }
    res.status(500).json({ success: false, message: "An unexpected error occurred" });
  }
};

export const getAllTodosController = async (
  req: AuthRequest,
  res: Response
): Promise<void> => {
  try {
    const userId = req.user!.userId;
    const todos = await getAllTodosService(userId);

    res.status(200).json({
      success: true,
      data: todos,
    });
  } catch (error: unknown) {
    res.status(500).json({ success: false, message: "An unexpected error occurred" });
  }
};

export const getTodoByIdController = async (
  req: AuthRequest,
  res: Response
): Promise<void> => {
  try {
    const validatedData = getTodoSchema.parse({ params: req.params });
    const userId = req.user!.userId;

    const todo = await getTodoByIdService(validatedData.params.id, userId);

    res.status(200).json({
      success: true,
      data: todo,
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
      res.status(error.message.includes("Unauthorized") ? 403 : 404).json({
        success: false,
        message: error.message,
      });
      return;
    }
    res.status(500).json({ success: false, message: "An unexpected error occurred" });
  }
};