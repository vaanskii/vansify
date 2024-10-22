import { defineStore } from "pinia";
import axios from "axios";

export const userStore = defineStore({
  id: "user",
  state: () => ({
    user: {
      isAuthenticated: false,
      id: null,
      username: null,
      email: null,
      profile_picture: null,
      gender: null,
      access: null,
      refresh: null,
    },
  }),
  actions: {
    initStore() {
      console.log("Initializing store...");
      const access = localStorage.getItem("user.access");
      const refresh = localStorage.getItem("user.refresh");
      if (access && refresh) {
        this.user.access = access;
        this.user.refresh = refresh;
        this.user.id = localStorage.getItem("id");
        this.user.username = localStorage.getItem("username");
        this.user.email = localStorage.getItem("email");
        this.user.profile_picture = localStorage.getItem("profile_picture");
        this.user.gender = localStorage.getItem("gender");
        this.user.isAuthenticated = true;
        axios.defaults.headers.common["Authorization"] = "Bearer " + this.user.access;
        console.log("Access and refresh tokens found. Starting refresh token timer.");
        this.startRefreshTokenTimer();
      }
    },
    setToken(data) {
      this.user.access = data.access;
      this.user.refresh = data.refresh;
      this.user.id = data.id;
      this.user.username = data.username;
      this.user.email = data.email;
      this.user.profile_picture = data.profile_picture;
      this.user.gender = data.gender;
      this.user.isAuthenticated = true;
      localStorage.setItem("user.access", data.access);
      localStorage.setItem("user.refresh", data.refresh);
      localStorage.setItem("id", data.id);
      localStorage.setItem("username", data.username);
      localStorage.setItem("email", data.email);
      localStorage.setItem("profile_picture", data.profile_picture);
      localStorage.setItem("gender", data.gender);
      this.startRefreshTokenTimer();
    },
    removeToken() {
      this.user.refresh = null;
      this.user.access = null;
      this.user.isAuthenticated = false;
      this.user.id = null;
      this.user.username = null;
      this.user.email = null;
      this.user.profile_picture = null;
      this.user.gender = null;
      localStorage.clear();
      this.stopRefreshTokenTimer();
    },
    async refreshToken() {
      try {
        console.log("Refreshing token...");
        const response = await axios.post("http://localhost:8080/v1/refresh-token", {
          refresh_token: this.user.refresh,
        });
        console.log("New Access Token:", response.data.access_token);
        this.user.access = response.data.access_token;
        localStorage.setItem("user.access", response.data.access_token);
        axios.defaults.headers.common["Authorization"] = "Bearer " + response.data.access_token;
        this.startRefreshTokenTimer();
      } catch (error) {
        console.log("Refresh token failed:", error);
        this.removeToken();
      }
    },
    startRefreshTokenTimer() {
      const jwtToken = JSON.parse(atob(this.user.access.split(".")[1]));
      const expires = new Date(jwtToken.exp * 1000);
      const timeout = expires.getTime() - Date.now() - 60 * 1000; // Adjust for test
      console.log("Setting refresh timer for:", timeout / 1000, "seconds");
      this.refreshTokenTimeout = setTimeout(this.refreshToken, timeout);
    },
    stopRefreshTokenTimer() {
      clearTimeout(this.refreshTokenTimeout);
    },
  },
});
