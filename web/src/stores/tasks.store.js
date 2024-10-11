import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useTasksStore = defineStore({
  id: "tasks",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, searchQuery = "") {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks?page_size=${page_size}&page=${page}&search=${searchQuery}`
      );
      return analysis.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks/${name}`
      );
      return analysis.data;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/tasks/${name}`
      );
      return analysis.data;
    },
  },
});
