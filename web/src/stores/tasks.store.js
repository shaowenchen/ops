import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useTasksStore = defineStore({
  id: "tasks",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks`
      );
      return analysis.data.list;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks/${name}`
      );
      return analysis.data.list;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/tasks/${name}`
      );
      return analysis.data.list;
    },
  },
});
