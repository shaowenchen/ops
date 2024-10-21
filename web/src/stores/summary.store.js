import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useSummaryStore = defineStore({
  id: "summary",
  state: () => ({
    alert: null,
  }),
  actions: {
    async get() {
      const res = await fetchWrapper.get(
        `/api/v1/summary`
      );
      return res.data;
    },
  },
});
