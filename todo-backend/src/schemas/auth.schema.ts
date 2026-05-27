import { z } from "zod";

export const registerSchema = z.object({
  body: z.object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(8, "Password must be at least 8 characters"),
  }),
});

export const loginSchema = z.object({
  body: z.object({
    email: z.string().email(),
    password: z.string(),
  }),
});

export type RegisterInput = z.infer<typeof registerSchema>["body"];
export type LoginInput = z.infer<typeof loginSchema>["body"];

export const forgotPasswordSchema = z.object({
  body: z.object({
    email: z.string().email(),
  }),
});
export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>["body"];

export const resetPasswordSchema = z.object({
  body: z.object({
    email: z.string().email(),
    password: z.string().min(8, "Password must be at least 8 characters"),
    otp: z.string().length(6, "OTP must be exactly 6 characters"),
  }),
});
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>["body"];
