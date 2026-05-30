import type {
  CreateTodoInput,
  UpdateTodoInput,
  GetAllTodosQueryInput,
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

export const getAllTodosService = async (userId: string, query: GetAllTodosQueryInput) => {
  const { page, limit, search, isCompleted } = query;
  
  const where: any = { userId };
  
  if (isCompleted !== undefined) {
    where.isCompleted = isCompleted;
  }
  
  if (search) {
    where.title = { contains: search, mode: "insensitive" };
  }

  const skip = (page - 1) * limit;

  const [todos, total] = await Promise.all([
    prisma.todo.findMany({
      where,
      orderBy: { createdAt: "desc" },
      skip,
      take: limit,
    }),
    prisma.todo.count({ where }),
  ]);

  return {
    todos,
    meta: {
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    },
  };
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

export const deleteTodoService = async (todoId: string, userId: string) => {
  const todo = await prisma.todo.findUnique({
    where: { id: todoId },
  });

  if (!todo) {
    throw new Error("Todo not found");
  }

  if (todo.userId !== userId) {
    throw new Error("Unauthorized to delete this todo");
  }

  await prisma.todo.delete({
    where: { id: todoId },
  });
};
