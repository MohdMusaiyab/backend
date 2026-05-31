import {
  OpenAPIRegistry,
  OpenApiGeneratorV3,
} from "@asteasolutions/zod-to-openapi";
import { z } from "zod";
import { extendZodWithOpenApi } from "@asteasolutions/zod-to-openapi";

extendZodWithOpenApi(z);

export const registry = new OpenAPIRegistry();

const bearerAuth = registry.registerComponent("securitySchemes", "bearerAuth", {
  type: "http",
  scheme: "bearer",
  bearerFormat: "JWT",
});

import {
  registerSchema,
  loginSchema,
  forgotPasswordSchema,
  resetPasswordSchema,
} from "../schemas/auth.schema.js";
import {
  createTodoSchema,
  updateTodoSchema,
  getTodoSchema,
  getAllTodosQuerySchema,
} from "../schemas/todo.schema.js";
import { updateUserSchema } from "../schemas/user.schema.js";

const successResponse = (dataSchema?: z.ZodTypeAny) => {
  const shape: any = {
    success: z.boolean().openapi({ example: true }),
    message: z.string().optional().openapi({ example: "Operation successful" }),
  };
  if (dataSchema) shape.data = dataSchema;
  return z.object(shape);
};

const errorResponse = z.object({
  success: z.boolean().openapi({ example: false }),
  message: z.string().openapi({ example: "Error message" }),
});

// --- AUTHENTICATION ROUTES ---
registry.registerPath({
  method: "post",
  path: "/auth/register",
  summary: "Register a new user",
  tags: ["Auth"],
  request: {
    body: {
      content: { "application/json": { schema: registerSchema.shape.body } },
    },
  },
  responses: {
    201: {
      description: "User registered",
      content: {
        "application/json": {
          schema: successResponse(
            z.object({ id: z.string(), email: z.string() }),
          ),
        },
      },
    },
    400: {
      description: "Validation error",
      content: { "application/json": { schema: errorResponse } },
    },
  },
});

registry.registerPath({
  method: "post",
  path: "/auth/login",
  summary: "Login user",
  tags: ["Auth"],
  request: {
    body: {
      content: { "application/json": { schema: loginSchema.shape.body } },
    },
  },
  responses: {
    200: {
      description: "User logged in",
      content: {
        "application/json": {
          schema: successResponse(
            z.object({
              user: z.object({ id: z.string(), email: z.string() }),
              accessToken: z.string(),
            }),
          ),
        },
      },
    },
    401: {
      description: "Invalid credentials",
      content: { "application/json": { schema: errorResponse } },
    },
  },
});

registry.registerPath({
  method: "post",
  path: "/auth/refresh",
  summary: "Refresh access token",
  tags: ["Auth"],
  description:
    "Uses the HttpOnly refreshToken cookie to issue a new accessToken.",
  responses: {
    200: {
      description: "Token refreshed",
      content: {
        "application/json": {
          schema: successResponse(z.object({ accessToken: z.string() })),
        },
      },
    },
    401: {
      description: "Unauthorized",
      content: { "application/json": { schema: errorResponse } },
    },
  },
});

registry.registerPath({
  method: "post",
  path: "/auth/logout",
  summary: "Logout user",
  tags: ["Auth"],
  description: "Clears the HttpOnly refreshToken cookie.",
  responses: {
    200: {
      description: "Logged out",
      content: { "application/json": { schema: successResponse() } },
    },
  },
});

registry.registerPath({
  method: "post",
  path: "/auth/forgot-password",
  summary: "Request password reset OTP",
  tags: ["Auth"],
  request: {
    body: {
      content: {
        "application/json": { schema: forgotPasswordSchema.shape.body },
      },
    },
  },
  responses: {
    200: {
      description: "OTP Sent",
      content: { "application/json": { schema: successResponse() } },
    },
  },
});

registry.registerPath({
  method: "post",
  path: "/auth/reset-password",
  summary: "Reset password with OTP",
  tags: ["Auth"],
  request: {
    body: {
      content: {
        "application/json": { schema: resetPasswordSchema.shape.body },
      },
    },
  },
  responses: {
    200: {
      description: "Password reset successful",
      content: {
        "application/json": {
          schema: successResponse(
            z.object({
              user: z.object({ id: z.string(), email: z.string() }),
              accessToken: z.string(),
            }),
          ),
        },
      },
    },
    400: {
      description: "Invalid or expired OTP",
      content: { "application/json": { schema: errorResponse } },
    },
  },
});

