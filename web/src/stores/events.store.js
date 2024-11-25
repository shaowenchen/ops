import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useEventsStore = defineStore({
  id: "events",
  state: () => ({
    alert: null,
  }),
  actions: {
    async get(event, page_size = 10, page = 1) {
      const analysis = await fetchWrapper.get(
        `/api/v1/events/${event}?page_size=${page_size}&page=${page}`
      );
      return analysis.data;
    },
    async list(search, page_size = 10, page = 1) {
      const analysis = await fetchWrapper.get(
        `/api/v1/events?page_size=${page_size}&page=${page}&search=${search}`
      );
      return analysis.data;
    },
  },
});
