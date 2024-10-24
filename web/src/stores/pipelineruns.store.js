import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";
import { useAlertStore } from "./alert.store";

export const usePipelineRunsStore = defineStore({
  id: "pipelineruns",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace, page_size = 10, page = 1) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/pipelineruns?page_size=${page_size}&page=${page}`
      );
      return res.data;
    },
    async get(namespace, name) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/pipelineruns/${name}`
      );
      return res.data;
    },
    async delete(namespace, name) {
      const res = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/pipelineruns/${name}`
      );
      return res.data;
    },
    async create(namespace, ref, vars) {
      const res = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/pipelineruns`,
        {
          pipelineRef: ref,
          variables: vars,
        }
      );
      if (res.code == 0) {
        useAlertStore().success(res.message);
      } else {
        useAlertStore().error(res.message);
      }
      return res;
    },
  },
});
