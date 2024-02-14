import { defineStore } from "pinia";

import { fetchWrapper } from "@/helpers";
import { useAlertStore } from "./alert.store";

export const useTaskRunsStore = defineStore({
  id: "taskruns",
  state: () => ({
    alert: null,
  }),
  actions: {
    async list(namespace) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/taskruns`
      );
      return res.data.list;
    },
    async get(namespace, name) {
      const res = await fetchWrapper.get(
        `/api/v1/namespaces/${namespace}/taskruns/${name}`
      );
      return res.data.list;
    },
    async delete(namespace, name) {
      const res = await fetchWrapper.delete(
        `/api/v1/namespaces/${namespace}/taskruns/${name}`
      );
      return res.data.list;
    },
    async create(namespace, taskRef, typeRef, nameRef) {
      const res = await fetchWrapper.post(
        `/api/v1/namespaces/${namespace}/taskruns`,
        {
          taskRef: taskRef,
          typeRef: typeRef,
          nameRef: nameRef,
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
