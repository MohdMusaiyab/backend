import { type Request,type  Response } from "express";
import prisma from "../prisma.js";
import { imageQueue } from "../queue/config.js";

export const uploadImage = async (req: Request, res: Response): Promise<void> => {
  try {
    if (!req.file) {
      res.status(400).json({ error: "No image file uploaded" });
      return;
    }

    const jobRecord = await prisma.job.create({
      data: {
        imageUrl: req.file.path, 
        status: "pending",
      },
    });

    await imageQueue.add("process-image", {
      jobId: jobRecord.id,
      filePath: req.file.path,
    });

    res.status(202).json({
      message: "Image uploaded and queued for processing",
      jobId: jobRecord.id,
    });
  } catch (error) {
    console.error("Error in uploadImage:", error);
    res.status(500).json({ error: "Internal server error" });
  }
};

export const getJobStatus = async (req: Request, res: Response): Promise<void> => {
  try {
    const id = req.params.id as string;

    const jobRecord = await prisma.job.findUnique({
      where: { id },
    });

    if (!jobRecord) {
      res.status(404).json({ error: "Job not found" });
      return;
    }

    res.status(200).json({
      jobId: jobRecord.id,
      status: jobRecord.status,
      imageUrl: jobRecord.imageUrl,
    });
  } catch (error) {
    console.error("Error in getJobStatus:", error);
    res.status(500).json({ error: "Internal server error" });
  }
};
