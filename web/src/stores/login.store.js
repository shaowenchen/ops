import { defineStore } from "pinia";
import Cookies from "js-cookie";
import { fetchWrapper } from "@/helpers";

export const useLoginStore = defineStore({
  id: "login",
  state: () => ({}),
  actions: {
    async save(token) {
      Cookies.set("opstoken", token, { expires: 7 });
      if (this.check()) {
        return true;
      } else {
        this.clear();
        return false;
      }
    },
    get() {
      return Cookies.get("opstoken");
    },
    clear() {
      Cookies.remove("opstoken");
    },
    async check() {
      const resp = await fetchWrapper.get(`/api/v1/login/check`);
      if (resp.status === 200) {
        return true;
      }
      return false;
    },
  },
});
