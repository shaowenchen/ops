import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useHostsStore = defineStore({
  id: "hosts",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/hosts?page_size=${page_size}&page=${page}`
      );
      return res.data;
    }
  },
});
