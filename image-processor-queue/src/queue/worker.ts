import { Worker } from 'bullmq';
import sharp from 'sharp';
import path from 'path';
import prisma from '../prisma.js';
import { connection } from './config.js';

// BullMQ Worker Initialization
// We listen to the exact same 'image-jobs' queue we created in config.ts
const imageWorker = new Worker('image-jobs', async (job) => {
  const { jobId, filePath } = job.data;
  
  console.log(`\n🔨 Picked up Job ${jobId}...`);

  try {
    // 1. Mark as Processing
    await prisma.job.update({
      where: { id: jobId },
      data: { status: 'processing' }
    });

    // We add a 3-second delay here just so you can actually see the 
    // "Processing" text in the UI before it finishes instantly!
    await new Promise(resolve => setTimeout(resolve, 3000));

    const outputFilename = `processed-${jobId}.jpg`;
    const outputPath = path.join(process.cwd(), 'processed', outputFilename);
    
    // 2. Perform heavy image processing with Sharp
    // Here we resize it, turn it black and white, and add a blur filter.
    await sharp(filePath)
      .resize(800)
      .grayscale()
      .blur(4)
      .toFile(outputPath);

    // 3. Mark as Completed and save the new URL
    await prisma.job.update({
      where: { id: jobId },
      data: { 
        status: 'completed',
        imageUrl: `/processed/${outputFilename}`
      }
    });

    console.log(`✅ Job ${jobId} completed and saved successfully!`);

  } catch (error) {
    console.error(`❌ Job ${jobId} failed:`, error);
    
    // If anything fails, we update the DB so the frontend knows
    await prisma.job.update({
      where: { id: jobId },
      data: { status: 'failed' }
    });
    
    throw error; // Let BullMQ know this job failed so it can be retried if configured
  }
}, { connection: connection as any });

imageWorker.on('failed', (job, err) => {
  console.error(`Job ${job?.id} failed with error:`, err.message);
});

console.log('👷 Background Worker is running and listening for jobs...');
