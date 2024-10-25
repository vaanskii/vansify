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
      const access = localStorage.getItem("user.access");
      const refresh = localStorage.getItem("user.refresh");
      if (access && refresh) {
        this.user.access = access;
        this.user.refresh = refresh;
        this.user.id = localStorage.getItem("id");
        this.user.username = localStorage.getItem("username");
        this.user.email = localStorage.getItem("email");
        this.user.isAuthenticated = true;
        axios.defaults.headers.common["Authorization"] = "Bearer " + this.user.access;
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
        const response = await axios.post("http://localhost:8080/v1/refresh-token", {
          refresh_token: this.user.refresh,
        });
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
      const timeout = expires.getTime() - Date.now() - 60 * 1000;
      console.log("Setting refresh timer for:", timeout / 1000, "seconds");
      this.refreshTokenTimeout = setTimeout(this.refreshToken, timeout);
    },
    stopRefreshTokenTimer() {
      clearTimeout(this.refreshTokenTimeout);
    },
  },
});
