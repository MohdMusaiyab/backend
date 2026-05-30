import { z } from "zod";

export const updateUserSchema = z.object({
  body: z.object({
    email: z.string().email("Invalid email format").optional(),
  }),
});

export type UpdateUserInput = z.infer<typeof updateUserSchema>["body"];
