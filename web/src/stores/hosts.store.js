import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useHostsStore = defineStore({
  id: "hosts",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/hosts`
      );
      return res.data.list;
    }
  },
});
