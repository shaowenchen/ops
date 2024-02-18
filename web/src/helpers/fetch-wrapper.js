import { useLoginStore } from "@/stores";

import { router } from "@/router";

export const fetchWrapper = {
  get: request("GET"),
  post: request("POST"),
  put: request("PUT"),
  delete: request("DELETE"),
};

function request(method) {
  return (url, body) => {
    const requestOptions = {
      method,
      headers: {},
      body: null,
    };
    if (body) {
      requestOptions.headers["Content-Type"] = "application/json";
      requestOptions.body = JSON.stringify(body);
    }
    const loginStore = useLoginStore();
    const token = loginStore.get();
    if (token) {
      requestOptions.headers["Authorization"] = `Bearer ${token}`;
    }
    return fetch(url, requestOptions).then(handleResponse);
  };
}

async function handleResponse(response) {
  const isJson = response.headers
    ?.get("content-type")
    ?.includes("application/json");
  const data = isJson ? await response.json() : null;

  // check for error response
  if (!response.ok) {
    if ([401, 403].includes(response.status)) {
      const loginStore = useLoginStore();
      loginStore.clear();
      router.push({ name: "login" });
    }

    // get error message from body or default to response status
    const error = (data && data.message) || response.status;
    return Promise.reject(error);
  }

  return data;
}
