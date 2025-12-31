import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";

const STORAGE_KEY = "ops-selected-namespace";
const DEFAULT_NAMESPACE = "ops-system";

export const useNamespacesStore = defineStore({
  id: "namespaces",
  state: () => ({
    namespaces: [],
    selectedNamespace: localStorage.getItem(STORAGE_KEY) || DEFAULT_NAMESPACE,
  }),
  actions: {
    async list() {
      const res = await fetchWrapper.get("/api/v1/namespaces");
      this.namespaces = res?.data || res || [];
      // If current selected namespace is not in the list, use the first one or default
      if (this.namespaces.length > 0 && !this.namespaces.includes(this.selectedNamespace)) {
        this.setSelectedNamespace(this.namespaces[0]);
      }
      return this.namespaces;
    },
    setSelectedNamespace(namespace) {
      this.selectedNamespace = namespace;
      localStorage.setItem(STORAGE_KEY, namespace);
    },
  },
});

