import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

export const useEventHooksStore = defineStore({
  id: "eventhooks",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1, search = "") {
      let url = `/api/v1/namespaces/${namespace}/eventhooks?page_size=${page_size}&page=${page}`;
      if (search) {
        url += `&search=${encodeURIComponent(search)}`;
      }
      const analysis = await fetchWrapper.get(url);
      return analysis.data;
    },
    async get(namespace, name) {
      const analysis = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/eventhooks/${name}`
      );
      return analysis.data;
    },
    async create(namespace, eventhook) {
      const analysis = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/eventhooks`,
        eventhook
      );
      return analysis.data;
    },
    async update(namespace, name, eventhook) {
      const analysis = await fetchWrapper.put(
        `/api/v1/namespaces/${namespace}/eventhooks/${name}`,
        eventhook
      );
      return analysis.data;
    },
    async delete(namespace, name) {
      const analysis = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/eventhooks/${name}`
      );
      return analysis.data;
    },
  },
});

