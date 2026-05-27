import express from "express";
import { env } from "./lib/env.js";
import authRoutes from "./routes/authRoutes.js";
import cors from "cors";
import cookieParser from "cookie-parser";
import userRoutes from "./routes/userRoutes.js";

const app = express();

app.use(express.json());
app.use(express.urlencoded({ extended: false }));
app.use(cors());
app.use(cookieParser());

app.use("/auth", authRoutes);
app.use("/users", userRoutes);

app.listen(env.PORT, () => {
  console.log(`Server is running on port ${env.PORT}`);
});
