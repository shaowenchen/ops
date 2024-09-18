import { defineStore } from "pinia";
import { fetchWrapper } from "@/helpers";

export const useCopilotStore = defineStore({
  id: "copilot",
  state: () => ({
    alert: null,
  }),
  actions: {
    async post(input) {
      const res = await fetchWrapper.post(`/api/v1/copilot`, {
        input: input,
      });
      return res.data;
    },
  },
});
