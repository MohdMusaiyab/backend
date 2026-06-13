import express from "express";
import path from "path";
import imageRoutes from "./routes/image.routes.js";

const app = express();

app.use(express.json());

app.use(express.static(path.join(process.cwd(), "public")));

app.use("/uploads", express.static(path.join(process.cwd(), "uploads")));
app.use("/processed", express.static(path.join(process.cwd(), "processed")));

app.use("/api/images", imageRoutes);

app.get("/health", (req, res) => {
  res.status(200).json({ status: "ok", message: "API is running" });
});

export default app;
