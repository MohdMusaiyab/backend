import expres from "express";
import {
  loginController,
  registerController,
} from "../controllers/authController.js";

const authRoutes = expres.Router();

authRoutes.post("/register", registerController);
authRoutes.post("/login", loginController);

export default authRoutes;
