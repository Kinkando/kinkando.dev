import { env } from 'next-runtime-env';

import type { AppConfig, FirebaseConfig } from './';

export interface Config {
  readonly app: AppConfig;
  readonly firebase: FirebaseConfig;
}

const config: Config = {
  app: {
    apiHost: env('NEXT_PUBLIC_API_URL')!
  },
  firebase: {
    apiKey: env('NEXT_PUBLIC_FIREBASE_API_KEY')!,
    authDomain: env('NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN')!,
    projectId: env('NEXT_PUBLIC_FIREBASE_PROJECT_ID')!,
    storageBucket: env('NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET')!,
    messagingSenderId: env('NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID')!,
    appId: env('NEXT_PUBLIC_FIREBASE_APP_ID')!
  }
};

export default config;
