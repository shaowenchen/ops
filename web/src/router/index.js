import { createRouter, createWebHistory } from "vue-router";

import { Home, Hosts, Clusters, Tasks, TaskRuns, Pipelines, PipelineRuns, Login, Copilot} from "@/views";

export const router = createRouter({
  history: createWebHistory(),
  linkActiveClass: "active",
  routes: [
    { path: "/", component: Home, name: "home" },
    { path: "/hosts", component: Hosts, name: "hosts" },
    { path: "/clusters", component: Clusters, name: "clusters" },
    { path: "/tasks", component: Tasks, name: "tasks" },
    { path: "/taskruns", component: TaskRuns, name: "taskruns" },
    { path: "/pipelines", component: Pipelines, name: "pipelines" },
    { path: "/pipelineruns", component: PipelineRuns, name: "pipelineruns" },
    { path: "/login", component: Login, name: "login" },
    { path: "/copilot", component: Copilot, name: "copilot" },
  ],
});

router.beforeEach(async (to) => {});
