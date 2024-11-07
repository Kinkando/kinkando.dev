import { PUBLIC_VERSION } from '$env/static/public';

export type Config = {
    readonly version: string;
}

const config: Config = {
    version: PUBLIC_VERSION,
}

export default config;
