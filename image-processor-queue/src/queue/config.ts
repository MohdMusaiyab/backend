import { Redis, type RedisOptions } from 'ioredis';
import 'dotenv/config';


const redisOptions: RedisOptions = {
  maxRetriesPerRequest: null,
};

export const connection = new Redis(process.env.REDIS_URL || 'redis://localhost:6379', redisOptions);

connection.on('error', (err) => {
  console.error('❌ Redis Connection Error:', err);
});

connection.on('connect', () => {
  console.log('✅ Connected to Redis successfully');
});
