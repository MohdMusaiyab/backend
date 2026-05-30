import type {
  CreateTodoInput,
  UpdateTodoInput,
} from "../schemas/todo.schema.js";
import { prisma } from "../lib/prisma.js";

export const createToDoService = async (
  data: CreateTodoInput,
  userId: string,
) => {
  return await prisma.todo.create({
    data: {
      title: data.title,
      endDate: new Date(data.endDate),
      userId,
    },
  });
};

export const updateToDoService = async (
  todoId: string,
  data: UpdateTodoInput,
  userId: string,
) => {
  const todo = await prisma.todo.findUnique({
    where: { id: todoId },
  });

  if (!todo) {
    throw new Error("Todo not found");
  }

  if (todo.userId !== userId) {
    throw new Error("Unauthorized to update this todo");
  }

  const updateData: any = {};
  if (data.title !== undefined) updateData.title = data.title;
  if (data.isCompleted !== undefined) updateData.isCompleted = data.isCompleted;
  if (data.endDate !== undefined) updateData.endDate = new Date(data.endDate);

  return await prisma.todo.update({
    where: { id: todoId },
    data: updateData,
  });
};

export const getAllTodosService = async (userId: string) => {
  return await prisma.todo.findMany({
    where: { userId },
    orderBy: { createdAt: "desc" },
  });
};

export const getTodoByIdService = async (todoId: string, userId: string) => {
  const todo = await prisma.todo.findUnique({
    where: { id: todoId },
  });

  if (!todo) {
    throw new Error("Todo not found");
  }

  if (todo.userId !== userId) {
    throw new Error("Unauthorized to access this todo");
  }

  return todo;
};
