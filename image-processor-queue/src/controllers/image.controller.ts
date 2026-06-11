import { type Request,type  Response } from "express";
import prisma from "../prisma.js";
import { imageQueue } from "../queue/config.js";

export const uploadImage = async (req: Request, res: Response): Promise<void> => {
  try {
    if (!req.file) {
      res.status(400).json({ error: "No image file uploaded" });
      return;
    }

    // 1. Create the Database Record
    // We register the job as 'pending' and store where Multer saved the file.
    const jobRecord = await prisma.job.create({
      data: {
        imageUrl: req.file.path, 
        status: "pending",
      },
    });

    // 2. Add to BullMQ
    // This pushes the data into Redis. We only pass the ID and the file path.
    await imageQueue.add("process-image", {
      jobId: jobRecord.id,
      filePath: req.file.path,
    });

    // 3. Respond Immediately
    // We do NOT wait for processing. We tell the user "Accepted!" instantly.
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

    // The frontend polls this endpoint. We just check the DB and return the status.
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
      imageUrl: jobRecord.imageUrl, // This will be the final image if it's completed
    });
  } catch (error) {
    console.error("Error in getJobStatus:", error);
    res.status(500).json({ error: "Internal server error" });
  }
};
