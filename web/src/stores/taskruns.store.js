import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useTaskRunsStore = defineStore({
  id: "taskruns",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/taskruns`
      );
      return analysis.data.list;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/taskruns/${name}`
      );
      return analysis.data.list;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/taskruns/${name}`
      );
      return analysis.data.list;
    },
  },
});
