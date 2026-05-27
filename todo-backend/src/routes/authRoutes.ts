import expres from "express";
import {
  loginController,
  registerController,
  refreshController,
  logoutController,
  forgotPasswordController,
  resetPasswordController,
} from "../controllers/authController.js";

const authRoutes = expres.Router();

authRoutes.post("/register", registerController);
authRoutes.post("/login", loginController);
authRoutes.post("/refresh", refreshController);
authRoutes.post("/logout", logoutController);
authRoutes.post("/forgot-password", forgotPasswordController);
authRoutes.post("/reset-password", resetPasswordController);

export default authRoutes;
