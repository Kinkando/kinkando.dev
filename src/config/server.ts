import { env } from '$env/dynamic/private';

export type Config = {
    readonly version: string;
    readonly mongo: {
        readonly host: string;
        readonly database: string;
        readonly username: string;
        readonly password: string;
    }
}

const config: Config = {
    version: env.VERSION,
    mongo: {
        host: env.MONGO_HOST,
        database: env.MONGO_DATABASE,
        username: env.MONGO_USERNAME,
        password: env.MONGO_PASSWORD,
    }
}

export default config;
