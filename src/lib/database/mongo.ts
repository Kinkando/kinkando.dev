import { MongoClient } from 'mongodb';
import { connect } from "mongoose";
import config from '$config/server';

export async function NewMongoClient() {
    const mongoURL = `mongodb+srv://${config.mongo.username}:${config.mongo.password}@${config.mongo.host}/${config.mongo.database}?ssl=true&retryWrites=true`;
    console.log(`mongo: connecting to ${config.mongo.host}`);
    await connect(mongoURL, { dbName: config.mongo.database });
    const client = new MongoClient(mongoURL);
    client.on('connectionCreated', () => console.log(`mongo: connected to ${config.mongo.host}`));
    await client.connect();
    return client.db(config.mongo.database);
}
