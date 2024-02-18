import { createRouter, createWebHistory } from "vue-router";

import { Home, Hosts, Clusters, Tasks, TaskRuns, Login} from "@/views";

export const router = createRouter({
  history: createWebHistory(),
  linkActiveClass: "active",
  routes: [
    { path: "/", component: Home, name: "home" },
    { path: "/hosts", component: Hosts, name: "hosts" },
    { path: "/clusters", component: Clusters, name: "clusters" },
    { path: "/tasks", component: Tasks, name: "tasks" },
    { path: "/taskruns", component: TaskRuns, name: "taskruns" },
    { path: "/login", component: Login, name: "login" },
  ],
});

router.beforeEach(async (to) => {});
