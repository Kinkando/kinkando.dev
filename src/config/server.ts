import { env } from '$env/dynamic/private';

export interface Config {
  readonly version: string;
  readonly mongo: {
    readonly host: string;
    readonly database: string;
    readonly username: string;
    readonly password: string;
  };
  readonly redis: {
    readonly host: string;
    readonly port: number;
    readonly username: string;
    readonly password: string;
  };
}

const config: Config = {
  version: env.VERSION,
  mongo: {
    host: env.MONGO_HOST,
    database: env.MONGO_DATABASE,
    username: env.MONGO_USERNAME,
    password: env.MONGO_PASSWORD
  },
  redis: {
    host: env.REDIS_HOST,
    port: Number(env.REDIS_PORT),
    username: env.REDIS_USERNAME,
    password: env.REDIS_PASSWORD
  }
};

export default config;
