import { createClient } from 'redis';

import config from '$config/server';

const redis = createClient({
  password: config.redis.password,
  socket: {
    host: config.redis.host,
    port: config.redis.port
  }
});

export async function NewRedisClient() {
  console.log(`redis: connecting to ${config.redis.host}:${config.redis.port}`);
  redis.on('connect', () => console.log(`redis: connected to ${config.redis.host}:${config.redis.port}`));
  return await redis.connect();
}

export { redis };
