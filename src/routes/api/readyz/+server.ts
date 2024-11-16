import { json, type RequestHandler } from '@sveltejs/kit';
import { mongoDB } from '$lib/database/mongo';
import { redis } from '$lib/database/redis';

export const GET: RequestHandler = async () => {
	try {
		await mongoDB.command({ ping: 1 });
		await redis.ping();
		return json({
			mongo: 'OK',
			redis: 'OK'
		});
	} catch (error) {
		return json({ error }, { status: 500 });
	}
};
