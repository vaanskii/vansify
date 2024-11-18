import { defineStore } from "pinia";
import axios from "axios";
import CryptoJS from "crypto-js";

export const userStore = defineStore({
  id: "user",
  state: () => ({
    user: {
      isAuthenticated: false,
      id: null,
      username: null,
      email: null,
      oauth_user: null,
      access: null,
      refresh: null,
    },
    refreshTokenTimeout: null,
  }),
  actions: {
    initStore() {
      const encryptedAccess = sessionStorage.getItem("sess_a");
      const encryptedRefresh = sessionStorage.getItem("sess_r");
      const userData = localStorage.getItem("user_data");
      const secretKey = "your-secret-key";

      if (encryptedAccess && encryptedRefresh && userData) {
        try {
          const access = CryptoJS.AES.decrypt(encryptedAccess, secretKey).toString(CryptoJS.enc.Utf8);
          const refresh = CryptoJS.AES.decrypt(encryptedRefresh, secretKey).toString(CryptoJS.enc.Utf8);

          const parsedUserData = JSON.parse(userData);

          const jwtToken = JSON.parse(atob(access.split(".")[1]));
          const expires = new Date(jwtToken.exp * 1000);

          if (expires < Date.now()) {
            this.refreshToken();
          } else {
            this.user.access = access;
            this.user.refresh = refresh;
            this.user.id = parsedUserData.id;
            this.user.username = parsedUserData.username;
            this.user.email = parsedUserData.email;
            this.user.oauth_user = parsedUserData.oauth_user;
            this.user.isAuthenticated = true;

            axios.defaults.headers.common["Authorization"] = "Bearer " + this.user.access;
            this.startRefreshTokenTimer();
          }
        } catch (error) {
          console.error("Error decrypting tokens:", error);
          this.removeToken(); // Clear invalid or corrupted tokens
        }
      }
    },
    setToken(data) {
      const secretKey = "your-secret-key"; // Use a secret key for encryption/decryption
      const encryptedAccess = CryptoJS.AES.encrypt(data.access, secretKey).toString();
      const encryptedRefresh = CryptoJS.AES.encrypt(data.refresh, secretKey).toString();

      // Combine user data into one object
      const userData = {
        id: data.id,
        username: data.username,
        email: data.email,
        oauth_user: data.oauth_user,
      };

      this.user.access = data.access;
      this.user.refresh = data.refresh;
      this.user.id = data.id;
      this.user.username = data.username;
      this.user.email = data.email;
      this.user.oauth_user = data.oauth_user;
      this.user.isAuthenticated = true;

      // Save encrypted tokens in sessionStorage
      sessionStorage.setItem("sess_a", encryptedAccess);
      sessionStorage.setItem("sess_r", encryptedRefresh);

      // Generate fake tokens for obfuscation
      const fakeAccess = this.generateFakeToken(10);
      const fakeRefresh = this.generateFakeToken(12);

      // Save combined user data and fake tokens in localStorage
      localStorage.setItem("user_data", JSON.stringify(userData));
      localStorage.setItem("auth_a", fakeAccess);
      localStorage.setItem("auth_r", fakeRefresh);

      axios.defaults.headers.common["Authorization"] = "Bearer " + this.user.access;
      this.startRefreshTokenTimer();
    },
    removeToken() {
      this.user.refresh = null;
      this.user.access = null;
      this.user.isAuthenticated = false;
      this.user.id = null;
      this.user.username = null;
      this.user.email = null;
      this.user.oauth_user = null;

      localStorage.clear();
      sessionStorage.clear();
      this.stopRefreshTokenTimer();
    },
    async refreshToken() {
      try {
        const response = await axios.post("/v1/refresh-token", {
          refresh_token: this.user.refresh,
        });
        const secretKey = "your-secret-key";
        const encryptedAccess = CryptoJS.AES.encrypt(response.data.access_token, secretKey).toString();

        // Update tokens
        this.user.access = response.data.access_token;
        sessionStorage.setItem("sess_a", encryptedAccess);
        axios.defaults.headers.common["Authorization"] = "Bearer " + this.user.access;
        this.startRefreshTokenTimer();
      } catch (error) {
        console.error("Refresh token failed:", error);
        this.removeToken();
        router.push("/login");
      }
    },
    startRefreshTokenTimer() {
      const jwtToken = JSON.parse(atob(this.user.access.split(".")[1]));
      const expires = new Date(jwtToken.exp * 1000);
      const timeout = expires.getTime() - Date.now() - 60 * 1000;
      this.refreshTokenTimeout = setTimeout(() => this.refreshToken(), timeout);
    },
    stopRefreshTokenTimer() {
      clearTimeout(this.refreshTokenTimeout);
    },
    generateFakeToken(length) {
      const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
      let result = "";
      for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * characters.length));
      }
      return result;
    },
  },
});