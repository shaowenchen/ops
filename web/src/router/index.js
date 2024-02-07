import { createRouter, createWebHistory } from "vue-router";

import { Home, Hosts, Clusters, Tasks, TaskRuns} from "@/views";

export const router = createRouter({
  history: createWebHistory(),
  linkActiveClass: "active",
  routes: [
    { path: "/", component: Home },
    { path: "/hosts", component: Hosts },
    { path: "/clusters", component: Clusters },
    { path: "/tasks", component: Tasks },
    { path: "/taskruns", component: TaskRuns },
  ],
});

router.beforeEach(async (to) => {});
