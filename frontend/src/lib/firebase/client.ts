import { getApp, getApps, initializeApp } from 'firebase/app';
import { getAuth } from 'firebase/auth';
import clientConfig from '$config/client';

const app = getApps().length ? getApp() : initializeApp(clientConfig.firebaseConfig);
export const auth = getAuth(app);
