import { defineStore } from "pinia";

export const useLoginStore = defineStore({
  id: "login",
  state: () => ({
    token: null,
  }),
  actions: {
    save(token) {
      this.token = token;
    },
    get() {
      return this.token;
    },
    clear() {
      this.token = null;
    }
  },
});
