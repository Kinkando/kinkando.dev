import { MongoClient } from 'mongodb';
import { connect } from 'mongoose';
import config from '$config/server';

const mongoURL = `mongodb+srv://${config.mongo.username}:${config.mongo.password}@${config.mongo.host}/${config.mongo.database}?ssl=true&retryWrites=true`;
const client = new MongoClient(mongoURL);

export async function NewMongoClient() {
	console.log(`mongo: connecting to ${config.mongo.host}`);
	await connect(mongoURL, { dbName: config.mongo.database });
	client.on('connectionCreated', () => console.log(`mongo: connected to ${config.mongo.host}`));
	return await client.connect();
}

export const mongoDB = client.db(config.mongo.database);
