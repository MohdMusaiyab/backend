import express from "express";
import path from "path";
import imageRoutes from "./routes/image.routes.js";

const app = express();

app.use(express.json());

// Serve the static frontend UI (index.html, style.css, main.js)
app.use(express.static(path.join(process.cwd(), "public")));

// Serve the image folders so the frontend can display the results
app.use("/uploads", express.static(path.join(process.cwd(), "uploads")));
app.use("/processed", express.static(path.join(process.cwd(), "processed")));

// Mount the image processing routes
app.use("/api/images", imageRoutes);

// Health check
app.get("/health", (req, res) => {
  res.status(200).json({ status: "ok", message: "API is running" });
});

export default app;
