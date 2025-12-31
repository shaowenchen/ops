import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useClustersStore = defineStore({
  id: "clusters",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, searchQuery = "") {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/clusters?page_size=${page_size}&page=${page}&search=${searchQuery}`
      );
      return res.data;
    },
    async listNodes(
      namespace,
      cluster,
      page_size = 10,
      page = 1,
      searchQuery = ""
    ) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/clusters/${cluster}/nodes?page_size=${page_size}&page=${page}&search=${searchQuery}`
      );
      return res.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/clusters/${name}`
      );
      return analysis.data;
    },
    async create(namespace, cluster) {
      const analysis = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/clusters`,
        cluster
      );
      return analysis.data;
    },
    async update(namespace, name, cluster) {
      const analysis = await fetchWrapper.put(
        `/api/v1/namespaces/${namespace}/clusters/${name}`,
        cluster
      );
      return analysis.data;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/clusters/${name}`
      );
      return analysis.data;
    },
  },
});
