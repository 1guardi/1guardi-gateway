import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";

// Your web app's Firebase configuration
const firebaseConfig = {
  apiKey: "AIzaSyAkKBBQg00PVBR8-RPiQatD-UQaMCYoJP4",
  authDomain: "aigateway-001.firebaseapp.com",
  projectId: "aigateway-001",
  storageBucket: "aigateway-001.firebasestorage.app",
  messagingSenderId: "991147385265",
  appId: "1:991147385265:web:4e37a30d46a33cdfa59d90",
  measurementId: "G-55J3RH6R96"
};

// Initialize Firebase
export const app = initializeApp(firebaseConfig);
export const analytics = typeof window !== "undefined" ? getAnalytics(app) : null;
