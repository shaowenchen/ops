import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useEventsStore = defineStore({
  id: "events",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(subject, page_size = 10, page = 1) {
      const analysis = await fetchWrapper.get(
        `/api/v1/events/${subject}?page_size=${page_size}&page=${page}&max_len=1000`
      );
      return analysis.data;
    },
  },
});