// --- TODO ROUTES ---
const todoSchema = z.object({
  id: z.string(),
  title: z.string(),
  isCompleted: z.boolean(),
  endDate: z.string().datetime(),
  userId: z.string(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

registry.registerPath({
  method: "post",
  path: "/todos/create",
  summary: "Create a new todo",
  tags: ["Todos"],
  security: [{ [bearerAuth.name]: [] }],
  request: {
    body: {
      content: { "application/json": { schema: createTodoSchema.shape.body } },
    },
  },
  responses: {
    201: {
      description: "Todo created",
      content: { "application/json": { schema: successResponse(todoSchema) } },
    },
  },
});

registry.registerPath({
  method: "get",
  path: "/todos",
  summary: "Get all todos",
  tags: ["Todos"],
  security: [{ [bearerAuth.name]: [] }],
  request: { query: getAllTodosQuerySchema.shape.query },
  responses: {
    200: {
      description: "List of todos",
      content: {
        "application/json": {
          schema: z.object({
            success: z.boolean().openapi({ example: true }),
            data: z.array(todoSchema),
            meta: z.object({
              total: z.number(),
              page: z.number(),
              limit: z.number(),
              totalPages: z.number(),
            }),
          }),
        },
      },
    },
  },
});

registry.registerPath({
  method: "get",
  path: "/todos/{id}",
  summary: "Get a specific todo",
  tags: ["Todos"],
  security: [{ [bearerAuth.name]: [] }],
  request: { params: getTodoSchema.shape.params },
  responses: {
    200: {
      description: "Todo details",
      content: { "application/json": { schema: successResponse(todoSchema) } },
    },
    404: {
      description: "Not found",
      content: { "application/json": { schema: errorResponse } },
    },
  },
});

registry.registerPath({
  method: "patch",
  path: "/todos/{id}",
  summary: "Update a specific todo",
  tags: ["Todos"],
  security: [{ [bearerAuth.name]: [] }],
  request: {
    params: updateTodoSchema.shape.params,
    body: {
      content: { "application/json": { schema: updateTodoSchema.shape.body } },
    },
  },
  responses: {
    200: {
      description: "Todo updated",
      content: { "application/json": { schema: successResponse(todoSchema) } },
    },
  },
});

registry.registerPath({
  method: "delete",
  path: "/todos/{id}",
  summary: "Delete a specific todo",
  tags: ["Todos"],
  security: [{ [bearerAuth.name]: [] }],
  request: { params: getTodoSchema.shape.params },
  responses: {
    200: {
      description: "Todo deleted",
      content: { "application/json": { schema: successResponse() } },
    },
  },
});

// --- USER ROUTES ---
registry.registerPath({
  method: "get",
  path: "/users/profile",
  summary: "Get user profile",
  tags: ["Users"],
  security: [{ [bearerAuth.name]: [] }],
  responses: {
    200: {
      description: "User profile",
      content: {
        "application/json": {
          schema: successResponse(
            z.object({
              id: z.string(),
              email: z.string(),
              todo: z.array(todoSchema),
            }),
          ),
        },
      },
    },
  },
});

registry.registerPath({
  method: "patch",
  path: "/users/profile",
  summary: "Update user profile",
  tags: ["Users"],
  security: [{ [bearerAuth.name]: [] }],
  request: {
    body: {
      content: { "application/json": { schema: updateUserSchema.shape.body } },
    },
  },
  responses: {
    200: {
      description: "User profile updated",
      content: {
        "application/json": {
          schema: successResponse(
            z.object({ id: z.string(), email: z.string() }),
          ),
        },
      },
    },
  },
});

registry.registerPath({
  method: "delete",
  path: "/users/profile",
  summary: "Delete user profile",
  tags: ["Users"],
  security: [{ [bearerAuth.name]: [] }],
  responses: {
    200: {
      description: "Profile deleted",
      content: { "application/json": { schema: successResponse() } },
    },
  },
});

export const openApiDocument = new OpenApiGeneratorV3(
  registry.definitions,
).generateDocument({
  openapi: "3.0.0",
  info: {
    version: "1.0.0",
    title: "Todo API",
    description: "Production Ready Todo API",
  },
});
