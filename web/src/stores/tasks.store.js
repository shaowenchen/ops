import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useTasksStore = defineStore({
  id: "tasks",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, search = "") {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks?page_size=${page_size}&page=${page}&search=${search}`
      );
      return analysis.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/tasks/${name}`
      );
      return analysis.data;
    },
    async update(namespace, name, task) {
      const analysis = await fetchWrapper.put(
        `/api/v1/namespaces/${namespace}/tasks/${name}`,
        task
      );
      return analysis.data;
    },
    async create(namespace, task) {
      const analysis = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/tasks`,
        task
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
