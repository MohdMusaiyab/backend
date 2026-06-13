import { Router } from "express";
import multer from "multer";
import path from "path";
import { uploadImage, getJobStatus } from "../controllers/image.controller.js";

const router = Router();

const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, "uploads/");
  },
  filename: (req, file, cb) => {
    const uniqueSuffix = Date.now() + "-" + Math.round(Math.random() * 1e9);
    cb(null, uniqueSuffix + path.extname(file.originalname));
  },
});

const upload = multer({ storage });

router.post("/upload", upload.single("image"), uploadImage);
router.get("/status/:id", getJobStatus);

export default router;
