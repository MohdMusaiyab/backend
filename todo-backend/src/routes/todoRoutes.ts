import { Router } from "express";
import {
  createTodoController,
  updateTodoController,
  getAllTodosController,
  getTodoByIdController,
  deleteTodoController,
} from "../controllers/todoController.js";
import { authMiddleware } from "../middlewares/authMiddleware.js";

const todoRoutes = Router();

todoRoutes.use(authMiddleware);

todoRoutes.get("/", getAllTodosController);
todoRoutes.get("/:id", getTodoByIdController);
todoRoutes.post("/create", createTodoController);
todoRoutes.patch("/:id", updateTodoController);
todoRoutes.delete("/:id", deleteTodoController);

export default todoRoutes;