import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useHostsStore = defineStore({
  id: "hosts",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, searchQuery = "") {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/hosts?page_size=${page_size}&page=${page}&search=${searchQuery}`
      );
      return res.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/hosts/${name}`
      );
      return analysis.data;
    },
    async create(namespace, host) {
      const analysis = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/hosts`,
        host
      );
      return analysis.data;
    },
    async update(namespace, name, host) {
      const analysis = await fetchWrapper.put(
        `/api/v1/namespaces/${namespace}/hosts/${name}`,
        host
      );
      return analysis.data;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/hosts/${name}`
      );
      return analysis.data;
    },
  },
});
