import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";

const firebaseConfig = {
  apiKey: "AIzaSyAkKBBQg00PVBR8-RPiQatD-UQaMCYoJP4",
  authDomain: "aigateway-001.firebaseapp.com",
  projectId: "aigateway-001",
  storageBucket: "aigateway-001.firebasestorage.app",
  messagingSenderId: "991147385265",
  appId: "1:991147385265:web:021db10614eed5a7a59d90",
  measurementId: "G-9GSZPMN486",
};

export const app = initializeApp(firebaseConfig);
export const analytics = getAnalytics(app);
