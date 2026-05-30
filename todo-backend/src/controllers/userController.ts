import type { Response } from "express";
import type { AuthRequest } from "../middlewares/authMiddleware.js";
import { updateUserSchema } from "../schemas/user.schema.js";
import {
  getUserProfileService,
  updateUserService,
  deleteUserService,
} from "../services/userService.js";
import { catchAsync } from "../lib/catchAsync.js";

export const getProfileController = catchAsync(async (
  req: AuthRequest,
  res: Response,
) => {
  const userId = req.user!.userId;
  const userProfile = await getUserProfileService(userId);

  res.status(200).json({
    success: true,
    data: userProfile,
  });
});

export const updateProfileController = catchAsync(async (
  req: AuthRequest,
  res: Response,
) => {
  const validatedData = updateUserSchema.parse({ body: req.body });
  const userId = req.user!.userId;

  const updatedUser = await updateUserService(userId, validatedData.body);

  res.status(200).json({
    success: true,
    message: "Profile updated successfully",
    data: updatedUser,
  });
});

export const deleteProfileController = catchAsync(async (
  req: AuthRequest,
  res: Response,
) => {
  const userId = req.user!.userId;

  await deleteUserService(userId);

  res.clearCookie("refreshToken");

  res.status(200).json({
    success: true,
    message: "User account and all associated data deleted successfully",
  });
});
