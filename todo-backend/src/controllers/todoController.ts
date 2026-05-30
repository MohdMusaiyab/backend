import type { Response } from "express";
import { createTodoSchema, updateTodoSchema, getTodoSchema, getAllTodosQuerySchema } from "../schemas/todo.schema.js";
import { createToDoService, updateToDoService, getAllTodosService, getTodoByIdService, deleteTodoService } from "../services/todoServices.js";
import type { AuthRequest } from "../middlewares/authMiddleware.js";
import { catchAsync } from "../lib/catchAsync.js";

export const createTodoController = catchAsync(async (
  req: AuthRequest,
  res: Response
) => {
  const validatedData = createTodoSchema.parse({ body: req.body });
  const userId = req.user!.userId;

  const todo = await createToDoService(validatedData.body, userId);

  res.status(201).json({
    success: true,
    message: "Todo created successfully",
    data: todo,
  });
});

export const updateTodoController = catchAsync(async (
  req: AuthRequest,
  res: Response
) => {
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
});

export const getAllTodosController = catchAsync(async (
  req: AuthRequest,
  res: Response
) => {
  const validatedData = getAllTodosQuerySchema.parse({ query: req.query });
  const userId = req.user!.userId;
  const result = await getAllTodosService(userId, validatedData.query);

  res.status(200).json({
    success: true,
    data: result.todos,
    meta: result.meta,
  });
});

export const getTodoByIdController = catchAsync(async (
  req: AuthRequest,
  res: Response
) => {
  const validatedData = getTodoSchema.parse({ params: req.params });
  const userId = req.user!.userId;

  const todo = await getTodoByIdService(validatedData.params.id, userId);

  res.status(200).json({
    success: true,
    data: todo,
  });
});

export const deleteTodoController = catchAsync(async (
  req: AuthRequest,
  res: Response
) => {
  const validatedData = getTodoSchema.parse({ params: req.params });
  const userId = req.user!.userId;

  await deleteTodoService(validatedData.params.id, userId);

  res.status(200).json({
    success: true,
    message: "Todo deleted successfully",
  });
});