import { z } from "zod";

export const createTodoSchema = z.object({
  body: z.object({
    title: z.string().min(1, "Title is required").max(100),
    endDate: z.string().datetime("Must be a valid ISO Date string"),
  }),
});

export const updateTodoSchema = z.object({
  params: z.object({
    id: z.string().uuid("Invalid Todo ID"),
  }),
  body: z.object({
    title: z.string().max(100).optional(),
    isCompleted: z.boolean().optional(),
    endDate: z.string().datetime().optional(),
  }),
});

export const getTodoSchema = z.object({
  params: z.object({
    id: z.string().uuid("Invalid Todo ID"),
  }),
});

export type CreateTodoInput = z.infer<typeof createTodoSchema>["body"];
export type UpdateTodoInput = z.infer<typeof updateTodoSchema>["body"];
