import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const usePipelinesStore = defineStore({
  id: "pipelines",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, search = "") {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/pipelines?search=${search}&page_size=${page_size}&page=${page}`
      );
      return analysis.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/pipelines/${name}`
      );
      return analysis.data;
    },
    async update(namespace, name, pipeline) {
      const analysis = await fetchWrapper.put(
        `/api/v1/namespaces/${namespace}/pipelines/${name}`,
        pipeline
      );
      return analysis.data;
    },
    async create(namespace, pipeline) {
      const analysis = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/pipelines`,
        pipeline
      );
      return analysis.data;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/pipelines/${name}`
      );
      return analysis.data;
    },
  },
});
