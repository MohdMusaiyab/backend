import type { Response } from "express";
import type { AuthRequest } from "../middlewares/authMiddleware.js";
import { ZodError } from "zod";
import { updateUserSchema } from "../schemas/user.schema.js";
import {
  getUserProfileService,
  updateUserService,
  deleteUserService,
} from "../services/userService.js";

export const getProfileController = async (
  req: AuthRequest,
  res: Response,
): Promise<void> => {
  try {
    const userId = req.user!.userId;
    const userProfile = await getUserProfileService(userId);

    res.status(200).json({
      success: true,
      data: userProfile,
    });
  } catch (error: unknown) {
    if (error instanceof Error) {
      res.status(404).json({ success: false, message: error.message });
      return;
    }
    res
      .status(500)
      .json({ success: false, message: "An unexpected error occurred" });
  }
};

export const updateProfileController = async (
  req: AuthRequest,
  res: Response,
): Promise<void> => {
  try {
    const validatedData = updateUserSchema.parse({ body: req.body });
    const userId = req.user!.userId;

    const updatedUser = await updateUserService(userId, validatedData.body);

    res.status(200).json({
      success: true,
      message: "Profile updated successfully",
      data: updatedUser,
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
      res.status(400).json({ success: false, message: error.message });
      return;
    }
    res
      .status(500)
      .json({ success: false, message: "An unexpected error occurred" });
  }
};

export const deleteProfileController = async (
  req: AuthRequest,
  res: Response,
): Promise<void> => {
  try {
    const userId = req.user!.userId;

    await deleteUserService(userId);

    res.clearCookie("refreshToken");

    res.status(200).json({
      success: true,
      message: "User account and all associated data deleted successfully",
    });
  } catch (error: unknown) {
    if (error instanceof Error) {
      res.status(400).json({ success: false, message: error.message });
      return;
    }
    res
      .status(500)
      .json({ success: false, message: "An unexpected error occurred" });
  }
};
