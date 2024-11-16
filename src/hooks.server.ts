import { NewMongoClient } from '$lib/database/mongo';
import { NewRedisClient } from '$lib/database/redis';

NewMongoClient();
NewRedisClient();
