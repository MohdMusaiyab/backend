import { Worker } from 'bullmq';
import sharp from 'sharp';
import path from 'path';
import prisma from '../prisma.js';
import { connection } from './config.js';

const imageWorker = new Worker('image-jobs', async (job) => {
  const { jobId, filePath } = job.data;
  
  console.log(`\n🔨 Picked up Job ${jobId}...`);

  try {
    await prisma.job.update({
      where: { id: jobId },
      data: { status: 'processing' }
    });

    await new Promise(resolve => setTimeout(resolve, 3000));

    const outputFilename = `processed-${jobId}.jpg`;
    const outputPath = path.join(process.cwd(), 'processed', outputFilename);
    
    await sharp(filePath)
      .resize(800)
      .grayscale()
      .blur(4)
      .toFile(outputPath);

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
    
    await prisma.job.update({
      where: { id: jobId },
      data: { status: 'failed' }
    });
    
    throw error;
  }
}, { connection: connection as any });

imageWorker.on('failed', (job, err) => {
  console.error(`Job ${job?.id} failed with error:`, err.message);
});

console.log('👷 Background Worker is running and listening for jobs...');
