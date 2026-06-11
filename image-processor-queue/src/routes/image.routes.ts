import { Router } from "express";
import multer from "multer";
import path from "path";
import { uploadImage, getJobStatus } from "../controllers/image.controller.js";

const router = Router();

// Configure Multer
// This tells Multer to intercept incoming files and save them physically to the "uploads/" folder
const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, "uploads/");
  },
  filename: (req, file, cb) => {
    // We add a random suffix to prevent filename collisions if multiple users upload "image.png"
    const uniqueSuffix = Date.now() + "-" + Math.round(Math.random() * 1e9);
    cb(null, uniqueSuffix + path.extname(file.originalname));
  },
});

const upload = multer({ storage });

// Bind the routes to our controller functions
router.post("/upload", upload.single("image"), uploadImage);
router.get("/status/:id", getJobStatus);

export default router;
