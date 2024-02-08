import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useClustersStore = defineStore({
  id: "clusters",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/clusters`
      );
      return res.data.list;
    },
  },
});
