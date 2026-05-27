import expres from "express";
import {
  loginController,
  registerController,
  refreshController,
  logoutController
} from "../controllers/authController.js";

const authRoutes = expres.Router();

authRoutes.post("/register", registerController);
authRoutes.post("/login", loginController);
authRoutes.post("/refresh", refreshController);
authRoutes.post("/logout", logoutController);

export default authRoutes;
