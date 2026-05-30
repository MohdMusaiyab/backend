import { Router } from "express";
import { authMiddleware } from "../middlewares/authMiddleware.js";
import {
  getProfileController,
  updateProfileController,
  deleteProfileController,
} from "../controllers/userController.js";

const userRoutes = Router();

// Protect all user routes
userRoutes.use(authMiddleware);

userRoutes.get("/profile", getProfileController);
userRoutes.patch("/profile", updateProfileController);
userRoutes.delete("/profile", deleteProfileController);

export default userRoutes;
