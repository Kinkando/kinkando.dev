import config from '$config/server';
import { createClient } from 'redis';

export default async function NewRedisClient() {
	console.log(`redis: connecting to ${config.redis.host}:${config.redis.port}`);
	const client = createClient({
		password: config.redis.password,
		socket: {
			host: config.redis.host,
			port: config.redis.port
		}
	});
	client.on('connect', () =>
		console.log(`redis: connected to ${config.redis.host}:${config.redis.port}`)
	);
	await client.connect();
	return client;
}
